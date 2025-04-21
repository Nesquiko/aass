package pkg

import (
	"github.com/oapi-codegen/runtime/types"

	"github.com/Nesquiko/aass/auth-service/api"
	doctorapi "github.com/Nesquiko/aass/auth-service/api/doctor-api"
	patientapi "github.com/Nesquiko/aass/auth-service/api/patient-api"
)

func doctorRegToApiDoctor(d doctorapi.Doctor) api.Doctor {
	return api.Doctor{
		Id:             *d.Id,
		Email:          types.Email(d.Email),
		FirstName:      d.FirstName,
		LastName:       d.LastName,
		Specialization: api.SpecializationEnum(d.Specialization),
		Role:           api.UserRoleDoctor,
	}
}

func patientRegToApiPatient(p patientapi.Patient) api.Patient {
	return api.Patient{
		Id:        *p.Id,
		Email:     types.Email(p.Email),
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Role:      api.UserRolePatient,
	}
}

func doctorToApiDoctor(d doctorapi.Doctor) api.Doctor {
	return api.Doctor{
		Id:             *d.Id,
		Email:          types.Email(d.Email),
		FirstName:      d.FirstName,
		LastName:       d.LastName,
		Specialization: api.SpecializationEnum(d.Specialization),
		Role:           api.UserRoleDoctor,
	}
}

func patientToApiPatient(p patientapi.Patient) api.Patient {
	return api.Patient{
		Id:        *p.Id,
		Email:     types.Email(p.Email),
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Role:      api.UserRolePatient,
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
