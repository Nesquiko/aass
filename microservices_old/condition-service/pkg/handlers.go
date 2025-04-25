package pkg

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Nesquiko/aass/common/server"
	"github.com/Nesquiko/aass/condition-service/api"
)

// ConditionDetail implements api.ServerInterface.
func (c ConditionServer) ConditionDetail(
	w http.ResponseWriter,
	r *http.Request,
	conditionId api.PathConditionId,
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
	patientId api.PathPatientId,
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
	conditionId api.PathConditionId,
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

func (c ConditionServer) GetConditionsByPatientAndRange(
	w http.ResponseWriter,
	r *http.Request,
	params api.GetConditionsByPatientAndRangeParams,
) {
	// Convert openapi_types.Date to time.Time
	// Assuming dates represent the start of the day in the service's local timezone
	from := params.From.Time
	// Make 'to' date inclusive by setting time to end of day
	to := params.To.Time
	endOfDayTo := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	conditions, err := c.app.GetConditionsByPatientAndRange(
		r.Context(),
		params.PatientId,
		from,
		endOfDayTo,
	)
	if err != nil {
		slog.ErrorContext(
			r.Context(),
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"GetConditionsByPatientAndRange",
			"patientId",
			params.PatientId,
		)
		encodeError(w, internalServerError())
		return
	}

	response := api.Conditions{
		Conditions: conditions,
	}

	encode(w, http.StatusOK, response)
}
