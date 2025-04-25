package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Nesquiko/wac/pkg/api"
	"github.com/Nesquiko/wac/pkg/data"
)

func (a MonolithApp) CreatePatient(
	ctx context.Context,
	p api.PatientRegistration,
) (api.Patient, error) {
	patient := patientRegToDataPatient(p)

	patient, err := a.db.CreatePatient(ctx, patient)
	if errors.Is(err, data.ErrDuplicateEmail) {
		return api.Patient{}, fmt.Errorf("CreatePatient duplicate emall: %w", ErrDuplicateEmail)
	} else if err != nil {
		return api.Patient{}, fmt.Errorf("CreatePatient: %w", err)
	}

	return dataPatientToApiPatient(patient), nil
}

func (a MonolithApp) PatientById(ctx context.Context, id uuid.UUID) (api.Patient, error) {
	patient, err := a.db.PatientById(ctx, id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return api.Patient{}, fmt.Errorf("PatientById: %w", ErrNotFound)
		}
		return api.Patient{}, fmt.Errorf("PatientById: %w", err)
	}

	return dataPatientToApiPatient(patient), nil
}

func (a MonolithApp) PatientByEmail(ctx context.Context, email string) (api.Patient, error) {
	patient, err := a.db.PatientByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return api.Patient{}, fmt.Errorf("PatientByEmail: %w", ErrNotFound)
		}
		return api.Patient{}, fmt.Errorf("PatientByEmail: %w", err)
	}

	return dataPatientToApiPatient(patient), nil
}

func (a MonolithApp) PatientsCalendar(
	ctx context.Context,
	patientId uuid.UUID,
	from api.From,
	to *api.To,
) (api.Appointments, error) {
	var toTime *time.Time = nil
	if to != nil {
		toTime = &to.Time
	}

	appts, err := a.db.AppointmentsByPatientId(ctx, patientId, from.Time, toTime)
	if err != nil {
		return api.Appointments{}, fmt.Errorf("PatientsCalendar appointments: %w", err)
	}

	calendar := api.Appointments{}
	var doctor *data.Doctor = nil
	if len(appts) != 0 {
		calendar.Appointments = asPtr(make([]api.AppointmentDisplay, len(appts)))

		d, err := a.db.DoctorById(ctx, appts[0].DoctorId)
		if err != nil {
			return api.Appointments{}, fmt.Errorf("PatientsCalendar doc find: %w", err)
		}
		doctor = &d
	}

	for i, appt := range appts {
		patient, err := a.db.PatientById(ctx, appt.PatientId)
		if err != nil {
			return api.Appointments{}, fmt.Errorf("PatientsCalendar patient find: %w", err)
		}
		(*calendar.Appointments)[i] = dataApptToApptDisplay(appt, patient, *doctor)
	}

	return calendar, nil
}
