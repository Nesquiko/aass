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

func (a MonolithApp) CreatePatientCondition(
	ctx context.Context,
	c api.NewCondition,
) (api.ConditionDisplay, error) {
	cond, err := a.db.CreateCondition(ctx, newCondToDataCond(c))
	if err != nil {
		return api.ConditionDisplay{}, fmt.Errorf("CreatePatientCondition: %w", err)
	}
	return dataCondToCondDisplay(cond), nil
}

func (a MonolithApp) ConditionById(ctx context.Context, id uuid.UUID) (api.Condition, error) {
	cond, err := a.db.ConditionById(ctx, id)
	if err != nil {
		return api.Condition{}, fmt.Errorf("ConditionById: %w", err)
	}

	appts, err := a.db.AppointmentsByConditionId(ctx, cond.Id)
	if err != nil {
		return api.Condition{}, fmt.Errorf("ConditionById find appts: %w", err)
	}

	patient, err := a.db.PatientById(ctx, cond.PatientId)
	if err != nil {
		return api.Condition{}, fmt.Errorf("ConditionById find patient: %w", err)
	}

	appointments := make([]api.AppointmentDisplay, len(appts))
	for i, appt := range appts {
		doc, err := a.db.DoctorById(ctx, appt.DoctorId)
		if err != nil {
			return api.Condition{}, fmt.Errorf("ConditionById find doctor: %w", err)
		}
		appointments[i] = dataApptToApptDisplay(appt, patient, doc)
	}

	return dataCondToCond(cond, appointments), nil
}

func (a MonolithApp) UpdatePatientCondition(
	ctx context.Context,
	conditionId uuid.UUID,
	updateData api.UpdateCondition,
) (api.Condition, error) {
	existingCondition, err := a.db.ConditionById(ctx, conditionId)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return api.Condition{}, fmt.Errorf("UpdatePatientCondition: %w", ErrNotFound)
		}
		return api.Condition{}, fmt.Errorf("UpdatePatientCondition fetch failed: %w", err)
	}

	updated := false
	if updateData.End != nil {
		if updateData.End.IsNull() && existingCondition.End != nil {
			existingCondition.End = nil
			updated = true
		} else if updateData.End.IsSpecified() && (existingCondition.End == nil || !existingCondition.End.Equal(updateData.End.MustGet())) {
			existingCondition.End = asPtr(updateData.End.MustGet())
			updated = true
		}
	}
	if updateData.Name != nil {
		if existingCondition.Name != *updateData.Name {
			existingCondition.Name = *updateData.Name
			updated = true
		}
	}
	if updateData.PatientId != nil {
		if existingCondition.PatientId != *updateData.PatientId {
			existingCondition.PatientId = *updateData.PatientId
			updated = true
		}
	}
	if updateData.Start != nil {
		if !existingCondition.Start.Equal(*updateData.Start) {
			existingCondition.Start = *updateData.Start
			updated = true
		}
	}

	var finalConditionData data.Condition
	if updated {
		updatedDbResult, err := a.db.UpdateCondition(
			ctx,
			conditionId,
			existingCondition,
		)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				return api.Condition{}, fmt.Errorf(
					"UpdatePatientCondition update target vanished: %w",
					ErrNotFound,
				)
			}
			return api.Condition{}, fmt.Errorf(
				"UpdatePatientCondition update failed: %w",
				err,
			)
		}
		finalConditionData = updatedDbResult
	} else {
		finalConditionData = existingCondition
	}

	apptsData, err := a.db.AppointmentsByConditionId(
		ctx,
		finalConditionData.Id,
	)
	if err != nil {
		return api.Condition{}, fmt.Errorf(
			"UpdatePatientCondition fetch appointments failed: %w",
			err,
		)
	}

	patientData, err := a.db.PatientById(ctx, finalConditionData.PatientId)
	if err != nil {
		return api.Condition{}, fmt.Errorf(
			"UpdatePatientCondition data inconsistency: patient %s not found: %w",
			finalConditionData.PatientId,
			err,
		)
	}

	apiAppointments := make([]api.AppointmentDisplay, len(apptsData))
	for i, appt := range apptsData {
		doctorData, err := a.db.DoctorById(ctx, appt.DoctorId)
		if err != nil {
			return api.Condition{}, fmt.Errorf(
				"UpdatePatientCondition data inconsistency: doctor %s for appointment %s not found: %w",
				appt.DoctorId,
				appt.Id,
				err,
			)
		}
		apiAppointments[i] = dataApptToApptDisplay(
			appt,
			patientData,
			doctorData,
		)
	}

	apiResponse := dataCondToCond(finalConditionData, apiAppointments)
	return apiResponse, nil
}

func (a MonolithApp) PatientConditionsOnDate(
	ctx context.Context,
	patientId uuid.UUID,
	from time.Time,
	to *time.Time,
) ([]api.ConditionDisplay, error) {
	dataConditions, err := a.db.FindConditionsByPatientId(ctx, patientId, from, to)
	if err != nil {
		return nil, fmt.Errorf("PatientConditionsOnDate failed: %w", err)
	}

	apiConditions := Map(dataConditions, dataCondToCondDisplay)
	return apiConditions, nil
}
