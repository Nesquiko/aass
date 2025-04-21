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

var (
	ErrNotFound          = errors.New("appointment not found")
	ErrDoctorUnavailable = errors.New("doctor unavailable at the specified time")
)

type MongoAppointmentDb struct {
	collection *mongo.Collection
}

const collection = "appointment"

func NewMongoAppointmentDb(
	ctx context.Context,
	uri string,
	db string,
) (*MongoAppointmentDb, error) {
	mongoDb, err := mongodb.ConnectMongo(ctx, uri, db)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to MongoDB", "uri", uri, "error", err)
		return nil, fmt.Errorf("NewMongoAppointmentDb: %w", err)
	}

	coll := mongoDb.Collection(collection)
	indexModel := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_appointment_status"),
		},
		{
			Keys:    bson.D{{Key: "appointmentDateTime", Value: 1}},
			Options: options.Index().SetName("idx_appointment_datetime"),
		},
	}
	for _, idx := range indexModel {
		_, err = coll.Indexes().CreateOne(ctx, idx)
		if err != nil {
			slog.WarnContext(ctx, "Could not ensure condition index", "error", err)
		}
	}

	return &MongoAppointmentDb{collection: coll}, nil
}

type Condition struct {
	Id        uuid.UUID  `bson:"_id"           json:"id"`
	PatientId uuid.UUID  `bson:"patientId"     json:"patientId"` // Reference to Patient._id
	Name      string     `bson:"name"          json:"name"`
	Start     time.Time  `bson:"start"         json:"start"`
	End       *time.Time `bson:"end,omitempty" json:"end,omitempty"`
}

func (m *MongoAppointmentDb) CreateCondition(
	ctx context.Context,
	condition Condition,
) (Condition, error) {
	condition.Id = uuid.New()

	_, err := m.collection.InsertOne(ctx, condition)
	if err != nil {
		return Condition{}, fmt.Errorf("CreateCondition: failed to insert document: %w", err)
	}

	return condition, nil
}

func (m *MongoAppointmentDb) ConditionById(ctx context.Context, id uuid.UUID) (Condition, error) {
	filter := bson.M{"_id": id}
	var condition Condition

	err := m.collection.FindOne(ctx, filter).Decode(&condition)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Condition{}, ErrNotFound
		}
		return Condition{}, fmt.Errorf("ConditionById: failed to find document: %w", err)
	}

	return condition, nil
}

func (m *MongoAppointmentDb) FindConditionsByPatientId(
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
	cursor, err := m.collection.Find(ctx, filter, opts)
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

func (m *MongoAppointmentDb) UpdateCondition(
	ctx context.Context,
	id uuid.UUID,
	condition Condition,
) (Condition, error) {
	filter := bson.M{"_id": id}

	opts := options.FindOneAndReplace().SetReturnDocument(options.After)

	var updatedCondition Condition
	err := m.collection.FindOneAndReplace(ctx, filter, condition, opts).
		Decode(&updatedCondition)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Condition{}, ErrNotFound
		}
		return Condition{}, fmt.Errorf("UpdateCondition failed: %w", err)
	}

	return updatedCondition, nil
}

func (m *MongoAppointmentDb) FindConditionsByPatientIdAndDate(
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

	cursor, err := m.collection.Find(ctx, filter, findOptions)
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

type Appointment struct {
	Id                  uuid.UUID `bson:"_id"                 json:"id"`
	PatientId           uuid.UUID `bson:"patientId"           json:"patientId"` // Reference to Patient._id
	DoctorId            uuid.UUID `bson:"doctorId"            json:"doctorId"`  // Reference to Doctor._id
	AppointmentDateTime time.Time `bson:"appointmentDateTime" json:"appointmentDateTime"`
	EndTime             time.Time `bson:"endTime"             json:"endTime"`

	Type        string     `bson:"type"                  json:"type"`
	Status      string     `bson:"status"                json:"status"`
	Reason      *string    `bson:"reason,omitempty"      json:"reason,omitempty"`
	ConditionId *uuid.UUID `bson:"conditionId,omitempty" json:"conditionId,omitempty"`

	CancellationReason *string `bson:"cancellationReason,omitempty" json:"cancellationReason,omitempty"`
	CancelledBy        *string `bson:"cancelledBy,omitempty"        json:"cancelledBy,omitempty"`
	DenialReason       *string `bson:"denialReason,omitempty"       json:"denialReason,omitempty"`
}

func (m *MongoAppointmentDb) CreateAppointment(
	ctx context.Context,
	appointment Appointment,
) (Appointment, error) {
	availabilityFilter := bson.M{
		"doctorId":            appointment.DoctorId,
		"appointmentDateTime": appointment.AppointmentDateTime,
		"status":              bson.M{"$nin": []string{"cancelled", "denied"}},
	}

	count, err := m.collection.CountDocuments(ctx, availabilityFilter)
	if err != nil {
		slog.Error(
			"Failed to count doctor appointments for availability check",
			"doctorId", appointment.DoctorId,
			"dateTime", appointment.AppointmentDateTime,
			"error", err,
		)
		return Appointment{}, fmt.Errorf(
			"CreateAppointment: doctor availability check failed: %w",
			err,
		)
	}

	if count > 0 {
		slog.Warn(
			"Attempted to schedule appointment when doctor is unavailable",
			"doctorId", appointment.DoctorId,
			"dateTime", appointment.AppointmentDateTime,
		)
		return Appointment{}, fmt.Errorf(
			"%w at %s",
			ErrDoctorUnavailable,
			appointment.AppointmentDateTime.Format(time.RFC3339),
		)
	}

	appointment.Id = uuid.New()
	_, err = m.collection.InsertOne(ctx, appointment)
	if err != nil {
		return Appointment{}, fmt.Errorf("CreateAppointment failed to insert document: %w", err)
	}

	return appointment, nil
}

func (m *MongoAppointmentDb) AppointmentById(
	ctx context.Context,
	id uuid.UUID,
) (Appointment, error) {
	filter := bson.M{"_id": id}

	var appt Appointment

	err := m.collection.FindOne(ctx, filter).Decode(&appt)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Appointment{}, ErrNotFound
		}
		return Appointment{}, fmt.Errorf("AppointmentById failed to find document: %w", err)
	}

	return appt, nil
}

func (m *MongoAppointmentDb) CancelAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
	by string,
	cancellationReason *string,
) error {
	if err := m.appointmentExists(ctx, appointmentId); err != nil {
		return fmt.Errorf("CancelAppointment appointment check failed: %w", err)
	}

	update := bson.M{
		"$set": bson.M{
			"status":             "cancelled",
			"cancellationReason": cancellationReason,
			"cancelledBy":        by,
		},
	}
	filter := bson.M{"_id": appointmentId}

	_, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("CancelAppointment failed to update appointment status: %w", err)
	}

	return nil
}

func (m *MongoAppointmentDb) DecideAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
	decision string,
	denyReason *string,
) (Appointment, error) {
	appointment, err := m.AppointmentById(ctx, appointmentId)
	if err != nil {
		return Appointment{}, fmt.Errorf("DecideAppointment: %w", err)
	}

	if appointment.Status != "requested" {
		return Appointment{}, fmt.Errorf(
			"DecideAppointment appointment %s is not in scheduled state",
			appointmentId,
		)
	}

	if decision == "accept" {
		_, err = m.scheduleAppointment(ctx, appointmentId)
		if err != nil {
			return Appointment{}, fmt.Errorf("DecideAppointment: %w", err)
		}
	} else if decision == "reject" {
		appointment, err = m.denyAppointment(ctx, appointmentId, denyReason)
		if err != nil {
			return Appointment{}, fmt.Errorf("DecideAppointment: %w", err)
		}
	} else {
		return Appointment{}, fmt.Errorf("DecideAppointment invalid decision %s for appointment %s", decision, appointmentId)
	}

	return appointment, nil
}

func (m *MongoAppointmentDb) AppointmentsByDoctorId(
	ctx context.Context,
	doctorId uuid.UUID,
	from time.Time,
	to *time.Time,
) ([]Appointment, error) {
	appts, err := m.appointmentsByIdField(ctx, "doctorId", doctorId, from, to)
	if err != nil {
		return nil, fmt.Errorf("AppointmentsByDoctorId cursor error: %w", err)
	}

	return appts, nil
}

func (m *MongoAppointmentDb) AppointmentsByPatientId(
	ctx context.Context,
	patientId uuid.UUID,
	from time.Time,
	to *time.Time,
) ([]Appointment, error) {
	appts, err := m.appointmentsByIdField(ctx, "patientId", patientId, from, to)
	if err != nil {
		return nil, fmt.Errorf("AppointmentsByPatientId cursor error: %w", err)
	}

	return appts, nil
}

func (m *MongoAppointmentDb) AppointmentsByDoctorIdAndDate(
	ctx context.Context,
	doctorId uuid.UUID,
	date time.Time,
) ([]Appointment, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1).Add(-1 * time.Nanosecond)

	appts, err := m.appointmentsByIdFieldAndDateRange(
		ctx,
		"doctorId",
		doctorId,
		startOfDay,
		endOfDay,
	)
	if err != nil {
		return nil, fmt.Errorf("AppointmentsByDoctorIdAndDate cursor error: %w", err)
	}

	return appts, nil
}

func (m *MongoAppointmentDb) RescheduleAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
	newDateTime time.Time,
) (Appointment, error) {
	appointment, err := m.AppointmentById(ctx, appointmentId)
	if err != nil {
		return Appointment{}, fmt.Errorf("RescheduleAppointment: %w", err)
	}

	if appointment.Status != "scheduled" && appointment.Status != "requested" {
		return Appointment{}, fmt.Errorf(
			"RescheduleAppointment appointment %s is not in a reschedulable state",
			appointmentId,
		)
	}

	availabilityFilter := bson.M{
		"doctorId":            appointment.DoctorId,
		"appointmentDateTime": newDateTime,
		"status":              bson.M{"$nin": []string{"cancelled", "denied"}},
	}

	count, err := m.collection.CountDocuments(ctx, availabilityFilter)
	if err != nil {
		return Appointment{}, fmt.Errorf(
			"RescheduleAppointment doctor availability check failed: %w",
			err,
		)
	}

	if count > 0 {
		return Appointment{}, fmt.Errorf(
			"%w at %s",
			ErrDoctorUnavailable,
			newDateTime.Format(time.RFC3339),
		)
	}

	update := bson.M{
		"$set": bson.M{
			"appointmentDateTime": newDateTime,
			"status":              "requested",
		},
	}
	filter := bson.M{"_id": appointmentId}

	_, err = m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return Appointment{}, fmt.Errorf(
			"RescheduleAppointment failed to update appointment: %w",
			err,
		)
	}

	return m.AppointmentById(ctx, appointmentId)
}

func (m *MongoAppointmentDb) AppointmentsByConditionId(
	ctx context.Context,
	conditionId uuid.UUID,
) ([]Appointment, error) {
	appointments := make([]Appointment, 0)
	filter := bson.M{"conditionId": conditionId}

	opts := options.Find().SetSort(bson.D{{Key: "appointmentDateTime", Value: -1}})

	cursor, err := m.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("AppointmentsByConditionId find failed: %w", err)
	}

	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			slog.Warn("Failed to close cursor in AppointmentsByConditionId", "error", cerr.Error())
		}
	}()

	if err = cursor.All(ctx, &appointments); err != nil {
		return nil, fmt.Errorf("AppointmentsByConditionId decode failed: %w", err)
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("AppointmentsByConditionId cursor error: %w", err)
	}

	return appointments, nil
}

func (m *MongoAppointmentDb) appointmentsByIdFieldAndDateRange(
	ctx context.Context,
	idField string,
	id uuid.UUID,
	start time.Time,
	end time.Time,
) ([]Appointment, error) {
	appointments := make([]Appointment, 0)

	filter := bson.M{
		idField: id,
		"appointmentDateTime": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "appointmentDateTime", Value: 1}})

	cursor, err := m.collection.Find(ctx, filter, opts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return appointments, nil
		}
		return nil, fmt.Errorf("appointmentsByIdFieldAndDateRange find failed: %w", err)
	}

	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			slog.Warn("Failed to close cursor", "error", cerr.Error())
		}
	}()

	if err = cursor.All(ctx, &appointments); err != nil {
		return nil, fmt.Errorf("appointmentsByIdFieldAndDateRange decode failed: %w", err)
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("appointmentsByIdFieldAndDateRange cursor error: %w", err)
	}

	return appointments, nil
}

func (m *MongoAppointmentDb) appointmentsByIdField(
	ctx context.Context,
	idField string,
	id uuid.UUID,
	from time.Time,
	to *time.Time,
) ([]Appointment, error) {
	appointments := make([]Appointment, 0)

	filter := bson.M{
		idField:               id,
		"appointmentDateTime": bson.M{"$gte": from},
	}
	if to != nil {
		filter["appointmentDateTime"] = bson.M{
			"$gte": from,
			"$lte": *to,
		}
	}
	opts := options.Find().SetSort(bson.D{{Key: "appointmentDateTime", Value: 1}})

	cursor, err := m.collection.Find(ctx, filter, opts)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return appointments, nil
		}
		return nil, fmt.Errorf("appointmentsByIdField find failed: %w", err)
	}

	defer func() {
		if cerr := cursor.Close(ctx); cerr != nil {
			slog.Warn("Failed to close cursor", "error", cerr.Error())
		}
	}()

	if err = cursor.All(ctx, &appointments); err != nil {
		return nil, fmt.Errorf("appointmentsByIdField decode failed: %w", err)
	}

	if err = cursor.Err(); err != nil {
		return nil, fmt.Errorf("appointmentsByIdField cursor error: %w", err)
	}

	return appointments, nil
}

func (m *MongoAppointmentDb) scheduleAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
) (Appointment, error) {
	update := bson.M{"$set": bson.M{"status": "scheduled"}}
	filter := bson.M{"_id": appointmentId}

	_, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return Appointment{}, fmt.Errorf("scheduleAppointment: %w", err)
	}

	return m.AppointmentById(ctx, appointmentId)
}

func (m *MongoAppointmentDb) denyAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
	reason *string,
) (Appointment, error) {
	update := bson.M{"$set": bson.M{"status": "denied", "denialReason": reason}}
	filter := bson.M{"_id": appointmentId}

	_, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return Appointment{}, fmt.Errorf("denyAppointment: %w", err)
	}

	return m.AppointmentById(ctx, appointmentId)
}

func (m *MongoAppointmentDb) appointmentExists(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"_id": id}

	count, err := m.collection.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("appointmentExists: failed patient count check: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}

	return nil
}
