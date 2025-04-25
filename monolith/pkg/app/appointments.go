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

// CreateAppointment implements App.
func (a MonolithApp) CreateAppointment(
	ctx context.Context,
	appt api.NewAppointmentRequest,
) (api.Appointment, error) {
	appointment, err := a.db.CreateAppointment(ctx, newApptToDataAppt(appt))
	if err != nil {
		return api.Appointment{}, fmt.Errorf("CreateAppointment create appointment: %w", err)
	}

	doc, err := a.db.DoctorById(ctx, appointment.DoctorId)
	if err != nil {
		return api.Appointment{}, fmt.Errorf("CreateAppointment find doctor: %w", err)
	}

	var cond *data.Condition = nil
	if appointment.ConditionId != nil {
		c, err := a.db.ConditionById(ctx, *appointment.ConditionId)
		if err != nil {
			return api.Appointment{}, fmt.Errorf("CreateAppointment find condition: %w", err)
		}
		cond = &c
	}

	prescriptions, err := a.db.PrescriptionByAppointmentId(ctx, appointment.Id)
	if err != nil {
		return api.Appointment{}, fmt.Errorf(
			"CreateAppointment fetch prescriptions: %w",
			err,
		)
	}

	return dataApptToPatientAppt(appointment, doc, cond, prescriptions), nil
}

func (a MonolithApp) CancelAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
	req api.AppointmentCancellation,
) error {
	err := a.db.CancelAppointment(ctx, appointmentId, string(req.By), req.Reason)
	if err != nil {
		return fmt.Errorf("CancelAppointment: %w", err)
	}
	return nil
}

func (a MonolithApp) AppointmentById(
	ctx context.Context,
	appointmentId uuid.UUID,
) (api.Appointment, error) {
	appointment, err := a.db.AppointmentById(ctx, appointmentId)
	if err != nil {
		return api.Appointment{}, fmt.Errorf("DoctorsAppointmentById: %w", err)
	}

	patient, err := a.db.PatientById(ctx, appointment.PatientId)
	if err != nil {
		return api.Appointment{}, fmt.Errorf("DoctorsAppointmentById find patient: %w", err)
	}

	var cond *data.Condition
	if appointment.ConditionId != nil {
		c, err := a.db.ConditionById(ctx, *appointment.ConditionId)
		if err != nil {
			return api.Appointment{}, fmt.Errorf(
				"DoctorsAppointmentById find condition: %w",
				err,
			)
		}
		cond = &c
	}

	resources, err := a.db.ResourcesByAppointmentId(ctx, appointmentId)
	if err != nil {
		return api.Appointment{}, fmt.Errorf(
			"DoctorsAppointmentById fetch reservations: %w",
			err,
		)
	}

	var facilities, equipment, medicine []data.Resource
	for _, resource := range resources {
		switch resource.Type {
		case data.ResourceTypeFacility:
			facilities = append(facilities, resource)
		case data.ResourceTypeEquipment:
			equipment = append(equipment, resource)
		case data.ResourceTypeMedicine:
			medicine = append(medicine, resource)
		}
	}

	prescriptions, err := a.db.PrescriptionByAppointmentId(ctx, appointmentId)
	if err != nil {
		return api.Appointment{}, fmt.Errorf(
			"DoctorsAppointmentById fetch prescriptions: %w",
			err,
		)
	}

	return dataApptToDoctorAppt(
		appointment,
		patient,
		cond,
		facilities,
		equipment,
		medicine,
		prescriptions,
	), nil
}

func (a MonolithApp) DecideAppointment(
	ctx context.Context,
	appointmentId uuid.UUID,
	decision api.AppointmentDecision,
) (api.Appointment, error) {
	var resources []data.Resource
	if decision.Facility != nil {
		resources = append(resources, data.Resource{
			Id:   *decision.Facility,
			Type: data.ResourceTypeFacility,
		})
	}
	if decision.Equipment != nil {
		resources = append(resources, data.Resource{
			Id:   *decision.Equipment,
			Type: data.ResourceTypeEquipment,
		})
	}
	if decision.Medicine != nil {
		resources = append(resources, data.Resource{
			Id:   *decision.Medicine,
			Type: data.ResourceTypeMedicine,
		})
	}

	appointment, err := a.db.DecideAppointment(
		ctx,
		appointmentId,
		string(decision.Action),
		decision.Reason,
		resources,
	)
	if err != nil {
		if errors.Is(err, data.ErrResourceUnavailable) {
			return api.Appointment{}, fmt.Errorf(
				"DecideAppointment: %w",
				ErrResourceUnavailable,
			)
		}
		return api.Appointment{}, fmt.Errorf("DecideAppointment: %w", err)
	}

	patient, err := a.db.PatientById(ctx, appointment.PatientId)
	if err != nil {
		return api.Appointment{}, fmt.Errorf("DecideAppointment fetch patient: %w", err)
	}

	var cond *data.Condition
	if appointment.ConditionId != nil {
		c, err := a.db.ConditionById(ctx, *appointment.ConditionId)
		if err != nil {
			return api.Appointment{}, fmt.Errorf("DecideAppointment fetch condition: %w", err)
		}
		cond = &c
	}

	var facilities, equipment, medicine []data.Resource
	for _, resource := range resources {
		switch resource.Type {
		case data.ResourceTypeFacility:
			facilities = append(facilities, resource)
		case data.ResourceTypeEquipment:
			equipment = append(equipment, resource)
		case data.ResourceTypeMedicine:
			medicine = append(medicine, resource)
		}
	}

	prescriptions, err := a.db.PrescriptionByAppointmentId(ctx, appointmentId)
	if err != nil {
		return api.Appointment{}, fmt.Errorf(
			"DecideAppointment fetch prescriptions: %w",
			err,
		)
	}

	doctorAppointment := dataApptToDoctorAppt(
		appointment,
		patient,
		cond,
		facilities,
		equipment,
		medicine,
		prescriptions,
	)

	return doctorAppointment, nil
}

func (a MonolithApp) DoctorTimeSlots(
	ctx context.Context,
	doctorId uuid.UUID,
	date time.Time,
) (api.DoctorTimeslots, error) {
	appointments, err := a.db.AppointmentsByDoctorIdAndDate(ctx, doctorId, date)
	if err != nil {
		return api.DoctorTimeslots{}, fmt.Errorf("DoctorTimeSlots: %w", err)
	}

	bookedHours := make(map[int]bool)
	for _, appt := range appointments {
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

func (a MonolithApp) RescheduleAppointment(
	ctx context.Context,
	appointmentId api.AppointmentId,
	newDateTime time.Time,
) (api.Appointment, error) {
	appt, err := a.db.RescheduleAppointment(ctx, appointmentId, newDateTime)
	if err != nil {
		if errors.Is(err, data.ErrDoctorUnavailable) {
			return api.Appointment{}, fmt.Errorf(
				"RescheduleAppointment: %w",
				ErrDoctorUnavailable,
			)
		}
		return api.Appointment{}, fmt.Errorf("RescheduleAppointment: %w", err)
	}

	doc, err := a.db.DoctorById(ctx, appt.DoctorId)
	if err != nil {
		return api.Appointment{}, fmt.Errorf("RescheduleAppointment find doctor: %w", err)
	}

	var cond *data.Condition
	if appt.ConditionId != nil {
		c, err := a.db.ConditionById(ctx, *appt.ConditionId)
		if err != nil {
			return api.Appointment{}, fmt.Errorf(
				"RescheduleAppointment find condition: %w",
				err,
			)
		}
		cond = &c
	}

	prescriptions, err := a.db.PrescriptionByAppointmentId(ctx, appt.Id)
	if err != nil {
		return api.Appointment{}, fmt.Errorf(
			"RescheduleAppointment fetch prescriptions: %w",
			err,
		)
	}

	return dataApptToPatientAppt(appt, doc, cond, prescriptions), nil
}
