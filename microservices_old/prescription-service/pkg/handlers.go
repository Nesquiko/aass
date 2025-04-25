package pkg

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Nesquiko/aass/common/server"
	"github.com/Nesquiko/aass/prescription-service/api"
)

// CreatePrescription implements api.ServerInterface.
func (s PrescriptionServer) CreatePrescription(w http.ResponseWriter, r *http.Request) {
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
func (s PrescriptionServer) DeletePrescription(
	w http.ResponseWriter,
	r *http.Request,
	prescriptionId api.PathPrescriptionId,
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
func (s PrescriptionServer) PrescriptionDetail(
	w http.ResponseWriter,
	r *http.Request,
	prescriptionId api.PathPrescriptionId,
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

func (s PrescriptionServer) GetPrescriptionsByPatientAndRange(
	w http.ResponseWriter,
	r *http.Request,
	params api.GetPrescriptionsByPatientAndRangeParams,
) {
	// Convert openapi_types.Date to time.Time
	// Assuming dates represent the start of the day in the service's local timezone
	from := params.From.Time
	// Make 'to' date inclusive by setting time to end of day
	to := params.To.Time
	endOfDayTo := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, to.Location())

	prescriptions, err := s.app.GetPrescriptionsByPatientAndRange(
		r.Context(),
		params.PatientId,
		from,
		endOfDayTo,
	)
	if err != nil {
		// Handle potential errors (e.g., database connection issues)
		slog.ErrorContext(
			r.Context(),
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"GetPrescriptionsByPatientAndRange",
			"patientId",
			params.PatientId,
		)
		encodeError(w, internalServerError()) // Return generic 500
		return
	}

	// Wrap the result in the structure defined in the OpenAPI response
	response := api.Prescriptions{
		Prescriptions: prescriptions,
	}

	// Encode the successful response (even if the list is empty)
	encode(w, http.StatusOK, response)
}

// UpdatePrescription implements api.ServerInterface.
func (s PrescriptionServer) UpdatePrescription(
	w http.ResponseWriter,
	r *http.Request,
	prescriptionId api.PathPrescriptionId,
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
