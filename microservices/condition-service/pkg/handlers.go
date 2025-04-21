package pkg

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Nesquiko/aass/common/server"
	"github.com/Nesquiko/aass/condition-service/api"
)

// ConditionDetail implements api.ServerInterface.
func (c ConditionServer) ConditionDetail(
	w http.ResponseWriter,
	r *http.Request,
	conditionId api.ConditionId,
) {
	cond, err := c.app.ConditionById(r.Context(), conditionId)
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "ConditionDetail")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, cond)
}

// ConditionsInDate implements api.ServerInterface.
func (c ConditionServer) ConditionsInDate(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
	params api.ConditionsInDateParams,
) {
	conditions, err := c.app.PatientConditionsOnDate(
		r.Context(),
		patientId,
		params.Date.Time,
	)
	if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"ConditionsInDate",
			"patientId",
			patientId.String(),
			"date",
			params.Date.String(),
		)
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, api.Conditions{Conditions: conditions})
}

// CreatePatientCondition implements api.ServerInterface.
func (s ConditionServer) CreatePatientCondition(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := Decode[api.NewCondition](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	cond, err := s.app.CreatePatientCondition(r.Context(), req)
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "CreatePatientCondition")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusCreated, cond)
}

// UpdateCondition implements api.ServerInterface.
func (s ConditionServer) UpdateCondition(
	w http.ResponseWriter,
	r *http.Request,
	conditionId api.ConditionId,
) {
	req, decodeErr := Decode[api.UpdateCondition](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	updatedCondition, err := s.app.UpdatePatientCondition(
		r.Context(),
		conditionId,
		req,
	)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			encodeError(w, notFoundId("Condition", conditionId))
			return
		}
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"UpdateCondition",
			"conditionId",
			conditionId.String(),
		)
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, updatedCondition)
}
