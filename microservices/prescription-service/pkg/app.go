package pkg

import (
	"context"
	"errors"
	"fmt"

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
	if updateData.AppointmentId != nil {
		if existingPrescription.AppointmentId == nil ||
			*existingPrescription.AppointmentId != *updateData.AppointmentId {
			existingPrescription.AppointmentId = updateData.AppointmentId
			updated = true
		}
	}
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
	if updateData.PatientId != nil {
		if existingPrescription.PatientId != *updateData.PatientId {
			existingPrescription.PatientId = *updateData.PatientId
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
