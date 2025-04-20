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

type Condition struct {
	Id        uuid.UUID  `bson:"_id"           json:"id"`
	PatientId uuid.UUID  `bson:"patientId"     json:"patientId"` // Reference to Patient._id
	Name      string     `bson:"name"          json:"name"`
	Start     time.Time  `bson:"start"         json:"start"`
	End       *time.Time `bson:"end,omitempty" json:"end,omitempty"`
}

func (m *MongoDb) CreateCondition(ctx context.Context, condition Condition) (Condition, error) {
	if err := m.patientExists(ctx, condition.PatientId); err != nil {
		return Condition{}, fmt.Errorf("CreateCondition patient check error: %w", err)
	}

	collection := m.Database.Collection(conditionsCollection)
	condition.Id = uuid.New()

	_, err := collection.InsertOne(ctx, condition)
	if err != nil {
		return Condition{}, fmt.Errorf("CreateCondition: failed to insert document: %w", err)
	}

	return condition, nil
}

func (m *MongoDb) ConditionById(ctx context.Context, id uuid.UUID) (Condition, error) {
	collection := m.Database.Collection(conditionsCollection)
	filter := bson.M{"_id": id}
	var condition Condition

	err := collection.FindOne(ctx, filter).Decode(&condition)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Condition{}, ErrNotFound
		}
		return Condition{}, fmt.Errorf("ConditionById: failed to find document: %w", err)
	}

	return condition, nil
}

func (m *MongoDb) FindConditionsByPatientId(
	ctx context.Context,
	patientId uuid.UUID,
	from time.Time,
	to *time.Time,
) ([]Condition, error) {
	collection := m.Database.Collection(conditionsCollection)
	conditions := make([]Condition, 0)

	filter := bson.M{"patientId": patientId}
	if to != nil {
		filter["start"] = bson.M{"$lte": *to}
	}
	filter["$or"] = []bson.M{{"end": bson.M{"$gte": from}}, {"end": nil}}

	opts := options.Find().SetSort(bson.D{{Key: "start", Value: 1}})
	cursor, err := collection.Find(ctx, filter, opts)
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

func (m *MongoDb) UpdateCondition(
	ctx context.Context,
	id uuid.UUID,
	condition Condition,
) (Condition, error) {
	if err := m.patientExists(ctx, condition.PatientId); err != nil {
		return Condition{}, fmt.Errorf(
			"UpdateCondition patient check error: %w",
			err,
		)
	}

	collection := m.Database.Collection(conditionsCollection)
	filter := bson.M{"_id": id}

	opts := options.FindOneAndReplace().SetReturnDocument(options.After)

	var updatedCondition Condition
	err := collection.FindOneAndReplace(ctx, filter, condition, opts).
		Decode(&updatedCondition)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Condition{}, ErrNotFound
		}
		return Condition{}, fmt.Errorf("UpdateCondition failed: %w", err)
	}

	return updatedCondition, nil
}

func (m *MongoDb) FindConditionsByPatientIdAndDate(
	ctx context.Context,
	patientId uuid.UUID,
	date time.Time,
) ([]Condition, error) {
	collection := m.Database.Collection(conditionsCollection)
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

	cursor, err := collection.Find(ctx, filter, findOptions)
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

func (m *MongoDb) conditionExists(ctx context.Context, id uuid.UUID) error {
	conditionsColl := m.Database.Collection(conditionsCollection)
	filter := bson.M{"_id": id}

	count, err := conditionsColl.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("conditionExists failed condition count check: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}

	return nil
}
