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

func (a MonolithApp) CreateDoctor(
	ctx context.Context,
	d api.DoctorRegistration,
) (api.Doctor, error) {
	doctor := doctorRegToDataDoctor(d)
	doctor, err := a.db.CreateDoctor(ctx, doctor)
	if errors.Is(err, data.ErrDuplicateEmail) {
		return api.Doctor{}, fmt.Errorf("CreateDoctor duplicate emall: %w", ErrDuplicateEmail)
	} else if err != nil {
		return api.Doctor{}, fmt.Errorf("CreateDoctor: %w", err)
	}

	return dataDoctorToApiDoctor(doctor), nil
}

func (a MonolithApp) DoctorById(ctx context.Context, id uuid.UUID) (api.Doctor, error) {
	doctor, err := a.db.DoctorById(ctx, id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return api.Doctor{}, fmt.Errorf("DoctorById: %w", ErrNotFound)
		}
		return api.Doctor{}, fmt.Errorf("DoctorById: %w", err)
	}

	return dataDoctorToApiDoctor(doctor), nil
}

func (a MonolithApp) DoctorByEmail(ctx context.Context, email string) (api.Doctor, error) {
	doctor, err := a.db.DoctorByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return api.Doctor{}, fmt.Errorf("DoctorByEmail: %w", ErrNotFound)
		}
		return api.Doctor{}, fmt.Errorf("DoctorByEmail: %w", err)
	}

	return dataDoctorToApiDoctor(doctor), nil
}

func (a MonolithApp) DoctorsCalendar(
	ctx context.Context,
	doctorId api.DoctorId,
	from api.From,
	to *api.To,
) (api.Appointments, error) {
	var toTime *time.Time = nil
	if to != nil {
		toTime = &to.Time
	}

	appts, err := a.db.AppointmentsByDoctorId(ctx, doctorId, from.Time, toTime)
	if err != nil {
		return api.Appointments{}, fmt.Errorf("DoctorCalendar: %w", err)
	}

	doctor, err := a.db.DoctorById(ctx, doctorId)
	if err != nil {
		return api.Appointments{}, fmt.Errorf("DoctorCalendar doc find: %w", err)
	}

	calendar := api.Appointments{
		Appointments: asPtr(make([]api.AppointmentDisplay, len(appts))),
	}

	for i, appt := range appts {
		patient, err := a.db.PatientById(ctx, appt.PatientId)
		if err != nil {
			return api.Appointments{}, fmt.Errorf("DoctorCalendar patient find: %w", err)
		}
		(*calendar.Appointments)[i] = dataApptToApptDisplay(appt, patient, doctor)

	}

	return calendar, nil
}

func (a MonolithApp) AvailableDoctors(
	ctx context.Context,
	dateTime time.Time,
) ([]api.Doctor, error) {
	availableDataDoctors, err := a.db.AvailableDoctors(ctx, dateTime)
	if err != nil {
		return nil, fmt.Errorf("AvailableDoctors failed: %w", err)
	}

	availableApiDoctors := Map(availableDataDoctors, dataDoctorToApiDoctor)
	return availableApiDoctors, nil
}

func (a MonolithApp) GetAllDoctors(ctx context.Context) ([]api.Doctor, error) {
	allDataDoctors, err := a.db.GetAllDoctors(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllDoctors failed: %w", err)
	}

	allApiDoctors := Map(allDataDoctors, dataDoctorToApiDoctor)
	return allApiDoctors, nil
}
