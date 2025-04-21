package pkg

import (
	"log/slog"
	"net/http"

	"github.com/Nesquiko/aass/auth-service/api"
	"github.com/Nesquiko/aass/common/server"
)

// LoginUser implements api.ServerInterface.
func (s AuthServer) LoginUser(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := Decode[api.LoginUserJSONBody](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	if req.Role == api.UserRoleDoctor {
		doc, err := s.app.DoctorByEmail(r.Context(), string(req.Email))
		if err != nil {
			slog.Error(
				server.UnexpectedError,
				"error",
				err.Error(),
				"where",
				"GetPatientById",
				"role",
				"doctor",
			)
			encodeError(w, internalServerError())
			return
		}

		encode(w, http.StatusOK, doc)
		return
	}

	patient, err := s.app.PatientByEmail(r.Context(), string(req.Email))
	if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"GetPatientById",
			"role",
			"patient",
		)
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, patient)
}

// RegisterUser implements api.ServerInterface.
func (s AuthServer) RegisterUser(w http.ResponseWriter, r *http.Request) {
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
		if err != nil {
			slog.Error(
				server.UnexpectedError,
				"error",
				err.Error(),
				"where",
				"RegisterUser",
				"role",
				"doctor",
			)
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
	if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"RegisterUser",
			"role",
			"patient",
		)
		encodeError(w, internalServerError())
		return
	}
	encode(w, http.StatusCreated, patient)
}
