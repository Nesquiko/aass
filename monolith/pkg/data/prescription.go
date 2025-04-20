package data

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
)

type Prescription struct {
	Id            uuid.UUID  `bson:"_id"                     json:"id"`
	PatientId     uuid.UUID  `bson:"patientId"               json:"patientId"`               // Reference to Patient._id
	AppointmentId *uuid.UUID `bson:"appointmentId,omitempty" json:"appointmentId,omitempty"` // Reference to Appointment._id
	Name          string     `bson:"name"                    json:"name"`
	Start         time.Time  `bson:"start"                   json:"start"`
	End           time.Time  `bson:"end"                     json:"end"`
	DoctorsNote   *string    `bson:"doctorsNote,omitempty"   json:"doctorsNote,omitempty"`
}

func (m *MongoDb) CreatePrescription(
	ctx context.Context,
	prescription Prescription,
) (Prescription, error) {
	if err := m.patientExists(ctx, prescription.PatientId); err != nil {
		return Prescription{}, fmt.Errorf("CreatePrescription patient check error: %w", err)
	}

	collection := m.Database.Collection(prescriptionsCollection)
	prescription.Id = uuid.New()

	_, err := collection.InsertOne(ctx, prescription)
	if err != nil {
		return Prescription{}, fmt.Errorf("CreatePrescription failed to insert document: %w", err)
	}

	return prescription, nil
}

func (m *MongoDb) PrescriptionById(ctx context.Context, id uuid.UUID) (Prescription, error) {
	collection := m.Database.Collection(prescriptionsCollection)
	filter := bson.M{"_id": id}
	var prescription Prescription

	err := collection.FindOne(ctx, filter).Decode(&prescription)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Prescription{}, ErrNotFound
		}
		return Prescription{}, fmt.Errorf("PrescriptionById failed to find document: %w", err)
	}

	return prescription, nil
}

func (m *MongoDb) FindPrescriptionsByPatientId(
	ctx context.Context,
	patientId uuid.UUID,
	from time.Time,
	to *time.Time,
) ([]Prescription, error) {
	collection := m.Database.Collection(prescriptionsCollection)
	prescriptions := make([]Prescription, 0)

	filter := bson.M{
		"patientId": patientId,
		"end":       bson.M{"$gte": from},
	}
	if to != nil {
		filter["start"] = bson.M{"$lte": *to}
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "start", Value: 1}})
	cursor, err := collection.Find(ctx, filter, findOptions)
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

func (m *MongoDb) UpdatePrescription(
	ctx context.Context,
	id uuid.UUID,
	prescription Prescription,
) (Prescription, error) {
	if err := m.patientExists(ctx, prescription.PatientId); err != nil {
		return Prescription{}, fmt.Errorf(
			"UpdatePrescription patient check error: %w",
			err,
		)
	}

	collection := m.Database.Collection(prescriptionsCollection)
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
	err := collection.FindOneAndUpdate(ctx, filter, update, opts).
		Decode(&updatedPrescription)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Prescription{}, ErrNotFound
		}
		return Prescription{}, fmt.Errorf("UpdatePrescription failed: %w", err)
	}

	return updatedPrescription, nil
}

func (m *MongoDb) PrescriptionByAppointmentId(
	ctx context.Context,
	appointmentId uuid.UUID,
) ([]Prescription, error) {
	collection := m.Database.Collection(prescriptionsCollection)
	prescriptions := make([]Prescription, 0)

	filter := bson.M{"appointmentId": appointmentId}

	findOptions := options.Find().SetSort(bson.D{{Key: "start", Value: 1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
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

func (m *MongoDb) DeletePrescription(ctx context.Context, id uuid.UUID) error {
	collection := m.Database.Collection(prescriptionsCollection)
	filter := bson.M{"_id": id}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("DeletePrescription failed: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}
