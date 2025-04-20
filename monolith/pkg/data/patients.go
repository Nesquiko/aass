package data

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Patient struct {
	Id        uuid.UUID `bson:"_id"       json:"id"`
	Email     string    `bson:"email"     json:"email"`
	FirstName string    `bson:"firstName" json:"firstName"`
	LastName  string    `bson:"lastName"  json:"lastName"`
}

func (m *MongoDb) CreatePatient(ctx context.Context, patient Patient) (Patient, error) {
	collection := m.Database.Collection(patientsCollection)
	patient.Id = uuid.New()

	_, err := collection.InsertOne(ctx, patient)
	if err != nil {
		var writeErr mongo.WriteException
		if errors.As(err, &writeErr) {
			for _, we := range writeErr.WriteErrors {
				if we.Code == 11000 {
					return Patient{}, ErrDuplicateEmail
				}
			}
		}
		return Patient{}, fmt.Errorf("CreatePatient: failed to insert document: %w", err)
	}

	return patient, nil
}

func (m *MongoDb) PatientById(ctx context.Context, id uuid.UUID) (Patient, error) {
	collection := m.Database.Collection(patientsCollection)
	filter := bson.M{"_id": id}
	var patient Patient

	err := collection.FindOne(ctx, filter).Decode(&patient)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Patient{}, ErrNotFound
		}
		return Patient{}, fmt.Errorf("PatientById: failed to find document: %w", err)
	}

	return patient, nil
}

func (m *MongoDb) PatientByEmail(ctx context.Context, email string) (Patient, error) {
	collection := m.Database.Collection(patientsCollection)
	filter := bson.M{"email": email}
	var patient Patient

	err := collection.FindOne(ctx, filter).Decode(&patient)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Patient{}, ErrNotFound
		}
		return Patient{}, fmt.Errorf("PatientByEmail: failed to find document: %w", err)
	}

	return patient, nil
}

func (m *MongoDb) patientExists(ctx context.Context, patientId uuid.UUID) error {
	patientsColl := m.Database.Collection(patientsCollection)
	filter := bson.M{"_id": patientId}

	count, err := patientsColl.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("patientExists: failed patient count check: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}

	return nil
}
