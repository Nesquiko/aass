package pkg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"

	"github.com/Nesquiko/aass/doctor-service/api"
	apptapi "github.com/Nesquiko/aass/doctor-service/api/appt-api"
)

type DoctorApp struct {
	db         *MongoDoctorDb
	apptClient *apptapi.ClientWithResponses
}

func (a DoctorApp) CreateDoctor(ctx context.Context, p api.NewDoctorRequest) (api.Doctor, error) {
	doctor := doctorRegToDataDoctor(p)

	doctor, err := a.db.CreateDoctor(ctx, doctor)
	if errors.Is(err, ErrDuplicateEmail) {
		return api.Doctor{}, fmt.Errorf("CreateDoctor duplicate emall: %w", ErrDuplicateEmail)
	} else if err != nil {
		return api.Doctor{}, fmt.Errorf("CreateDoctor: %w", err)
	}

	return dataDoctorToApiDoctor(doctor), nil
}

func (a DoctorApp) DoctorById(ctx context.Context, id uuid.UUID) (api.Doctor, error) {
	doctor, err := a.db.DoctorById(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return api.Doctor{}, fmt.Errorf("DoctorById: %w", ErrNotFound)
		}
		return api.Doctor{}, fmt.Errorf("DoctorById: %w", err)
	}

	return dataDoctorToApiDoctor(doctor), nil
}

func (a DoctorApp) DoctorByEmail(ctx context.Context, email string) (api.Doctor, error) {
	doctor, err := a.db.DoctorByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return api.Doctor{}, fmt.Errorf("Doctor not found: %w", ErrNotFound)
		}
		return api.Doctor{}, fmt.Errorf("Doctor: %w", err)
	}

	return dataDoctorToApiDoctor(doctor), nil
}

func (a DoctorApp) GetAllDoctors(ctx context.Context) ([]api.Doctor, error) {
	allDataDoctors, err := a.db.GetAllDoctors(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetAllDoctors failed: %w", err)
	}

	allApiDoctors := Map(allDataDoctors, dataDoctorToApiDoctor)
	return allApiDoctors, nil
}

func (a DoctorApp) DoctorTimeSlots(
	ctx context.Context,
	doctorId uuid.UUID,
	date time.Time,
) (api.DoctorTimeslots, error) {
	appointments, err := a.apptClient.GetDoctorAppointmentsWithResponse(
		ctx,
		doctorId,
		&apptapi.GetDoctorAppointmentsParams{Date: &types.Date{Time: date}},
	)
	if err != nil {
		return api.DoctorTimeslots{}, fmt.Errorf("DoctorTimeSlots: %w", err)
	}

	bookedHours := make(map[int]bool)
	for _, appt := range appointments.JSON200.Appointments {
		bookedHours[appt.AppointmentDateTime.Hour()] = true
	}

	var slots []api.TimeSlot
	for hour := 8; hour <= 14; hour++ {
		status := api.Available
		if bookedHours[hour] {
			status = api.Unavailable
		}

		slotTime := time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, date.Location()).
			Format("15:04")

		slots = append(slots, api.TimeSlot{
			Status: status,
			Time:   slotTime,
		})
	}

	return api.DoctorTimeslots{Slots: slots}, nil
}

func (a DoctorApp) DoctorsCalendar(
	ctx context.Context,
	doctorId api.DoctorId,
	from api.From,
	to *api.To,
) (api.DoctorCalendar, error) {
	appts, err := a.apptClient.GetDoctorAppointmentsWithResponse(
		ctx,
		doctorId,
		&apptapi.GetDoctorAppointmentsParams{From: &from, To: to},
	)
	if err != nil {
		return api.DoctorCalendar{}, fmt.Errorf("DoctorCalendar: %w", err)
	}

	calendar := api.DoctorCalendar{
		Appointments: asPtr(make([]api.AppointmentDisplay, len(appts.JSON200.Appointments))),
	}
	doctor, err := a.db.DoctorById(ctx, doctorId)
	if err != nil {
		return api.DoctorCalendar{}, fmt.Errorf("DoctorCalendar doc find: %w", err)
	}

	for i, appt := range appts.JSON200.Appointments {
		(*calendar.Appointments)[i] = docApptToApptDisplay(appt, doctor)
	}

	return calendar, nil
}
