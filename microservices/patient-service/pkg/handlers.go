package pkg

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Nesquiko/aass/common/server"
	"github.com/Nesquiko/aass/patient-service/api"
)

func (s PatientServer) CreatePatient(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := Decode[api.NewPatientRequest](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	patient, err := s.app.CreatePatient(r.Context(), req)
	if errors.Is(err, ErrDuplicateEmail) {
		apiErr := &ApiError{
			ErrorDetail: api.ErrorDetail{
				Code:   "patient.email-exists",
				Title:  "Conflict",
				Detail: fmt.Sprintf("A patient with email %q already exists.", req.Email),
				Status: http.StatusConflict,
			},
		}
		encodeError(w, apiErr)
		return
	} else if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "RegisterUser", "role", "patient")
		encodeError(w, internalServerError())
		return
	}
	encode(w, http.StatusCreated, patient)
}

func (s PatientServer) GetPatientById(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
) {
	patient, err := s.app.PatientById(r.Context(), patientId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiErr := &ApiError{
				ErrorDetail: api.ErrorDetail{
					Code:   "patient.not-found",
					Title:  "Not Found",
					Detail: fmt.Sprintf("Patient with ID %q not found.", patientId),
					Status: http.StatusNotFound,
				},
			}
			encodeError(w, apiErr)
			return
		}
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "GetPatientById")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, patient)
}

func (s PatientServer) GetPatientByEmail(
	w http.ResponseWriter,
	r *http.Request,
	params api.GetPatientByEmailParams,
) {
	patient, err := s.app.PatientByEmail(r.Context(), string(params.Email))
	if errors.Is(err, ErrNotFound) {
		apiErr := &ApiError{
			ErrorDetail: api.ErrorDetail{
				Code:   "patient.not-found",
				Title:  "Not Found",
				Detail: fmt.Sprintf("Patient with email %q not found.", params.Email),
				Status: http.StatusNotFound,
			},
		}
		encodeError(w, apiErr)
		return
	} else if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "GetPatientByEmail", "role", "patient")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, patient)
}

// GetPatientCalendar implements api.ServerInterface.
func (s PatientServer) GetPatientCalendar(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
	params api.GetPatientCalendarParams,
) {
	from := params.From.Time
	to := params.To.Time
	endOfDayTo := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	calendarView, err := s.app.GetPatientCalendar(r.Context(), patientId, from, endOfDayTo)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiErr := &ApiError{
				ErrorDetail: api.ErrorDetail{
					Code:   "patient.not-found", // Consistent code
					Title:  "Not Found",
					Detail: fmt.Sprintf("Patient with ID %q not found.", patientId),
					Status: http.StatusNotFound,
				},
			}
			encodeError(w, apiErr)
			return
		}
		// Handle potential downstream errors
		if errors.Is(err, ErrDownstreamService) {
			slog.ErrorContext(
				r.Context(),
				"Failed to get patient calendar due to downstream service error",
				"patientId",
				patientId,
				"error",
				err,
			)
			encodeError(w, internalServerError()) // Return generic 500
			return
		}
		// Handle other unexpected errors
		slog.ErrorContext(
			r.Context(),
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"GetPatientCalendar",
			"patientId",
			patientId,
		)
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, calendarView)
}
