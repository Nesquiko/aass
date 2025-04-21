package pkg

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Nesquiko/aass/common/server"
	"github.com/Nesquiko/aass/prescription-service/api"
)

// CreatePrescription implements api.ServerInterface.
func (s ResourceServer) CreatePrescription(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := Decode[api.NewPrescription](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	presc, err := s.app.CreatePatientPrescription(r.Context(), req)
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "CreatePrescription")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusCreated, presc)
}

// DeletePrescription implements api.ServerInterface.
func (s ResourceServer) DeletePrescription(
	w http.ResponseWriter,
	r *http.Request,
	prescriptionId api.PrescriptionId,
) {
	err := s.app.db.DeletePrescription(r.Context(), prescriptionId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			encodeError(w, notFoundId("Prescription", prescriptionId))
			return
		}

		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"DeletePrescription",
			"prescriptionId",
			prescriptionId.String(),
		)
		encodeError(w, internalServerError())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PrescriptionDetail implements api.ServerInterface.
func (s ResourceServer) PrescriptionDetail(
	w http.ResponseWriter,
	r *http.Request,
	prescriptionId api.PrescriptionId,
) {
	prescription, err := s.app.PrescriptionById(r.Context(), prescriptionId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			encodeError(w, notFoundId("Prescription", prescriptionId))
			return
		}
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"PrescriptionDetail",
			"prescriptionId",
			prescriptionId.String(),
		)
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, prescription)
}

// UpdatePrescription implements api.ServerInterface.
func (s ResourceServer) UpdatePrescription(
	w http.ResponseWriter,
	r *http.Request,
	prescriptionId api.PrescriptionId,
) {
	req, decodeErr := Decode[api.UpdatePrescription](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	updatedPrescription, err := s.app.UpdatePatientPrescription(
		r.Context(),
		prescriptionId,
		req,
	)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			encodeError(w, notFoundId("Prescription", prescriptionId))
			return
		}
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"UpdatePrescription",
			"prescriptionId",
			prescriptionId.String(),
		)
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, updatedPrescription)
}
