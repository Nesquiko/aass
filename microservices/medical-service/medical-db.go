package main

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

const (
	conditionsCollection    = "conditions"
	prescriptionsCollection = "prescriptions"
)

var ErrNotFound = errors.New("resource not found")

type Condition struct {
	Id        uuid.UUID  `bson:"_id"           json:"id"`
	PatientId uuid.UUID  `bson:"patientId"     json:"patientId"`
	Name      string     `bson:"name"          json:"name"`
	Start     time.Time  `bson:"start"         json:"start"`
	End       *time.Time `bson:"end,omitempty" json:"end,omitempty"`
}

type mongoMedicalDb struct {
	conditions    *mongo.Collection
	prescriptions *mongo.Collection
}

func newMongoMedicalDb(ctx context.Context, uri string, db string) (mongoMedicalDb, error) {
	mongoDb, err := mongodb.ConnectMongo(ctx, uri, db)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to MongoDB", "uri", uri, "error", err)
		return mongoMedicalDb{}, fmt.Errorf("newMongoMedicalDb: %w", err)
	}

	conditionsColl := mongoDb.Collection(conditionsCollection)
	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "patientId", Value: 1}},
			Options: options.Index().SetName("idx_condition_patientId"),
		},
		{
			Keys:    bson.D{{Key: "patientId", Value: 1}, {Key: "start", Value: 1}},
			Options: options.Index().SetName("idx_condition_patientId_start"),
		},
	}

	for _, idx := range indexModels {
		_, err = conditionsColl.Indexes().CreateOne(ctx, idx)
		if err != nil {
			slog.WarnContext(
				ctx,
				"Could not ensure conditions index (may already exist)",
				"error",
				err,
			)
		}
	}

	prescriptionsColl := mongoDb.Collection(prescriptionsCollection)
	indexModels = []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "patientId", Value: 1}},
			Options: options.Index().SetName("idx_presription_patientId"),
		},
		{
			Keys:    bson.D{{Key: "patientId", Value: 1}, {Key: "start", Value: 1}},
			Options: options.Index().SetName("idx_presription_patientId_start"),
		},
	}

	for _, idx := range indexModels {
		_, err = prescriptionsColl.Indexes().CreateOne(ctx, idx)
		if err != nil {
			slog.WarnContext(
				ctx,
				"Could not ensure prescriptions index (may already exist)",
				"error",
				err,
			)
		}
	}

	return mongoMedicalDb{conditions: conditionsColl, prescriptions: prescriptionsColl}, nil
}

func (db mongoMedicalDb) Disconnect(ctx context.Context) error {
	return db.prescriptions.Database().Client().Disconnect(ctx)
}

func (m *mongoMedicalDb) CreateCondition(
	ctx context.Context,
	condition Condition,
) (Condition, error) {
	condition.Id = uuid.New()

	_, err := m.conditions.InsertOne(ctx, condition)
	if err != nil {
		return Condition{}, fmt.Errorf("CreateCondition: failed to insert document: %w", err)
	}

	return condition, nil
}

func (m *mongoMedicalDb) ConditionById(ctx context.Context, id uuid.UUID) (Condition, error) {
	filter := bson.M{"_id": id}
	var condition Condition

	err := m.conditions.FindOne(ctx, filter).Decode(&condition)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Condition{}, ErrNotFound
		}
		return Condition{}, fmt.Errorf("ConditionById: failed to find document: %w", err)
	}

	return condition, nil
}

func (m *mongoMedicalDb) FindConditionsByPatientId(
	ctx context.Context,
	patientId uuid.UUID,
	from time.Time,
	to *time.Time,
) ([]Condition, error) {
	conditions := make([]Condition, 0)

	filter := bson.M{"patientId": patientId}
	if to != nil {
		filter["start"] = bson.M{"$lte": *to}
	}
	filter["$or"] = []bson.M{{"end": bson.M{"$gte": from}}, {"end": nil}}

	opts := options.Find().SetSort(bson.D{{Key: "start", Value: 1}})
	cursor, err := m.conditions.Find(ctx, filter, opts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return conditions, nil
		}
		return nil, fmt.Errorf("FindConditionsByPatientId: find failed: %w", err)
	}

	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			slog.Warn("Failed to close cursor", "error", cerr.Error())
		}
	}()

	if err = cursor.All(ctx, &conditions); err != nil {
		return nil, fmt.Errorf("FindConditionsByPatientId: decode failed: %w", err)
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("FindConditionsByPatientIdPaginated: cursor error: %w", err)
	}

	return conditions, nil
}

func (m *mongoMedicalDb) UpdateCondition(
	ctx context.Context,
	id uuid.UUID,
	condition Condition,
) (Condition, error) {
	filter := bson.M{"_id": id}

	opts := options.FindOneAndReplace().SetReturnDocument(options.After)

	var updatedCondition Condition
	err := m.conditions.FindOneAndReplace(ctx, filter, condition, opts).
		Decode(&updatedCondition)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Condition{}, ErrNotFound
		}
		return Condition{}, fmt.Errorf("UpdateCondition failed: %w", err)
	}

	return updatedCondition, nil
}

func (m *mongoMedicalDb) FindConditionsByPatientIdAndDate(
	ctx context.Context,
	patientId uuid.UUID,
	date time.Time,
) ([]Condition, error) {
	conditions := make([]Condition, 0)

	year, month, day := date.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, date.Location())

	filter := bson.M{
		"patientId": patientId,
		"start":     bson.M{"$lte": startOfDay},
		"$or": []bson.M{
			{"end": bson.M{"$exists": false}},
			{"end": bson.M{"$gte": startOfDay}},
		},
	}

	findOptions := options.Find().SetSort(bson.D{{Key: "start", Value: 1}})

	cursor, err := m.conditions.Find(ctx, filter, findOptions)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return conditions, nil
		}
		return nil, fmt.Errorf("FindConditionsByPatientIdAndDate find failed: %w", err)
	}

	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			slog.Warn("Failed to close conditions cursor", "error", cerr.Error())
		}
	}()

	if err = cursor.All(ctx, &conditions); err != nil {
		slog.Error("Failed to decode condition documents from cursor", "error", err)
		return nil, fmt.Errorf("FindConditionsByPatientIdAndDate decode failed: %w", err)
	}

	if err = cursor.Err(); err != nil {
		slog.Error("Conditions cursor iteration error", "error", err)
		return nil, fmt.Errorf("FindConditionsByPatientIdAndDate cursor error: %w", err)
	}

	return conditions, nil
}

type Prescription struct {
	Id            uuid.UUID  `bson:"_id"                     json:"id"`
	PatientId     uuid.UUID  `bson:"patientId"               json:"patientId"`
	AppointmentId *uuid.UUID `bson:"appointmentId,omitempty" json:"appointmentId,omitempty"`
	Name          string     `bson:"name"                    json:"name"`
	Start         time.Time  `bson:"start"                   json:"start"`
	End           time.Time  `bson:"end"                     json:"end"`
	DoctorsNote   *string    `bson:"doctorsNote,omitempty"   json:"doctorsNote,omitempty"`
}

func (m *mongoMedicalDb) CreatePrescription(
	ctx context.Context,
	prescription Prescription,
) (Prescription, error) {
	prescription.Id = uuid.New()

	_, err := m.prescriptions.InsertOne(ctx, prescription)
	if err != nil {
		return Prescription{}, fmt.Errorf("CreatePrescription failed to insert document: %w", err)
	}

	return prescription, nil
}

func (m *mongoMedicalDb) PrescriptionById(ctx context.Context, id uuid.UUID) (Prescription, error) {
	filter := bson.M{"_id": id}
	var prescription Prescription

	err := m.prescriptions.FindOne(ctx, filter).Decode(&prescription)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Prescription{}, ErrNotFound
		}
		return Prescription{}, fmt.Errorf("PrescriptionById failed to find document: %w", err)
	}

	return prescription, nil
}

func (m *mongoMedicalDb) FindPrescriptionsByPatientId(
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
	cursor, err := m.prescriptions.Find(ctx, filter, findOptions)
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

func (m *mongoMedicalDb) UpdatePrescription(
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
	err := m.prescriptions.FindOneAndUpdate(ctx, filter, update, opts).
		Decode(&updatedPrescription)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Prescription{}, ErrNotFound
		}
		return Prescription{}, fmt.Errorf("UpdatePrescription failed: %w", err)
	}

	return updatedPrescription, nil
}

func (m *mongoMedicalDb) PrescriptionByAppointmentId(
	ctx context.Context,
	appointmentId uuid.UUID,
) ([]Prescription, error) {
	prescriptions := make([]Prescription, 0)

	filter := bson.M{"appointmentId": appointmentId}

	findOptions := options.Find().SetSort(bson.D{{Key: "start", Value: 1}})

	cursor, err := m.prescriptions.Find(ctx, filter, findOptions)
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

func (m *mongoMedicalDb) DeletePrescription(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	result, err := m.prescriptions.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("DeletePrescription failed: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrNotFound
	}

	return nil
}
