package pkg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/Nesquiko/aass/common/mongodb"
)

type Patient struct {
	Id        uuid.UUID `bson:"_id"       json:"id"`
	Email     string    `bson:"email"     json:"email"`
	FirstName string    `bson:"firstName" json:"firstName"`
	LastName  string    `bson:"lastName"  json:"lastName"`
}

var (
	ErrNotFound       = errors.New("patient not found")
	ErrDuplicateEmail = errors.New("patient email address already exists")
)

type MongoPatientDb struct {
	collection *mongo.Collection
}

const patiensCollection = "patients"

func NewMongoPatientDb(ctx context.Context, uri string, db string) (*MongoPatientDb, error) {
	mongoDb, err := mongodb.ConnectMongo(ctx, uri, db)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to MongoDB", "uri", uri, "error", err)
		return nil, fmt.Errorf("NewMongoPatientDb: %w", err)
	}

	coll := mongoDb.Collection(patiensCollection)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("idx_patient_email_unique"),
	}
	_, err = coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		slog.WarnContext(
			ctx,
			"Could not ensure patient email index (may already exist)",
			"error",
			err,
		)
	}

	return &MongoPatientDb{collection: coll}, nil
}

func (db *MongoPatientDb) FindPatientById(ctx context.Context, id uuid.UUID) (Patient, error) {
	filter := bson.M{"_id": id}
	var patient Patient

	err := db.collection.FindOne(ctx, filter).Decode(&patient)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Patient{}, ErrNotFound
		}
		return Patient{}, fmt.Errorf("FindPatientById failed to find patient by id %s: %w", id, err)
	}

	return patient, nil
}

func (db *MongoPatientDb) CreatePatient(ctx context.Context, patient Patient) (Patient, error) {
	patient.Id = uuid.New()

	_, err := db.collection.InsertOne(ctx, patient)
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

func (db *MongoPatientDb) FindPatientByEmail(ctx context.Context, email string) (Patient, error) {
	filter := bson.M{"email": email}
	var patient Patient

	err := db.collection.FindOne(ctx, filter).Decode(&patient)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Patient{}, ErrNotFound
		}
		return Patient{}, fmt.Errorf("FindPatientById failed to find document: %w", err)
	}

	return patient, nil
}

func (db *MongoPatientDb) Disconnect(ctx context.Context) error {
	return db.collection.Database().Client().Disconnect(ctx)
}
