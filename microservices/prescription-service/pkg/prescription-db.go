package pkg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/Nesquiko/aass/common/mongodb"
)

var ErrNotFound = errors.New("prescription not found")

type MongoPrescriptionDb struct {
	collection *mongo.Collection
}

const collection = "prescription"

func NewMongoPrescriptionDb(
	ctx context.Context,
	uri string,
	db string,
) (*MongoPrescriptionDb, error) {
	mongoDb, err := mongodb.ConnectMongo(ctx, uri, db)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to MongoDB", "uri", uri, "error", err)
		return nil, fmt.Errorf("NewMongoPrescriptionDb: %w", err)
	}

	coll := mongoDb.Collection(collection)
	indexModel := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "patientId", Value: 1}},
			Options: options.Index().SetName("idx_presription_patientId"),
		},
		{
			Keys:    bson.D{{Key: "patientId", Value: 1}, {Key: "start", Value: 1}},
			Options: options.Index().SetName("idx_presription_patientId_start"),
		},
	}
	for _, idx := range indexModel {
		_, err = coll.Indexes().CreateOne(ctx, idx)
		if err != nil {
			slog.WarnContext(ctx, "Could not ensure resources index", "error", err)
		}
	}

	return &MongoPrescriptionDb{collection: coll}, nil
}

type Prescription struct {
	Id            uuid.UUID  `bson:"_id"                     json:"id"`
	PatientId     uuid.UUID  `bson:"patientId"               json:"patientId"`               // Reference to Patient._id
	AppointmentId *uuid.UUID `bson:"appointmentId,omitempty" json:"appointmentId,omitempty"` // Reference to Appointment._id
	Name          string     `bson:"name"                    json:"name"`
	Start         time.Time  `bson:"start"                   json:"start"`
	End           time.Time  `bson:"end"                     json:"end"`
	DoctorsNote   *string    `bson:"doctorsNote,omitempty"   json:"doctorsNote,omitempty"`
}

func (m *MongoPrescriptionDb) CreatePrescription(
	ctx context.Context,
	prescription Prescription,
) (Prescription, error) {
	prescription.Id = uuid.New()

	_, err := m.collection.InsertOne(ctx, prescription)
	if err != nil {
		return Prescription{}, fmt.Errorf("CreatePrescription failed to insert document: %w", err)
	}

	return prescription, nil
}

func (m *MongoPrescriptionDb) PrescriptionById(
	ctx context.Context,
	id uuid.UUID,
) (Prescription, error) {
	filter := bson.M{"_id": id}
	var prescription Prescription

	err := m.collection.FindOne(ctx, filter).Decode(&prescription)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Prescription{}, ErrNotFound
		}
		return Prescription{}, fmt.Errorf("PrescriptionById failed to find document: %w", err)
	}

	return prescription, nil
}

func (m *MongoPrescriptionDb) FindPrescriptionsByPatientId(
	ctx context.Context,
	patientId uuid.UUID,
	from time.Time,
	to *time.Time,
) ([]Prescription, error) {
	prescriptions := make([]Prescription, 0)

	filter := bson.M{
		"patientId": patientId,
		"end":       bson.M{"$gte": from},
	}
	if to != nil {
		filter["start"] = bson.M{"$lte": *to}
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "start", Value: 1}})
	cursor, err := m.collection.Find(ctx, filter, findOptions)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return prescriptions, nil
		}
		return nil, fmt.Errorf("FindPrescriptionsByPatientId find failed: %w", err)
	}

	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			slog.Warn("Failed to close prescriptions cursor", "error", cerr.Error())
		}
	}()

	if err = cursor.All(ctx, &prescriptions); err != nil {
		slog.Error("Failed to decode documents from cursor", "error", err)
		return nil, fmt.Errorf("FindPrescriptionsByPatientId decode failed: %w", err)
	}

	if err = cursor.Err(); err != nil {
		slog.Error("Prescriptions cursor iteration error", "error", err)
		return nil, fmt.Errorf("FindPrescriptionsByPatientId cursor error: %w", err)
	}

	return prescriptions, nil
}

func (m *MongoPrescriptionDb) UpdatePrescription(
	ctx context.Context,
	id uuid.UUID,
	prescription Prescription,
) (Prescription, error) {
	filter := bson.M{"_id": id}

	updatePayload := bson.M{
		"patientId":     prescription.PatientId,
		"appointmentId": prescription.AppointmentId,
		"name":          prescription.Name,
		"start":         prescription.Start,
		"end":           prescription.End,
		"doctorsNote":   prescription.DoctorsNote,
	}
	update := bson.M{"$set": updatePayload}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedPrescription Prescription
	err := m.collection.FindOneAndUpdate(ctx, filter, update, opts).
		Decode(&updatedPrescription)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Prescription{}, ErrNotFound
		}
		return Prescription{}, fmt.Errorf("UpdatePrescription failed: %w", err)
	}

	return updatedPrescription, nil
}

func (m *MongoPrescriptionDb) PrescriptionByAppointmentId(
	ctx context.Context,
	appointmentId uuid.UUID,
) ([]Prescription, error) {
	prescriptions := make([]Prescription, 0)

	filter := bson.M{"appointmentId": appointmentId}

	findOptions := options.Find().SetSort(bson.D{{Key: "start", Value: 1}})

	cursor, err := m.collection.Find(ctx, filter, findOptions)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return prescriptions, nil
		}
		return nil, fmt.Errorf("PrescriptionByAppointmentId find failed: %w", err)
	}

	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			slog.Warn("Failed to close prescriptions by appointment cursor", "error", cerr.Error())
		}
	}()

	if err = cursor.All(ctx, &prescriptions); err != nil {
		slog.Error("Failed to decode prescription documents from cursor", "error", err)
		return nil, fmt.Errorf("PrescriptionByAppointmentId decode failed: %w", err)
	}

	if err = cursor.Err(); err != nil {
		slog.Error("Prescriptions by appointment cursor iteration error", "error", err)
		return nil, fmt.Errorf("PrescriptionByAppointmentId cursor error: %w", err)
	}

	return prescriptions, nil
}

func (m *MongoPrescriptionDb) DeletePrescription(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	result, err := m.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("DeletePrescription failed: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}
