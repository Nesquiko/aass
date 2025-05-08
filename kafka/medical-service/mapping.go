package main

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/Nesquiko/aass/common/server"
	"github.com/Nesquiko/aass/medical-service/api"
	appointmentapi "github.com/Nesquiko/aass/medical-service/appointment-api"
)

func dataCondToCond(c Condition, appts []appointmentapi.AppointmentDisplay) api.Condition {
	appointments := make([]api.AppointmentDisplay, len(appts))
	for i, appt := range appts {
		appointments[i] = api.AppointmentDisplay{
			Id:                  appt.Id,
			AppointmentDateTime: appt.AppointmentDateTime,
			DoctorName:          appt.DoctorName,
			PatientName:         appt.PatientName,
			Status:              api.AppointmentStatus(appt.Status),
			Type:                api.AppointmentType(appt.Type),
		}
	}

	return api.Condition{
		Id:           &c.Id,
		Name:         c.Name,
		Start:        c.Start,
		End:          c.End,
		Appointments: appointments,
		AppointmentsIds: server.AsPtr(
			server.Map(
				appointments,
				func(appt api.AppointmentDisplay) uuid.UUID { return appt.Id },
			),
		),
	}
}

func dataCondToCondDisplay(c Condition) api.ConditionDisplay {
	return api.ConditionDisplay{
		Id:    &c.Id,
		Name:  c.Name,
		Start: c.Start,
		End:   c.End,
	}
}

func newCondToDataCond(c api.NewCondition) Condition {
	return Condition{
		PatientId: c.PatientId,
		Name:      c.Name,
		Start:     c.Start,
		End:       c.End,
	}
}

func newPrescToDataPresc(p api.NewPrescription) Prescription {
	return Prescription{
		PatientId:     p.PatientId,
		Name:          p.Name,
		Start:         p.Start,
		End:           p.End,
		DoctorsNote:   p.DoctorsNote,
		AppointmentId: p.AppointmentId,
	}
}

func dataPrescToPresc(
	p Prescription,
	appt *appointmentapi.Appointment,
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
		presc.Appointment = server.AsPtr(apptToApptDisplay(*appt))
	}
	return presc
}

func dataPrescToPrescDisplay(p Prescription) api.PrescriptionDisplay {
	return api.PrescriptionDisplay{
		Id:            &p.Id,
		AppointmentId: p.AppointmentId,
		End:           p.End,
		Name:          p.Name,
		Start:         p.Start,
	}
}

func apptToApptDisplay(appt appointmentapi.Appointment) api.AppointmentDisplay {
	return api.AppointmentDisplay{
		Id:                  appt.Id,
		AppointmentDateTime: appt.AppointmentDateTime,
		DoctorName:          fmt.Sprintf("%s %s", appt.Doctor.FirstName, appt.Doctor.LastName),
		PatientName:         fmt.Sprintf("%s %s", appt.Patient.FirstName, appt.Patient.LastName),
		Status:              api.AppointmentStatus(appt.Status),
		Type:                api.AppointmentType(appt.Type),
	}
}
