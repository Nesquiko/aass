package pkg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/Nesquiko/aass/prescription-service/api"
)

type PrescriptionApp struct {
	db *MongoPrescriptionDb
}

func (a PrescriptionApp) CreatePatientPrescription(
	ctx context.Context,
	pres api.NewPrescription,
) (api.Prescription, error) {
	prescription, err := a.db.CreatePrescription(ctx, newPrescToDataPresc(pres))
	if err != nil {
		return api.Prescription{}, fmt.Errorf("CreatePatientPrescription: %w", err)
	}

	return dataPrescToPresc(prescription), nil
}

func (a PrescriptionApp) PrescriptionById(
	ctx context.Context,
	prescriptionId uuid.UUID,
) (api.Prescription, error) {
	prescription, err := a.db.PrescriptionById(ctx, prescriptionId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return api.Prescription{}, fmt.Errorf("PrescriptionById: %w", ErrNotFound)
		}
		return api.Prescription{}, fmt.Errorf("PrescriptionById fetch prescription failed: %w", err)
	}

	return dataPrescToPresc(prescription), nil
}

func (a PrescriptionApp) UpdatePatientPrescription(
	ctx context.Context,
	prescriptionId uuid.UUID,
	updateData api.UpdatePrescription,
) (api.Prescription, error) {
	existingPrescription, err := a.db.PrescriptionById(ctx, prescriptionId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return api.Prescription{}, fmt.Errorf("UpdatePatientPrescription: %w", ErrNotFound)
		}
		return api.Prescription{}, fmt.Errorf("UpdatePatientPrescription fetch failed: %w", err)
	}

	updated := false
	if updateData.DoctorsNote != nil {
		if existingPrescription.DoctorsNote == nil ||
			*existingPrescription.DoctorsNote != *updateData.DoctorsNote {
			existingPrescription.DoctorsNote = updateData.DoctorsNote
			updated = true
		}
	}
	if updateData.End != nil {
		if !existingPrescription.End.Equal(*updateData.End) {
			existingPrescription.End = *updateData.End
			updated = true
		}
	}
	if updateData.Name != nil {
		if existingPrescription.Name != *updateData.Name {
			existingPrescription.Name = *updateData.Name
			updated = true
		}
	}
	if updateData.Start != nil {
		if !existingPrescription.Start.Equal(*updateData.Start) {
			existingPrescription.Start = *updateData.Start
			updated = true
		}
	}

	var updatedDbPrescription Prescription
	if updated {
		updatedDbPrescription, err = a.db.UpdatePrescription(
			ctx,
			prescriptionId,
			existingPrescription,
		)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return api.Prescription{}, fmt.Errorf(
					"UpdatePatientPrescription update target vanished: %w",
					ErrNotFound,
				)
			}
			return api.Prescription{}, fmt.Errorf(
				"UpdatePatientPrescription update failed: %w",
				err,
			)
		}
	} else {
		updatedDbPrescription = existingPrescription
	}

	return dataPrescToPresc(updatedDbPrescription), nil
}

func (a PrescriptionApp) GetPrescriptionsByPatientAndRange(
	ctx context.Context,
	patientId uuid.UUID,
	from time.Time,
	to time.Time,
) ([]api.PrescriptionDisplay, error) {
	dbPrescriptions, err := a.db.FindPrescriptionsByPatientIdAndRange(ctx, patientId, from, to)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Failed to get prescriptions from DB",
			"patientId",
			patientId,
			"error",
			err,
		)
		return nil, fmt.Errorf("failed to retrieve prescriptions: %w", err)
	}

	apiPrescriptions := dataPrescSliceToPrescDisplaySlice(dbPrescriptions)
	return apiPrescriptions, nil
}
