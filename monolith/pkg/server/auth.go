package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Nesquiko/wac/pkg/api"
	"github.com/Nesquiko/wac/pkg/app"
)

func (s Server) RegisterUser(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := Decode[api.Registration](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	regType, err := req.Discriminator()
	if err != nil {
		encodeError(w, decodeErrToApiError(err))
		return
	}
	if regType == string(api.UserRoleDoctor) {
		doctor, err := req.AsDoctorRegistration()
		if err != nil {
			encodeError(w, decodeErrToApiError(err))
			return
		}
		doc, err := s.app.CreateDoctor(r.Context(), doctor)
		if errors.Is(err, app.ErrDuplicateEmail) {
			apiErr := &ApiError{
				ErrorDetail: api.ErrorDetail{
					Code:   "doctor.email-exists",
					Title:  "Conflict",
					Detail: fmt.Sprintf("A doctor with email %q already exists.", doctor.Email),
					Status: http.StatusConflict,
				},
			}
			encodeError(w, apiErr)
			return
		} else if err != nil {
			slog.Error(UnexpectedError, "error", err.Error(), "where", "RegisterUser", "role", "doctor")
			encodeError(w, internalServerError())
			return
		}
		encode(w, http.StatusCreated, doc)
		return
	}

	pat, err := req.AsPatientRegistration()
	if err != nil {
		encodeError(w, decodeErrToApiError(err))
		return
	}
	patient, err := s.app.CreatePatient(r.Context(), pat)
	if errors.Is(err, app.ErrDuplicateEmail) {
		apiErr := &ApiError{
			ErrorDetail: api.ErrorDetail{
				Code:   "patient.email-exists",
				Title:  "Conflict",
				Detail: fmt.Sprintf("A patient with email %q already exists.", pat.Email),
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

// LoginUser implements api.ServerInterface.
func (s Server) LoginUser(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := Decode[api.LoginUserJSONBody](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	if req.Role == api.UserRoleDoctor {
		doc, err := s.app.DoctorByEmail(r.Context(), string(req.Email))
		if errors.Is(err, app.ErrNotFound) {
			apiErr := &ApiError{
				ErrorDetail: api.ErrorDetail{
					Code:   "doctor.not-found",
					Title:  "Not Found",
					Detail: fmt.Sprintf("Doctor with email %q not found.", req.Email),
					Status: http.StatusNotFound,
				},
			}
			encodeError(w, apiErr)
			return
		} else if err != nil {
			slog.Error(UnexpectedError, "error", err.Error(), "where", "GetPatientById", "role", "doctor")
			encodeError(w, internalServerError())
			return
		}

		encode(w, http.StatusOK, doc)
		return
	}

	patient, err := s.app.PatientByEmail(r.Context(), string(req.Email))
	if errors.Is(err, app.ErrNotFound) {
		apiErr := &ApiError{
			ErrorDetail: api.ErrorDetail{
				Code:   "patient.not-found",
				Title:  "Not Found",
				Detail: fmt.Sprintf("Patient with email %q not found.", req.Email),
				Status: http.StatusNotFound,
			},
		}
		encodeError(w, apiErr)
		return
	} else if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "GetPatientById", "role", "patient")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, patient)
}
