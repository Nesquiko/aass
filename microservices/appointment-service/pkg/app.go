package pkg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Nesquiko/aass/appointment-service/api"
	doctorapi "github.com/Nesquiko/aass/appointment-service/api/doctor-api"
	patientapi "github.com/Nesquiko/aass/appointment-service/api/patient-api"
)

var (
	ErrAppointmentNotFound = errors.New("appointment not found")
	ErrDoctorNotAvailable  = errors.New("doctor is not available at the requested time")
	ErrAppointmentBadState = errors.New(
		"appointment is not in a state valid for the requested operation",
	)
	ErrInvalidDecisionAction = errors.New("invalid decision action provided")
	ErrMissingDenialReason   = errors.New("denial reason is required when rejecting an appointment")
	ErrIdMismatch            = errors.New("provided ID does not match appointment record")
)

const defaultAppointmentDuration = 1 * time.Hour

type AppointmentApp struct {
	db        *MongoAppointmentDb
	docClient *doctorapi.ClientWithResponses
	patClient *patientapi.ClientWithResponses
}

// NewAppointmentApp creates a new AppointmentApp instance
func NewAppointmentApp(db *MongoAppointmentDb) *AppointmentApp {
	return &AppointmentApp{db: db}
}

func (a *AppointmentApp) RequestAppointment(
	ctx context.Context,
	req api.NewAppointmentRequest,
) (Appointment, error) {
	// Map API request to DB model
	dbAppointment := Appointment{
		PatientId:           req.PatientId,
		DoctorId:            req.DoctorId,
		AppointmentDateTime: req.AppointmentDateTime,
		EndTime: req.AppointmentDateTime.Add(
			defaultAppointmentDuration,
		), // Calculate end time
		Type:        string(req.Type),
		Status:      string(api.Requested), // Initial status
		Reason:      req.Reason,
		ConditionId: req.ConditionId,
	}

	createdAppt, err := a.db.CreateAppointment(ctx, dbAppointment)
	if err != nil {
		if errors.Is(err, ErrDoctorUnavailable) {
			slog.WarnContext(
				ctx,
				"Doctor unavailable on create",
				"doctorId",
				req.DoctorId,
				"time",
				req.AppointmentDateTime,
			)
			return Appointment{}, ErrDoctorNotAvailable // Return specific app error
		}
		slog.ErrorContext(ctx, "Failed to create appointment in DB", "error", err)
		return Appointment{}, fmt.Errorf("failed to create appointment: %w", err)
	}

	return createdAppt, nil
}

func (a *AppointmentApp) CancelAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
	req api.AppointmentCancellation,
) error {
	err := a.db.CancelAppointment(ctx, appointmentId, string(req.By), &req.Reason)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrAppointmentNotFound
		}
		slog.ErrorContext(
			ctx,
			"Failed to cancel appointment in DB",
			"appointmentId",
			appointmentId,
			"error",
			err,
		)
		return fmt.Errorf("failed to cancel appointment: %w", err)
	}
	return nil
}

func (a *AppointmentApp) RescheduleAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
	req api.AppointmentReschedule,
) (Appointment, error) {
	updatedAppt, err := a.db.RescheduleAppointment(ctx, appointmentId, req.NewAppointmentDateTime)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Appointment{}, ErrAppointmentNotFound
		}
		if errors.Is(err, ErrDoctorUnavailable) {
			slog.WarnContext(
				ctx,
				"Doctor unavailable on reschedule",
				"appointmentId",
				appointmentId,
				"newTime",
				req.NewAppointmentDateTime,
			)
			return Appointment{}, ErrDoctorNotAvailable
		}
		// Check for the specific error message from the DB layer about state
		if strings.Contains(err.Error(), "not in a reschedulable state") {
			return Appointment{}, ErrAppointmentBadState
		}
		slog.ErrorContext(
			ctx,
			"Failed to reschedule appointment in DB",
			"appointmentId",
			appointmentId,
			"error",
			err,
		)
		return Appointment{}, fmt.Errorf("failed to reschedule appointment: %w", err)
	}
	// Note: Reschedule sets status back to 'requested' in DB layer
	return updatedAppt, nil
}

func (a *AppointmentApp) DecideAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
	req api.AppointmentDecision,
) (Appointment, error) {
	// Basic validation
	if req.Action == api.Reject && (req.Reason == nil || *req.Reason == "") {
		return Appointment{}, ErrMissingDenialReason
	}

	updatedAppt, err := a.db.DecideAppointment(ctx, appointmentId, string(req.Action), req.Reason)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Appointment{}, ErrAppointmentNotFound
		}
		// Check for specific error messages from DB layer
		if strings.Contains(
			err.Error(),
			"is not in scheduled state",
		) { // DB layer uses 'scheduled' but means 'requested'
			return Appointment{}, ErrAppointmentBadState
		}
		if strings.Contains(err.Error(), "invalid decision") {
			return Appointment{}, ErrInvalidDecisionAction
		}
		slog.ErrorContext(
			ctx,
			"Failed to decide appointment in DB",
			"appointmentId",
			appointmentId,
			"action",
			req.Action,
			"error",
			err,
		)
		return Appointment{}, fmt.Errorf("failed to decide appointment: %w", err)
	}

	return updatedAppt, nil
}

func (a *AppointmentApp) GetAppointmentById(
	ctx context.Context,
	appointmentId uuid.UUID,
) (Appointment, error) {
	appt, err := a.db.AppointmentById(ctx, appointmentId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return Appointment{}, ErrAppointmentNotFound
		}
		slog.ErrorContext(
			ctx,
			"Failed to get appointment by ID from DB",
			"appointmentId",
			appointmentId,
			"error",
			err,
		)
		return Appointment{}, fmt.Errorf("failed to get appointment: %w", err)
	}
	return appt, nil
}

// Helper to parse date/range params and call appropriate DB function
func (a *AppointmentApp) getAppointments(
	ctx context.Context,
	idField string, // "patientId" or "doctorId"
	id uuid.UUID,
	paramsDate *api.QueryDate,
	paramsFrom *api.QueryFrom,
	paramsTo *api.QueryTo,
) ([]Appointment, error) {
	var appointments []Appointment
	var err error

	if paramsDate != nil {
		// Use specific date query
		date := paramsDate.Time
		appointments, err = a.db.appointmentsByIdFieldAndDateRange(
			ctx,
			idField,
			id,
			time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()),
			time.Date(
				date.Year(),
				date.Month(),
				date.Day(),
				23,
				59,
				59,
				999999999,
				date.Location(),
			),
		)
	} else {
		// Use date range query (defaulting 'from' to beginning of time if nil)
		from := time.Time{} // Default to earliest possible time
		if paramsFrom != nil {
			from = paramsFrom.Time
		}

		var toPtr *time.Time
		if paramsTo != nil {
			// Make 'to' inclusive by going to end of day
			t := paramsTo.Time
			endOfDay := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
			toPtr = &endOfDay
		}

		appointments, err = a.db.appointmentsByIdField(ctx, idField, id, from, toPtr)
	}

	if err != nil {
		// DB layer handles ErrNoDocuments returning empty slice, so only log real errors
		slog.ErrorContext(
			ctx,
			"Failed to get appointments from DB",
			"idField",
			idField,
			"id",
			id,
			"error",
			err,
		)
		return nil, fmt.Errorf("failed to retrieve appointments: %w", err)
	}

	return appointments, nil
}

func (a *AppointmentApp) GetDoctorAppointments(
	ctx context.Context,
	doctorId uuid.UUID,
	params api.GetDoctorAppointmentsParams,
) ([]Appointment, error) {
	return a.getAppointments(ctx, "doctorId", doctorId, params.Date, params.From, params.To)
}

func (a *AppointmentApp) GetPatientAppointments(
	ctx context.Context,
	patientId uuid.UUID,
	params api.GetPatientAppointmentsParams,
) ([]Appointment, error) {
	// Patient view doesn't support single date filter in API spec, only From/To
	return a.getAppointments(ctx, "patientId", patientId, nil, params.From, params.To)
}
