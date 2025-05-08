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
	appointmentCollection = "appointments"
)

var (
	ErrNotFound          = errors.New("resource not found")
	ErrDoctorUnavailable = errors.New("doctor unavailable at the specified time")
)

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

	Medicines  []Resource `bson:"medicines,omitempty"  json:"medicines,omitempty"`
	Facilities []Resource `bson:"facilities,omitempty" json:"facilities,omitempty"`
	Equipment  []Resource `bson:"equipment,omitempty"  json:"equipment,omitempty"`
}

type ResourceType string

const (
	ResourceTypeMedicine  ResourceType = "medicine"
	ResourceTypeFacility  ResourceType = "facility"
	ResourceTypeEquipment ResourceType = "equipment"
)

type Resource struct {
	Id   uuid.UUID    `bson:"_id"  json:"id"`
	Name string       `bson:"name" json:"name"`
	Type ResourceType `bson:"type" json:"type"`
}

type mongoAppointmentDb struct {
	appointments *mongo.Collection
}

func newMongoAppointmentDb(ctx context.Context, uri string, db string) (mongoAppointmentDb, error) {
	mongoDb, err := mongodb.ConnectMongo(ctx, uri, db)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to connect to MongoDB", "uri", uri, "error", err)
		return mongoAppointmentDb{}, fmt.Errorf("newMongoAppointmentDb: %w", err)
	}

	appointmentColl := mongoDb.Collection(appointmentCollection)
	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_appointment_status"),
		},
		{
			Keys:    bson.D{{Key: "appointmentDateTime", Value: 1}},
			Options: options.Index().SetName("idx_appointment_datetime"),
		},
	}

	for _, idx := range indexModels {
		_, err = appointmentColl.Indexes().CreateOne(ctx, idx)
		if err != nil {
			slog.WarnContext(
				ctx,
				"Could not ensure appointment index (may already exist)",
				"error",
				err,
			)
		}
	}

	return mongoAppointmentDb{appointments: appointmentColl}, nil
}

func (db mongoAppointmentDb) Disconnect(ctx context.Context) error {
	return db.appointments.Database().Client().Disconnect(ctx)
}

func (m *mongoAppointmentDb) CreateAppointment(
	ctx context.Context,
	appointment Appointment,
) (Appointment, error) {
	appointmentsColl := m.appointments
	availabilityFilter := bson.M{
		"doctorId":            appointment.DoctorId,
		"appointmentDateTime": appointment.AppointmentDateTime,
		"status":              bson.M{"$nin": []string{"cancelled", "denied"}},
	}

	count, err := appointmentsColl.CountDocuments(ctx, availabilityFilter)
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
	_, err = appointmentsColl.InsertOne(ctx, appointment)
	if err != nil {
		return Appointment{}, fmt.Errorf("CreateAppointment: failed to insert document: %w", err)
	}

	return appointment, nil
}

func (m *mongoAppointmentDb) AppointmentById(
	ctx context.Context,
	id uuid.UUID,
) (Appointment, error) {
	appointmentsColl := m.appointments
	filter := bson.M{"_id": id}

	var appt Appointment

	err := appointmentsColl.FindOne(ctx, filter).Decode(&appt)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Appointment{}, ErrNotFound
		}
		return Appointment{}, fmt.Errorf("AppointmentById failed to find document: %w", err)
	}

	return appt, nil
}

func (m *mongoAppointmentDb) CancelAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
	by string,
	cancellationReason *string,
) error {
	if err := m.appointmentExists(ctx, appointmentId); err != nil {
		return fmt.Errorf("CancelAppointment appointment check failed: %w", err)
	}

	appointmentsColl := m.appointments
	update := bson.M{
		"$set": bson.M{
			"status":             "cancelled",
			"cancellationReason": cancellationReason,
			"cancelledBy":        by,
		},
	}
	filter := bson.M{"_id": appointmentId}

	_, err := appointmentsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("CancelAppointment failed to update appointment status: %w", err)
	}

	return nil
}

func (m *mongoAppointmentDb) DecideAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
	decision string,
	denyReason *string,
	resources []Resource,
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

		appointment, err = m.updateAppointmentResources(ctx, appointmentId, resources)
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

func (m *mongoAppointmentDb) AppointmentsByDoctorId(
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

func (m *mongoAppointmentDb) AppointmentsByPatientId(
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

func (m *mongoAppointmentDb) AppointmentsByDoctorIdAndDate(
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

func (m *mongoAppointmentDb) RescheduleAppointment(
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

	appointmentsColl := m.appointments
	count, err := appointmentsColl.CountDocuments(ctx, availabilityFilter)
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

	_, err = appointmentsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return Appointment{}, fmt.Errorf(
			"RescheduleAppointment failed to update appointment: %w",
			err,
		)
	}

	return m.AppointmentById(ctx, appointmentId)
}

func (m *mongoAppointmentDb) AppointmentsByConditionId(
	ctx context.Context,
	conditionId uuid.UUID,
) ([]Appointment, error) {
	appointmentsColl := m.appointments
	appointments := make([]Appointment, 0)
	filter := bson.M{"conditionId": conditionId}

	opts := options.Find().SetSort(bson.D{{Key: "appointmentDateTime", Value: -1}})

	cursor, err := appointmentsColl.Find(ctx, filter, opts)
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

func (m *mongoAppointmentDb) appointmentsByIdFieldAndDateRange(
	ctx context.Context,
	idField string,
	id uuid.UUID,
	start time.Time,
	end time.Time,
) ([]Appointment, error) {
	appointmentsColl := m.appointments
	appointments := make([]Appointment, 0)

	filter := bson.M{
		idField: id,
		"appointmentDateTime": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "appointmentDateTime", Value: 1}})

	cursor, err := appointmentsColl.Find(ctx, filter, opts)
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

func (m *mongoAppointmentDb) appointmentsByIdField(
	ctx context.Context,
	idField string,
	id uuid.UUID,
	from time.Time,
	to *time.Time,
) ([]Appointment, error) {
	appointmentsColl := m.appointments
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

	cursor, err := appointmentsColl.Find(ctx, filter, opts)
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

func (m *mongoAppointmentDb) scheduleAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
) (Appointment, error) {
	appointmentsColl := m.appointments
	update := bson.M{"$set": bson.M{"status": "scheduled"}}
	filter := bson.M{"_id": appointmentId}

	_, err := appointmentsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return Appointment{}, fmt.Errorf("scheduleAppointment: %w", err)
	}

	return m.AppointmentById(ctx, appointmentId)
}

func (m *mongoAppointmentDb) denyAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
	reason *string,
) (Appointment, error) {
	appointmentsColl := m.appointments
	update := bson.M{"$set": bson.M{"status": "denied", "denialReason": reason}}
	filter := bson.M{"_id": appointmentId}

	_, err := appointmentsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return Appointment{}, fmt.Errorf("denyAppointment: %w", err)
	}

	return m.AppointmentById(ctx, appointmentId)
}

func (m *mongoAppointmentDb) updateAppointmentResources(
	ctx context.Context,
	appointmentId uuid.UUID,
	resources []Resource,
) (Appointment, error) {
	appointmentsColl := m.appointments

	var facilities, equipment, medicine []Resource
	for _, resource := range resources {
		switch resource.Type {
		case ResourceTypeFacility:
			facilities = append(facilities, resource)
		case ResourceTypeEquipment:
			equipment = append(equipment, resource)
		case ResourceTypeMedicine:
			medicine = append(medicine, resource)
		}
	}

	update := bson.M{
		"$set": bson.M{
			"facilities": facilities,
			"equipment":  equipment,
			"medicines":  medicine,
		},
	}
	filter := bson.M{"_id": appointmentId}

	_, err := appointmentsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return Appointment{}, fmt.Errorf("updateAppointmentResources: %w", err)
	}

	return m.AppointmentById(ctx, appointmentId)
}

func (m *mongoAppointmentDb) appointmentExists(ctx context.Context, id uuid.UUID) error {
	appointmentsColl := m.appointments
	filter := bson.M{"_id": id}

	count, err := appointmentsColl.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("appointmentExists: failed patient count check: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}

	return nil
}

func (m *mongoAppointmentDb) UpdateAppointmentResources(
	ctx context.Context,
	appointmentId uuid.UUID,
	facilities []Resource,
	equipment []Resource,
	medicine []Resource,
) (Appointment, error) {
	_, err := m.AppointmentById(ctx, appointmentId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Appointment{}, ErrNotFound
		}
		return Appointment{}, fmt.Errorf("UpdateAppointmentResources: failed check: %w", err)
	}

	appointmentsColl := m.appointments
	filter := bson.M{"_id": appointmentId}
	update := bson.M{
		"$set": bson.M{
			"facilities": facilities,
			"equipment":  equipment,
			"medicines":  medicine,
		},
	}

	_, err = appointmentsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return Appointment{}, fmt.Errorf("UpdateAppointmentResources: failed update: %w", err)
	}

	return m.AppointmentById(ctx, appointmentId)
}
