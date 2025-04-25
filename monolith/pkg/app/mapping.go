package app

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/runtime/types"

	"github.com/Nesquiko/wac/pkg/api"
	"github.com/Nesquiko/wac/pkg/data"
)

func patientRegToDataPatient(p api.PatientRegistration) data.Patient {
	return data.Patient{
		Email:     string(p.Email),
		FirstName: p.FirstName,
		LastName:  p.LastName,
	}
}

func dataPatientToApiPatient(p data.Patient) api.Patient {
	return api.Patient{
		Id:        p.Id,
		Email:     types.Email(p.Email),
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Role:      api.UserRolePatient,
	}
}

func doctorRegToDataDoctor(d api.DoctorRegistration) data.Doctor {
	doctor := data.Doctor{
		Email:          string(d.Email),
		FirstName:      d.FirstName,
		LastName:       d.LastName,
		Specialization: string(d.Specialization),
	}

	return doctor
}

func dataDoctorToApiDoctor(d data.Doctor) api.Doctor {
	return api.Doctor{
		Id:             d.Id,
		Email:          types.Email(d.Email),
		FirstName:      d.FirstName,
		LastName:       d.LastName,
		Specialization: api.SpecializationEnum(d.Specialization),
		Role:           api.UserRoleDoctor,
	}
}

func newCondToDataCond(c api.NewCondition) data.Condition {
	return data.Condition{
		PatientId: c.PatientId,
		Name:      c.Name,
		Start:     c.Start,
		End:       c.End,
	}
}

func dataCondToCondDisplay(c data.Condition) api.ConditionDisplay {
	return api.ConditionDisplay{
		Id:    &c.Id,
		Name:  c.Name,
		Start: c.Start,
		End:   c.End,
	}
}

func dataCondToCond(c data.Condition, appts []api.AppointmentDisplay) api.Condition {
	return api.Condition{
		Id:           &c.Id,
		Name:         c.Name,
		Start:        c.Start,
		End:          c.End,
		Appointments: appts,
		AppointmentsIds: asPtr(
			Map(appts, func(appt api.AppointmentDisplay) uuid.UUID { return appt.Id }),
		),
	}
}

func newPrescToDataPresc(p api.NewPrescription) data.Prescription {
	return data.Prescription{
		PatientId:     p.PatientId,
		Name:          p.Name,
		Start:         p.Start,
		End:           p.End,
		DoctorsNote:   p.DoctorsNote,
		AppointmentId: p.AppointmentId,
	}
}

func dataPrescToPresc(
	p data.Prescription,
	appt *data.Appointment,
	patient *data.Patient,
	doctor *data.Doctor,
) api.Prescription {
	presc := api.Prescription{
		Id:            &p.Id,
		Name:          p.Name,
		Start:         p.Start,
		End:           p.End,
		DoctorsNote:   p.DoctorsNote,
		AppointmentId: p.AppointmentId,
	}
	if appt != nil {
		presc.AppointmentId = &appt.Id
		presc.Appointment = asPtr(dataApptToApptDisplay(*appt, *patient, *doctor))
	}
	return presc
}

func dataPrescToPrescDisplay(p data.Prescription) api.PrescriptionDisplay {
	return api.PrescriptionDisplay{
		Id:            &p.Id,
		AppointmentId: p.AppointmentId,
		End:           p.End,
		Name:          p.Name,
		Start:         p.Start,
	}
}

func resourceToEquipment(r data.Resource) api.Equipment {
	return api.Equipment{Id: r.Id, Name: r.Name}
}

func resourceToFacility(r data.Resource) api.Facility {
	return api.Facility{Id: r.Id, Name: r.Name}
}

func resourceToMedicine(r data.Resource) api.Medicine {
	return api.Medicine{Id: r.Id, Name: r.Name}
}

func newApptToDataAppt(a api.NewAppointmentRequest) data.Appointment {
	appt := data.Appointment{
		PatientId:           a.PatientId,
		DoctorId:            a.DoctorId,
		AppointmentDateTime: a.AppointmentDateTime,
		EndTime:             a.AppointmentDateTime.Add(time.Hour),
		Status:              string(api.Requested),
		Reason:              a.Reason,
		ConditionId:         a.ConditionId,
	}

	if a.Type != nil {
		appt.Type = string(*a.Type)
	}
	return appt
}

func dataApptToApptDisplay(
	appt data.Appointment,
	patient data.Patient,
	doctor data.Doctor,
) api.AppointmentDisplay {
	return api.AppointmentDisplay{
		Id:                  appt.Id,
		AppointmentDateTime: appt.AppointmentDateTime,
		DoctorName:          fmt.Sprintf("%s %s", doctor.FirstName, doctor.LastName),
		PatientName:         fmt.Sprintf("%s %s", patient.FirstName, patient.LastName),
		Status:              api.AppointmentStatus(appt.Status),
		Type:                api.AppointmentType(appt.Type),
	}
}

func dataApptToPatientAppt(
	a data.Appointment,
	d data.Doctor,
	c *data.Condition,
	prescriptions []data.Prescription,
) api.Appointment {
	appt := api.Appointment{
		Id:                  a.Id,
		AppointmentDateTime: a.AppointmentDateTime,
		Doctor:              dataDoctorToApiDoctor(d),
		Reason:              a.Reason,
		Status:              api.AppointmentStatus(a.Status),
		Type:                api.AppointmentType(a.Type),
		CancellationReason:  a.CancellationReason,
		CanceledBy:          (*api.UserRole)(a.CancelledBy),
		DenialReason:        a.DenialReason,
	}

	if c != nil {
		cond := dataCondToCondDisplay(*c)
		appt.Condition = &cond
	}

	appt.Prescriptions = asPtr(Map(prescriptions, dataPrescToPrescDisplay))

	return appt
}

func dataApptToDoctorAppt(
	appt data.Appointment,
	patient data.Patient,
	cond *data.Condition,
	facilities []data.Resource,
	equipment []data.Resource,
	medicine []data.Resource,
	prescriptions []data.Prescription,
) api.Appointment {
	doctorAppt := api.Appointment{
		Id:                  appt.Id,
		AppointmentDateTime: appt.AppointmentDateTime,
		CancellationReason:  appt.CancellationReason,
		CanceledBy:          (*api.UserRole)(appt.CancelledBy),
		Patient:             dataPatientToApiPatient(patient),
		Reason:              appt.Reason,
		Status:              api.AppointmentStatus(appt.Status),
		Type:                api.AppointmentType(appt.Type),
		DenialReason:        appt.DenialReason,
	}

	if cond != nil {
		doctorAppt.Condition = asPtr(dataCondToCondDisplay(*cond))
	}

	if len(facilities) > 0 {
		doctorAppt.Facilities = asPtr(Map(facilities, resourceToFacility))
	}
	if len(equipment) > 0 {
		doctorAppt.Equipment = asPtr(Map(equipment, resourceToEquipment))
	}
	if len(medicine) > 0 {
		doctorAppt.Medicine = asPtr(Map(medicine, resourceToMedicine))
	}

	doctorAppt.Prescriptions = asPtr(Map(prescriptions, dataPrescToPrescDisplay))

	return doctorAppt
}

func asPtr[T any](v T) *T {
	return &v
}

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}
