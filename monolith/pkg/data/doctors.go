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

type Doctor struct {
	Id             uuid.UUID `bson:"_id"            json:"id"`
	Email          string    `bson:"email"          json:"email"`
	FirstName      string    `bson:"firstName"      json:"firstName"`
	LastName       string    `bson:"lastName"       json:"lastName"`
	Specialization string    `bson:"specialization" json:"specialization"`
}

func (m *MongoDb) CreateDoctor(ctx context.Context, doctor Doctor) (Doctor, error) {
	collection := m.Database.Collection(doctorsCollection)
	doctor.Id = uuid.New()

	_, err := collection.InsertOne(ctx, doctor)
	if err != nil {
		var writeErr mongo.WriteException
		if errors.As(err, &writeErr) {
			for _, we := range writeErr.WriteErrors {
				if we.Code == 11000 {
					return Doctor{}, ErrDuplicateEmail
				}
			}
		}
		return Doctor{}, fmt.Errorf("CreateDoctor: failed to insert document: %w", err)
	}

	return doctor, nil
}

func (m *MongoDb) DoctorById(ctx context.Context, id uuid.UUID) (Doctor, error) {
	collection := m.Database.Collection(doctorsCollection)
	filter := bson.M{"_id": id}
	var doctor Doctor

	err := collection.FindOne(ctx, filter).Decode(&doctor)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Doctor{}, ErrNotFound
		}
		return Doctor{}, fmt.Errorf("DoctorById: failed to find document: %w", err)
	}

	return doctor, nil
}

func (m *MongoDb) DoctorByEmail(ctx context.Context, email string) (Doctor, error) {
	collection := m.Database.Collection(doctorsCollection)
	filter := bson.M{"email": email}
	var doctor Doctor

	err := collection.FindOne(ctx, filter).Decode(&doctor)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Doctor{}, ErrNotFound
		}
		return Doctor{}, fmt.Errorf("DoctorByEmail: failed to find document: %w", err)
	}

	return doctor, nil
}

func (m *MongoDb) AvailableDoctors(ctx context.Context, dateTime time.Time) ([]Doctor, error) {
	apptCollection := m.Database.Collection(appointmentsCollection)
	doctorCollection := m.Database.Collection(doctorsCollection)

	busyStatuses := []string{"requested", "Scheduled"}

	appointmentFilter := bson.M{
		"appointmentDateTime": bson.M{"$lte": dateTime},
		"endTime":             bson.M{"$gt": dateTime},
		"status":              bson.M{"$in": busyStatuses},
	}

	findOptions := options.Find().SetProjection(bson.M{"doctorId": 1, "_id": 0})

	cursor, err := apptCollection.Find(ctx, appointmentFilter, findOptions)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
		} else {
			return nil, fmt.Errorf("AvailableDoctors find conflicting appointments failed: %w", err)
		}
	}

	busyDoctorIdsMap := make(map[uuid.UUID]struct{})
	if cursor != nil {
		defer func() {
			if cerr := cursor.Close(ctx); cerr != nil {
				slog.Warn("Failed to close conflicting appointments cursor", "error", cerr.Error())
			}
		}()

		var results []struct {
			DoctorId uuid.UUID `bson:"doctorId"`
		}
		if err = cursor.All(ctx, &results); err != nil {
			slog.Error("Failed to decode conflicting appointment documents", "error", err)
			return nil, fmt.Errorf(
				"AvailableDoctors decode conflicting appointments failed: %w",
				err,
			)
		}

		for _, result := range results {
			busyDoctorIdsMap[result.DoctorId] = struct{}{}
		}

		if err = cursor.Err(); err != nil {
			slog.Error("Conflicting appointments cursor iteration error", "error", err)
			return nil, fmt.Errorf(
				"AvailableDoctors conflicting appointments cursor error: %w",
				err,
			)
		}
	}

	busyDoctorIds := make([]uuid.UUID, 0, len(busyDoctorIdsMap))
	for id := range busyDoctorIdsMap {
		busyDoctorIds = append(busyDoctorIds, id)
	}

	doctorFilter := bson.M{
		"_id": bson.M{"$nin": busyDoctorIds},
	}

	doctorCursor, err := doctorCollection.Find(
		ctx,
		doctorFilter,
		options.Find().SetSort(bson.D{{Key: "lastName", Value: 1}, {Key: "firstName", Value: 1}}),
	)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []Doctor{}, nil
		}
		return nil, fmt.Errorf("AvailableDoctors find available doctors failed: %w", err)
	}

	defer func() {
		if cerr := doctorCursor.Close(ctx); cerr != nil {
			slog.Warn(
				"Failed to close available doctors cursor",
				"error",
				cerr.Error(),
			)
		}
	}()

	var availableDoctors []Doctor
	if err = doctorCursor.All(ctx, &availableDoctors); err != nil {
		slog.Error("Failed to decode available doctor documents", "error", err)
		return nil, fmt.Errorf("AvailableDoctors decode available doctors failed: %w", err)
	}

	if err = doctorCursor.Err(); err != nil {
		slog.Error("Available doctors cursor iteration error", "error", err)
		return nil, fmt.Errorf("AvailableDoctors available doctors cursor error: %w", err)
	}

	return availableDoctors, nil
}

func (m *MongoDb) doctorExists(ctx context.Context, id uuid.UUID) error {
	doctorsColl := m.Database.Collection(doctorsCollection)
	filter := bson.M{"_id": id}

	count, err := doctorsColl.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("doctorExists failed doctor count check: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}

	return nil
}

func (m *MongoDb) GetAllDoctors(ctx context.Context) ([]Doctor, error) {
	collection := m.Database.Collection(doctorsCollection)
	doctors := make([]Doctor, 0)

	filter := bson.M{}

	findOptions := options.Find().SetSort(
		bson.D{{Key: "lastName", Value: 1}, {Key: "firstName", Value: 1}},
	)

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return doctors, nil
		}
		return nil, fmt.Errorf("GetAllDoctors find failed: %w", err)
	}

	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			slog.Warn("Failed to close doctors cursor", "error", cerr.Error())
		}
	}()

	if err = cursor.All(ctx, &doctors); err != nil {
		slog.Error("Failed to decode doctor documents from cursor", "error", err)
		return nil, fmt.Errorf("GetAllDoctors decode failed: %w", err)
	}

	if err = cursor.Err(); err != nil {
		slog.Error("Doctors cursor iteration error", "error", err)
		return nil, fmt.Errorf("GetAllDoctors cursor error: %w", err)
	}

	return doctors, nil
}
