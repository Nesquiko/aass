package pkg

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

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
		slog.Error(UnexpectedError, "error", err.Error(), "where", "RegisterUser", "role", "patient")
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
		slog.Error(UnexpectedError, "error", err.Error(), "where", "GetPatientById")
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
		slog.Error(UnexpectedError, "error", err.Error(), "where", "GetPatientByEmail", "role", "patient")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, patient)
}
