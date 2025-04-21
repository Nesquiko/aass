package pkg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Nesquiko/aass/condition-service/api"
)

type ConditionApp struct {
	db *MongoConditionDb
}

func (a ConditionApp) CreatePatientCondition(
	ctx context.Context,
	c api.NewCondition,
) (api.ConditionDisplay, error) {
	cond, err := a.db.CreateCondition(ctx, newCondToDataCond(c))
	if err != nil {
		return api.ConditionDisplay{}, fmt.Errorf("CreatePatientCondition: %w", err)
	}
	return dataCondToCondDisplay(cond), nil
}

func (a ConditionApp) ConditionById(ctx context.Context, id uuid.UUID) (api.Condition, error) {
	cond, err := a.db.ConditionById(ctx, id)
	if err != nil {
		return api.Condition{}, fmt.Errorf("ConditionById: %w", err)
	}

	return dataCondToCond(cond), nil
}

func (a ConditionApp) UpdatePatientCondition(
	ctx context.Context,
	conditionId uuid.UUID,
	updateData api.UpdateCondition,
) (api.Condition, error) {
	existingCondition, err := a.db.ConditionById(ctx, conditionId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
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

	var finalConditionData Condition
	if updated {
		updatedDbResult, err := a.db.UpdateCondition(
			ctx,
			conditionId,
			existingCondition,
		)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
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

	apiResponse := dataCondToCond(finalConditionData)
	return apiResponse, nil
}

func (a ConditionApp) PatientConditionsOnDate(
	ctx context.Context,
	patientId uuid.UUID,
	date time.Time,
) ([]api.ConditionDisplay, error) {
	dataConditions, err := a.db.FindConditionsByPatientIdAndDate(ctx, patientId, date)
	if err != nil {
		return nil, fmt.Errorf("PatientConditionsOnDate failed: %w", err)
	}

	apiConditions := Map(dataConditions, dataCondToCondDisplay)
	return apiConditions, nil
}
