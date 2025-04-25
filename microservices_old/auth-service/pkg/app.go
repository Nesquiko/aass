package pkg

import (
	"context"
	"fmt"
	"net/http"

	"github.com/oapi-codegen/runtime/types"

	"github.com/Nesquiko/aass/auth-service/api"
	doctorapi "github.com/Nesquiko/aass/auth-service/api/doctor-api"
	patientapi "github.com/Nesquiko/aass/auth-service/api/patient-api"
)

type AuthApp struct {
	patientClient *patientapi.ClientWithResponses
	doctorClient  *doctorapi.ClientWithResponses
}

func (a AuthApp) CreateDoctor(ctx context.Context, d api.DoctorRegistration) (api.Doctor, error) {
	doc, err := a.doctorClient.CreateDoctorWithResponse(ctx, doctorapi.NewDoctorRequest{
		Email:          d.Email,
		FirstName:      d.FirstName,
		LastName:       d.LastName,
		Specialization: doctorapi.SpecializationEnum(d.Specialization),
	})
	if err != nil {
		return api.Doctor{}, fmt.Errorf("CreateDoctor: %w", err)
	}

	return doctorRegToApiDoctor(*doc.JSON201), nil
}

func (a AuthApp) CreatePatient(
	ctx context.Context,
	p api.PatientRegistration,
) (api.Patient, error) {
	pat, err := a.patientClient.CreatePatientWithResponse(
		ctx,
		patientapi.CreatePatientJSONRequestBody{
			Email:     p.Email,
			FirstName: p.FirstName,
			LastName:  p.LastName,
		},
	)
	if err != nil {
		return api.Patient{}, fmt.Errorf("CreatePatient: %w", err)
	}

	return patientRegToApiPatient(*pat.JSON201), nil
}

func (a AuthApp) DoctorByEmail(ctx context.Context, email string) (api.Doctor, error) {
	doctor, err := a.doctorClient.GetDoctorByEmailWithResponse(ctx, doctorapi.DoctorEmail(email))
	if err != nil || doctor.StatusCode() != http.StatusOK {
		return api.Doctor{}, fmt.Errorf("DoctorByEmail: %w", err)
	}

	return doctorToApiDoctor(*doctor.JSON200), nil
}

func (a AuthApp) PatientByEmail(ctx context.Context, email string) (api.Patient, error) {
	patient, err := a.patientClient.GetPatientByEmailWithResponse(
		ctx,
		&patientapi.GetPatientByEmailParams{Email: types.Email(email)},
	)
	if err != nil || patient.StatusCode() != http.StatusOK {
		return api.Patient{}, fmt.Errorf("PatientByEmail: %w", err)
	}

	return patientToApiPatient(*patient.JSON200), nil
}
