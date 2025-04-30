package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/httplog/v2"

	"github.com/Nesquiko/aass/common/server"
	commonapi "github.com/Nesquiko/aass/common/server/api"
	"github.com/Nesquiko/aass/user-service/api"
)

type userServer struct{ db mongoUserDb }

func newUserServer(
	db mongoUserDb,
	logger *httplog.Logger,
	opts commonapi.ChiServerOptions,
) http.Handler {
	srv := userServer{db: db}

	middlewares := make([]api.MiddlewareFunc, len(opts.Middlewares))
	for i, mid := range opts.Middlewares {
		middlewares[i] = api.MiddlewareFunc(mid)
	}

	mappedOpts := api.ChiServerOptions{
		BaseURL:          opts.BaseURL,
		BaseRouter:       opts.BaseRouter,
		Middlewares:      middlewares,
		ErrorHandlerFunc: opts.ErrorHandlerFunc,
	}

	return api.HandlerWithOptions(srv, mappedOpts)
}

// GetDoctorById implements api.ServerInterface.
func (u userServer) GetDoctorById(w http.ResponseWriter, r *http.Request, doctorId api.DoctorId) {
	doctor, err := u.db.DoctorById(r.Context(), doctorId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiErr := &server.ApiError{
				ErrorDetail: commonapi.ErrorDetail{
					Code:   "doctor.not-found",
					Title:  "Not Found",
					Detail: fmt.Sprintf("Doctor with ID %q not found.", doctorId),
					Status: http.StatusNotFound,
				},
			}
			server.EncodeError(w, apiErr)
			return
		}
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "GetDoctorById")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	server.Encode(w, http.StatusOK, dataDoctorToApiDoctor(doctor))
}

// GetDoctors implements api.ServerInterface.
func (u userServer) GetDoctors(w http.ResponseWriter, r *http.Request) {
	doctors, err := u.db.GetAllDoctors(r.Context())
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "GetDoctors")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	server.Encode(
		w,
		http.StatusOK,
		api.Doctors{Doctors: server.Map(doctors, dataDoctorToApiDoctor)},
	)
}

// GetPatientById implements api.ServerInterface.
func (u userServer) GetPatientById(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
) {
	patient, err := u.db.FindPatientById(r.Context(), patientId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			apiErr := &server.ApiError{
				ErrorDetail: commonapi.ErrorDetail{
					Code:   "patient.not-found",
					Title:  "Not Found",
					Detail: fmt.Sprintf("Patient with ID %q not found.", patientId),
					Status: http.StatusNotFound,
				},
			}
			server.EncodeError(w, apiErr)
			return
		}
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "GetPatientById")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	server.Encode(w, http.StatusOK, dataPatientToApiPatient(patient))
}

// LoginUser implements api.ServerInterface.
func (u userServer) LoginUser(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := server.Decode[api.LoginUserJSONBody](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	if req.Role == api.UserRoleDoctor {
		doc, err := u.db.DoctorByEmail(r.Context(), string(req.Email))
		if errors.Is(err, ErrNotFound) {
			apiErr := &server.ApiError{
				ErrorDetail: commonapi.ErrorDetail{
					Code:   "doctor.not-found",
					Title:  "Not Found",
					Detail: fmt.Sprintf("Doctor with email %q not found.", req.Email),
					Status: http.StatusNotFound,
				},
			}
			server.EncodeError(w, apiErr)
			return
		} else if err != nil {
			slog.Error(server.UnexpectedError, "error", err.Error(), "where", "GetPatientById", "role", "doctor")
			server.EncodeError(w, server.InternalServerError())
			return
		}

		server.Encode(w, http.StatusOK, dataDoctorToApiDoctor(doc))
		return
	}

	patient, err := u.db.FindPatientByEmail(r.Context(), string(req.Email))
	if errors.Is(err, ErrNotFound) {
		apiErr := &server.ApiError{
			ErrorDetail: commonapi.ErrorDetail{
				Code:   "patient.not-found",
				Title:  "Not Found",
				Detail: fmt.Sprintf("Patient with email %q not found.", req.Email),
				Status: http.StatusNotFound,
			},
		}
		server.EncodeError(w, apiErr)
		return
	} else if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "GetPatientById", "role", "patient")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	server.Encode(w, http.StatusOK, dataPatientToApiPatient(patient))
}

// RegisterUser implements api.ServerInterface.
func (u userServer) RegisterUser(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := server.Decode[api.Registration](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	regType, err := req.Discriminator()
	if err != nil {
		server.EncodeError(w, server.DecodeErrToApiError(err))
		return
	}
	if regType == string(api.UserRoleDoctor) {
		doctor, err := req.AsDoctorRegistration()
		if err != nil {
			server.EncodeError(w, server.DecodeErrToApiError(err))
			return
		}
		doc, err := u.db.CreateDoctor(r.Context(), doctorRegToDataDoctor(doctor))
		if errors.Is(err, ErrDuplicateEmail) {
			apiErr := &server.ApiError{
				ErrorDetail: commonapi.ErrorDetail{
					Code:   "doctor.email-exists",
					Title:  "Conflict",
					Detail: fmt.Sprintf("A doctor with email %q already exists.", doctor.Email),
					Status: http.StatusConflict,
				},
			}
			server.EncodeError(w, apiErr)
			return
		} else if err != nil {
			slog.Error(server.UnexpectedError, "error", err.Error(), "where", "RegisterUser", "role", "doctor")
			server.EncodeError(w, server.InternalServerError())
			return
		}
		server.Encode(w, http.StatusCreated, dataDoctorToApiDoctor(doc))
		return
	}

	pat, err := req.AsPatientRegistration()
	if err != nil {
		server.EncodeError(w, server.DecodeErrToApiError(err))
		return
	}
	patient, err := u.db.CreatePatient(r.Context(), patientRegToDataPatient(pat))
	if errors.Is(err, ErrDuplicateEmail) {
		apiErr := &server.ApiError{
			ErrorDetail: commonapi.ErrorDetail{
				Code:   "patient.email-exists",
				Title:  "Conflict",
				Detail: fmt.Sprintf("A patient with email %q already exists.", pat.Email),
				Status: http.StatusConflict,
			},
		}
		server.EncodeError(w, apiErr)
		return
	} else if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "RegisterUser", "role", "patient")
		server.EncodeError(w, server.InternalServerError())
		return
	}
	server.Encode(w, http.StatusCreated, dataPatientToApiPatient(patient))
}
