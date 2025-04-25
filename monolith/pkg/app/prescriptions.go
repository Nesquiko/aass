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

func (a MonolithApp) CreatePatientPrescription(
	ctx context.Context,
	pres api.NewPrescription,
) (api.Prescription, error) {
	prescription, err := a.db.CreatePrescription(ctx, newPrescToDataPresc(pres))
	if err != nil {
		return api.Prescription{}, fmt.Errorf("CreatePatientPrescription: %w", err)
	}

	var appt *data.Appointment = nil
	var patient *data.Patient = nil
	var doctor *data.Doctor = nil

	if pres.AppointmentId != nil {
		appointment, err := a.db.AppointmentById(ctx, *pres.AppointmentId)
		if err != nil {
			return api.Prescription{}, fmt.Errorf("CreatePatientPrescription appt find: %w", err)
		}
		p, err := a.db.PatientById(ctx, appointment.PatientId)
		if err != nil {
			return api.Prescription{}, fmt.Errorf("CreatePatientPrescription patient find: %w", err)
		}
		patient = &p

		d, err := a.db.DoctorById(ctx, appointment.DoctorId)
		if err != nil {
			return api.Prescription{}, fmt.Errorf("CreatePatientPrescription doc find: %w", err)
		}
		doctor = &d
	}

	return dataPrescToPresc(prescription, appt, patient, doctor), nil
}

func (a MonolithApp) UpdatePatientPrescription(
	ctx context.Context,
	prescriptionId uuid.UUID,
	updateData api.UpdatePrescription,
) (api.Prescription, error) {
	existingPrescription, err := a.db.PrescriptionById(ctx, prescriptionId)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
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

	var updatedDbPrescription data.Prescription
	if updated {
		updatedDbPrescription, err = a.db.UpdatePrescription(
			ctx,
			prescriptionId,
			existingPrescription,
		)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
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

	patientData, err := a.db.PatientById(ctx, updatedDbPrescription.PatientId)
	if err != nil {
		return api.Prescription{}, fmt.Errorf(
			"UpdatePatientPrescription post-update patient fetch failed: %w",
			err,
		)
	}

	var apptData *data.Appointment = nil
	var doctorData *data.Doctor = nil

	if updatedDbPrescription.AppointmentId != nil {
		appointment, err := a.db.AppointmentById(
			ctx,
			*updatedDbPrescription.AppointmentId,
		)
		if err != nil {
			return api.Prescription{}, fmt.Errorf(
				"UpdatePatientPrescription post-update appointment fetch failed: %w",
				err,
			)
		}
		apptData = &appointment

		doctor, err := a.db.DoctorById(ctx, appointment.DoctorId)
		if err != nil {
			return api.Prescription{}, fmt.Errorf(
				"UpdatePatientPrescription post-update doctor fetch failed: %w",
				err,
			)
		}
		doctorData = &doctor
	}

	return dataPrescToPresc(updatedDbPrescription, apptData, &patientData, doctorData), nil
}

func (a MonolithApp) PrescriptionById(
	ctx context.Context,
	prescriptionId uuid.UUID,
) (api.Prescription, error) {
	prescription, err := a.db.PrescriptionById(ctx, prescriptionId)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return api.Prescription{}, fmt.Errorf("PrescriptionById: %w", ErrNotFound)
		}
		return api.Prescription{}, fmt.Errorf("PrescriptionById fetch prescription failed: %w", err)
	}

	patient, err := a.db.PatientById(ctx, prescription.PatientId)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return api.Prescription{}, fmt.Errorf(
				"PrescriptionById data inconsistency patient %s not found for prescription %s: %w",
				prescription.PatientId,
				prescriptionId,
				err,
			)
		}
		return api.Prescription{}, fmt.Errorf("PrescriptionById fetch patient failed: %w", err)
	}

	var apptData *data.Appointment = nil
	var doctorData *data.Doctor = nil

	if prescription.AppointmentId != nil {
		appointment, err := a.db.AppointmentById(ctx, *prescription.AppointmentId)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				return api.Prescription{}, fmt.Errorf(
					"PrescriptionById data inconsistency appointment %s not found for prescription %s: %w",
					*prescription.AppointmentId,
					prescriptionId,
					err,
				)
			}
			return api.Prescription{}, fmt.Errorf(
				"PrescriptionById fetch appointment failed: %w",
				err,
			)
		}
		apptData = &appointment

		doctor, err := a.db.DoctorById(ctx, appointment.DoctorId)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				return api.Prescription{}, fmt.Errorf(
					"PrescriptionById data inconsistency: doctor %s not found for appointment %s: %w",
					appointment.DoctorId,
					appointment.Id,
					err,
				)
			}
			return api.Prescription{}, fmt.Errorf(
				"PrescriptionById fetch doctor failed: %w",
				err,
			)
		}
		doctorData = &doctor
	}

	return dataPrescToPresc(prescription, apptData, &patient, doctorData), nil
}

func (a MonolithApp) DeletePrescription(ctx context.Context, id uuid.UUID) error {
	err := a.db.DeletePrescription(ctx, id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return fmt.Errorf(
				"DeletePrescription prescription with id %s not found: %w",
				id,
				ErrNotFound,
			)
		}
		return fmt.Errorf("DeletePrescription failed: %w", err)
	}

	return nil
}

func (a MonolithApp) PatientPrescriptionsInDateRange(
	ctx context.Context,
	patientId uuid.UUID,
	from time.Time,
	to *time.Time,
) ([]api.PrescriptionDisplay, error) {
	dataPresc, err := a.db.FindPrescriptionsByPatientId(ctx, patientId, from, to)
	if err != nil {
		return nil, fmt.Errorf("PatientPrescriptionsInDateRange failed: %w", err)
	}

	apiConditions := Map(dataPresc, dataPrescToPrescDisplay)
	return apiConditions, nil
}
