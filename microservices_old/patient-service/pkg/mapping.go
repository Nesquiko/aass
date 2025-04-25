package pkg

import (
	"github.com/oapi-codegen/runtime/types"

	"github.com/Nesquiko/aass/patient-service/api"
)

func patientRegToDataPatient(p api.NewPatientRequest) Patient {
	return Patient{
		Email:     string(p.Email),
		FirstName: p.FirstName,
		LastName:  p.LastName,
	}
}

func dataPatientToApiPatient(p Patient) api.Patient {
	return api.Patient{
		Id:        &p.Id,
		Email:     types.Email(p.Email),
		FirstName: p.FirstName,
		LastName:  p.LastName,
		Role:      asPtr(api.UserRolePatient),
	}
}

func asPtr[T any](v T) *T {
	return &v
}
