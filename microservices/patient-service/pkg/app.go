package pkg

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/Nesquiko/aass/patient-service/api"
)

type PatientApp struct{ db *MongoPatientDb }

func (a PatientApp) CreatePatient(
	ctx context.Context,
	p api.NewPatientRequest,
) (api.Patient, error) {
	patient := patientRegToDataPatient(p)

	patient, err := a.db.CreatePatient(ctx, patient)
	if errors.Is(err, ErrDuplicateEmail) {
		return api.Patient{}, fmt.Errorf("CreatePatient duplicate emall: %w", ErrDuplicateEmail)
	} else if err != nil {
		return api.Patient{}, fmt.Errorf("CreatePatient: %w", err)
	}

	return dataPatientToApiPatient(patient), nil
}

func (a PatientApp) PatientById(ctx context.Context, id uuid.UUID) (api.Patient, error) {
	patient, err := a.db.FindPatientById(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return api.Patient{}, fmt.Errorf("PatientById: %w", ErrNotFound)
		}
		return api.Patient{}, fmt.Errorf("PatientById: %w", err)
	}

	return dataPatientToApiPatient(patient), nil
}

func (a PatientApp) PatientByEmail(ctx context.Context, email string) (api.Patient, error) {
	patient, err := a.db.FindPatientByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return api.Patient{}, fmt.Errorf("PatientByEmail: %w", ErrNotFound)
		}
		return api.Patient{}, fmt.Errorf("PatientByEmail: %w", err)
	}

	return dataPatientToApiPatient(patient), nil
}
