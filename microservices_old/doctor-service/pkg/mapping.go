package pkg

import (
	"fmt"

	"github.com/oapi-codegen/runtime/types"

	"github.com/Nesquiko/aass/doctor-service/api"
	apptapi "github.com/Nesquiko/aass/doctor-service/api/appt-api"
)

func doctorRegToDataDoctor(d api.NewDoctorRequest) Doctor {
	doctor := Doctor{
		Email:          string(d.Email),
		FirstName:      d.FirstName,
		LastName:       d.LastName,
		Specialization: string(d.Specialization),
	}

	return doctor
}

func dataDoctorToApiDoctor(d Doctor) api.Doctor {
	return api.Doctor{
		Id:             &d.Id,
		Email:          types.Email(d.Email),
		FirstName:      d.FirstName,
		LastName:       d.LastName,
		Specialization: api.SpecializationEnum(d.Specialization),
		Role:           asPtr(api.UserRoleDoctor),
	}
}

func docApptToApptDisplay(appt apptapi.DoctorAppointment, doctor Doctor) api.AppointmentDisplay {
	return api.AppointmentDisplay{
		Id:                  *appt.Id,
		AppointmentDateTime: appt.AppointmentDateTime,
		DoctorName:          fmt.Sprintf("%s %s", doctor.FirstName, doctor.LastName),
		PatientName:         fmt.Sprintf("%s %s", appt.Patient.FirstName, appt.Patient.LastName),
		Status:              api.AppointmentStatus(appt.Status),
		Type:                api.AppointmentType(appt.Type),
	}
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
