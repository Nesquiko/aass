package main

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

const (
	patiensCollection = "patients"
	doctorsCollection = "doctors"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrDuplicateEmail = errors.New("email address already exists")
)

type Patient struct {
	Id        uuid.UUID `bson:"_id"       json:"id"`
	Email     string    `bson:"email"     json:"email"`
	FirstName string    `bson:"firstName" json:"firstName"`
	LastName  string    `bson:"lastName"  json:"lastName"`
}

type Doctor struct {
	Id             uuid.UUID `bson:"_id"            json:"id"`
	Email          string    `bson:"email"          json:"email"`
	FirstName      string    `bson:"firstName"      json:"firstName"`
	LastName       string    `bson:"lastName"       json:"lastName"`
	Specialization string    `bson:"specialization" json:"specialization"`
}

type mongoUserDb struct {
	patients *mongo.Collection
	doctors  *mongo.Collection
}

func newMongoUserDb(ctx context.Context, uri string, db string) (mongoUserDb, error) {
	mongoDb, err := mongodb.ConnectMongo(ctx, uri, db)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to MongoDB", "uri", uri, "error", err)
		return mongoUserDb{}, fmt.Errorf("NewMongoUserDb: %w", err)
	}

	doctorsColl := mongoDb.Collection(doctorsCollection)
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("idx_doctor_email_unique"),
	}
	_, err = doctorsColl.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		slog.WarnContext(
			ctx,
			"Could not ensure doctor email index (may already exist)",
			"error",
			err,
		)
	}

	patientsColl := mongoDb.Collection(patiensCollection)
	indexModel = mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("idx_patient_email_unique"),
	}
	_, err = patientsColl.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		slog.WarnContext(
			ctx,
			"Could not ensure patient email index (may already exist)",
			"error",
			err,
		)
	}

	return mongoUserDb{doctors: doctorsColl, patients: patientsColl}, nil
}

func (db mongoUserDb) Disconnect(ctx context.Context) error {
	return db.patients.Database().Client().Disconnect(ctx)
}

func (db *mongoUserDb) FindPatientById(ctx context.Context, id uuid.UUID) (Patient, error) {
	filter := bson.M{"_id": id}
	var patient Patient

	err := db.patients.FindOne(ctx, filter).Decode(&patient)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Patient{}, ErrNotFound
		}
		return Patient{}, fmt.Errorf("FindPatientById failed to find patient by id %s: %w", id, err)
	}

	return patient, nil
}

func (db *mongoUserDb) CreatePatient(ctx context.Context, patient Patient) (Patient, error) {
	patient.Id = uuid.New()

	_, err := db.patients.InsertOne(ctx, patient)
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

func (db *mongoUserDb) FindPatientByEmail(ctx context.Context, email string) (Patient, error) {
	filter := bson.M{"email": email}
	var patient Patient

	err := db.patients.FindOne(ctx, filter).Decode(&patient)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Patient{}, ErrNotFound
		}
		return Patient{}, fmt.Errorf("FindPatientById failed to find document: %w", err)
	}

	return patient, nil
}

func (m *mongoUserDb) CreateDoctor(ctx context.Context, doctor Doctor) (Doctor, error) {
	doctor.Id = uuid.New()

	_, err := m.doctors.InsertOne(ctx, doctor)
	if err != nil {
		var writeErr mongo.WriteException
		if errors.As(err, &writeErr) {
			for _, we := range writeErr.WriteErrors {
				if we.Code == 11000 {
					return Doctor{}, ErrDuplicateEmail
				}
			}
		}
		return Doctor{}, fmt.Errorf("CreateDoctor failed to insert document: %w", err)
	}

	return doctor, nil
}

func (m *mongoUserDb) DoctorById(ctx context.Context, id uuid.UUID) (Doctor, error) {
	filter := bson.M{"_id": id}
	var doctor Doctor

	err := m.doctors.FindOne(ctx, filter).Decode(&doctor)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Doctor{}, ErrNotFound
		}
		return Doctor{}, fmt.Errorf("DoctorById failed to find document: %w", err)
	}

	return doctor, nil
}

func (m *mongoUserDb) DoctorByEmail(ctx context.Context, email string) (Doctor, error) {
	filter := bson.M{"email": email}
	var doctor Doctor

	err := m.doctors.FindOne(ctx, filter).Decode(&doctor)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Doctor{}, ErrNotFound
		}
		return Doctor{}, fmt.Errorf("DoctorByEmail failed to find document: %w", err)
	}

	return doctor, nil
}

func (m *mongoUserDb) GetAllDoctors(ctx context.Context) ([]Doctor, error) {
	doctors := make([]Doctor, 0)
	filter := bson.M{}
	findOptions := options.Find().SetSort(
		bson.D{{Key: "lastName", Value: 1}, {Key: "firstName", Value: 1}},
	)

	cursor, err := m.doctors.Find(ctx, filter, findOptions)
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
