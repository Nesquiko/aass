package data

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDb struct {
	*mongo.Database
}

const (
	patientsCollection      = "patients"
	doctorsCollection       = "doctors"
	conditionsCollection    = "conditions"
	prescriptionsCollection = "prescriptions"
	appointmentsCollection  = "appointments"
	resourcesCollection     = "resources"
	reservationsCollection  = "reservations"
)

var Collections = []string{
	patientsCollection,
	doctorsCollection,
	conditionsCollection,
	prescriptionsCollection,
	appointmentsCollection,
	resourcesCollection,
	reservationsCollection,
}

var (
	tUUID       = reflect.TypeOf(uuid.UUID{})
	uuidSubtype = byte(0x04)
)

func ConnectMongo(ctx context.Context, uri string, db string) (*MongoDb, error) {
	mongoRegistry := bson.NewRegistry()
	mongoRegistry.RegisterTypeEncoder(tUUID, bson.ValueEncoderFunc(uuidEncodeValue))
	mongoRegistry.RegisterTypeDecoder(tUUID, bson.ValueDecoderFunc(uuidDecodeValue))

	opts := options.Client().ApplyURI(uri).SetRegistry(mongoRegistry)
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, fmt.Errorf("ConnectMongo: %w", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ConnectMongo: failed to ping mongo server: %w", err)
	}

	mongo := client.Database(db)
	err = initCollections(ctx, mongo, Collections)
	if err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ConnectMongo: failed to init collections: %w", err)
	}

	err = initIndexes(ctx, mongo)
	if err != nil {
		_ = client.Disconnect(ctx)
		return nil, fmt.Errorf("ConnectMongo: failed to init indexes: %w", err)
	}

	mongoDb := &MongoDb{Database: mongo}
	err = seedResources(ctx, mongoDb)

	return mongoDb, nil
}

func initCollections(ctx context.Context, mongoDb *mongo.Database, cols []string) error {
	if len(cols) == 0 {
		return nil
	}

	for _, collName := range cols {
		slog.Info("creating collection", "collection", collName)
		err := mongoDb.CreateCollection(ctx, collName)
		if err != nil {
			return fmt.Errorf(
				"initCollections: failed to create collection '%s': %w",
				collName,
				err,
			)
		}
	}

	return nil
}

func initIndexes(ctx context.Context, mongoDb *mongo.Database) error {
	indexes := map[string][]mongo.IndexModel{
		patientsCollection: {
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true).SetName("idx_patient_email_unique"),
			},
		},
		doctorsCollection: {
			{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetUnique(true).SetName("idx_doctor_email_unique"),
			},
		},
		conditionsCollection: {
			{
				Keys:    bson.D{{Key: "patientId", Value: 1}},
				Options: options.Index().SetName("idx_condition_patientId"),
			},
			{
				Keys:    bson.D{{Key: "patientId", Value: 1}, {Key: "start", Value: 1}},
				Options: options.Index().SetName("idx_condition_patientId_start"),
			},
		},
		prescriptionsCollection: {
			{
				Keys:    bson.D{{Key: "patientId", Value: 1}},
				Options: options.Index().SetName("idx_presription_patientId"),
			},
			{
				Keys:    bson.D{{Key: "patientId", Value: 1}, {Key: "start", Value: 1}},
				Options: options.Index().SetName("idx_presription_patientId_start"),
			},
		},
		appointmentsCollection: {
			{
				Keys:    bson.D{{Key: "status", Value: 1}},
				Options: options.Index().SetName("idx_appointment_status"),
			},
			{
				Keys:    bson.D{{Key: "appointmentDateTime", Value: 1}},
				Options: options.Index().SetName("idx_appointment_datetime"),
			},
		},
		resourcesCollection: {
			{
				Keys:    bson.D{{Key: "type", Value: 1}},
				Options: options.Index().SetName("idx_resource_type"),
			},
		},
		reservationsCollection: {
			{
				Keys: bson.D{
					{Key: "resourceId", Value: 1},
					{Key: "startTime", Value: 1},
				},
				Options: options.Index().SetName("idx_reservation_resource_time"),
			},
			{
				Keys:    bson.D{{Key: "appointmentId", Value: 1}},
				Options: options.Index().SetName("idx_reservation_appointmentId"),
			},
		},
	}

	for collName, indexModels := range indexes {
		coll := mongoDb.Collection(collName)
		indexNames, err := coll.Indexes().CreateMany(ctx, indexModels)
		if err != nil {
			slog.Warn(
				"Could not create one or more indexes (may already exist or other issue)",
				"collection", collName,
				"indexModels", indexModels,
				"error", err.Error(),
				"createdIndexNames",
				indexNames,
			)
		} else {
			slog.Info("Successfully created/verified indexes", "collection", collName, "indexNames", indexNames)
		}
	}

	return nil
}

func (m *MongoDb) Disconnect(ctx context.Context) error {
	return m.Database.Client().Disconnect(ctx)
}

func uuidEncodeValue(ec bson.EncodeContext, vw bson.ValueWriter, val reflect.Value) error {
	if !val.IsValid() || val.Type() != tUUID {
		return bsoncodec.ValueEncoderError{
			Name:     "uuidEncodeValue",
			Types:    []reflect.Type{tUUID},
			Received: val,
		}
	}
	b := val.Interface().(uuid.UUID)
	return vw.WriteBinaryWithSubtype(b[:], uuidSubtype)
}

func uuidDecodeValue(dc bson.DecodeContext, vr bson.ValueReader, val reflect.Value) error {
	if !val.CanSet() || val.Type() != tUUID {
		return bsoncodec.ValueDecoderError{
			Name:     "uuidDecodeValue",
			Types:    []reflect.Type{tUUID},
			Received: val,
		}
	}

	var data []byte
	var subtype byte
	var err error
	switch vrType := vr.Type(); vrType {
	case bson.TypeBinary:
		data, subtype, err = vr.ReadBinary()
		if subtype != uuidSubtype {
			return fmt.Errorf("unsupported binary subtype %v for UUID", subtype)
		}
	case bson.TypeNull:
		err = vr.ReadNull()
	case bson.TypeUndefined:
		err = vr.ReadUndefined()
	default:
		return fmt.Errorf("cannot decode %v into a UUID", vrType)
	}

	if err != nil {
		return err
	}
	uuid2, err := uuid.FromBytes(data)
	if err != nil {
		return err
	}
	val.Set(reflect.ValueOf(uuid2))
	return nil
}

func seedResources(ctx context.Context, db *MongoDb) error {
	facilities := []string{
		"Operating Room 1",
		"Consultation Room A",
		"Radiology Suite",
		"Physical Therapy Gym",
		"Emergency Bay 3",
	}

	medicines := []string{
		"Aspirin 100mg",
		"Amoxicillin 500mg",
		"Metformin 1000mg",
		"Salbutamol Inhaler",
		"Atorvastatin 20mg",
	}

	equipment := []string{
		"MRI Scanner",
		"X-Ray Machine",
		"Ultrasound Device",
		"Ventilator",
		"ECG Monitor",
	}

	for _, name := range facilities {
		_, err := db.CreateResource(ctx, name, ResourceTypeFacility)
		if err != nil {
			slog.Error("Failed to create facility", "name", name, "error", err)
			return fmt.Errorf("failed to seed facility %q: %w", name, err)
		}
	}

	for _, name := range medicines {
		_, err := db.CreateResource(ctx, name, ResourceTypeMedicine)
		if err != nil {
			slog.Error("Failed to create medicine", "name", name, "error", err)
			return fmt.Errorf("failed to seed medicine %q: %w", name, err)
		}
	}

	for _, name := range equipment {
		_, err := db.CreateResource(ctx, name, ResourceTypeEquipment)
		if err != nil {
			slog.Error("Failed to create equipment", "name", name, "error", err)
			return fmt.Errorf("failed to seed equipment %q: %w", name, err)
		}
	}

	return nil
}
