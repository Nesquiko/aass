package pkg

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/Nesquiko/aass/appointment-service/api"
)

// Define specific error codes for this service
const (
	AppointmentNotFoundCode   = "appointment.not.found"
	DoctorUnavailableCode     = "appointment.doctor.unavailable"
	AppointmentBadStateCode   = "appointment.invalid.state"
	AppointmentConflictCode   = "appointment.conflict" // General conflict/bad state
	AppointmentValidationCode = "appointment.validation.failed"
)

// RequestAppointment implements api.ServerInterface.
func (s AppointmentServer) RequestAppointment(w http.ResponseWriter, r *http.Request) {
	req, apiErr := Decode[api.NewAppointmentRequest](w, r)
	if apiErr != nil {
		encodeError(w, apiErr)
		return
	}

	// Basic validation (could be expanded)
	if req.AppointmentDateTime.IsZero() || req.DoctorId == uuid.Nil || req.PatientId == uuid.Nil {
		encodeError(w, &ApiError{api.ErrorDetail{
			Code:   ValidationErrorCode,
			Title:  ValidationErrorTitle,
			Detail: "Missing required fields: appointmentDateTime, doctorId, or patientId",
			Status: http.StatusBadRequest,
		}})
		return
	}

	createdAppt, err := s.app.RequestAppointment(r.Context(), req)
	if err != nil {
		handleAppError(w, err, "Appointment", uuid.Nil)
		return
	}

	doc, err := s.app.docClient.GetDoctorByIdWithResponse(r.Context(), createdAppt.DoctorId)
	if err != nil {
		slog.ErrorContext(
			r.Context(),
			"Can't fetch doc",
			"id",
			createdAppt.DoctorId,
			"error",
			err.Error(),
		)
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusCreated, mapAppointmentToPatientAppointment(createdAppt, *doc.JSON200))
}

// CancelAppointment implements api.ServerInterface.
func (s AppointmentServer) CancelAppointment(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	req, apiErr := Decode[api.AppointmentCancellation](w, r)
	if apiErr != nil {
		encodeError(w, apiErr)
		return
	}

	// Basic validation
	if req.By != api.UserRoleDoctor && req.By != api.UserRolePatient {
		encodeError(w, &ApiError{api.ErrorDetail{
			Code:   ValidationErrorCode,
			Title:  ValidationErrorTitle,
			Detail: "Invalid value for 'by' field",
			Status: http.StatusBadRequest,
		}})
		return
	}

	err := s.app.CancelAppointment(r.Context(), appointmentId, req)
	if err != nil {
		handleAppError(w, err, "Appointment", appointmentId)
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 No Content on successful delete/cancel
}

// RescheduleAppointment implements api.ServerInterface.
func (s AppointmentServer) RescheduleAppointment(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	req, apiErr := Decode[api.AppointmentReschedule](w, r)
	if apiErr != nil {
		encodeError(w, apiErr)
		return
	}

	if req.NewAppointmentDateTime.IsZero() {
		encodeError(w, &ApiError{api.ErrorDetail{
			Code:   ValidationErrorCode,
			Title:  ValidationErrorTitle,
			Detail: "Missing required field: newAppointmentDateTime",
			Status: http.StatusBadRequest,
		}})
		return
	}

	updatedAppt, err := s.app.RescheduleAppointment(r.Context(), appointmentId, req)
	if err != nil {
		handleAppError(w, err, "Appointment", appointmentId)
		return
	}

	doc, err := s.app.docClient.GetDoctorByIdWithResponse(r.Context(), updatedAppt.DoctorId)
	if err != nil {
		slog.ErrorContext(
			r.Context(),
			"Can't fetch doc",
			"id",
			updatedAppt.DoctorId,
			"error",
			err.Error(),
		)
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, mapAppointmentToPatientAppointment(updatedAppt, *doc.JSON200))
}

// DecideAppointment implements api.ServerInterface.
func (s AppointmentServer) DecideAppointment(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	req, apiErr := Decode[api.AppointmentDecision](w, r)
	if apiErr != nil {
		encodeError(w, apiErr)
		return
	}

	// Basic validation
	if req.Action != api.Accept && req.Action != api.Reject {
		encodeError(w, &ApiError{api.ErrorDetail{
			Code:   ValidationErrorCode,
			Title:  ValidationErrorTitle,
			Detail: "Invalid value for 'action' field",
			Status: http.StatusBadRequest,
		}})
		return
	}

	updatedAppt, err := s.app.DecideAppointment(r.Context(), appointmentId, req)
	if err != nil {
		handleAppError(w, err, "Appointment", appointmentId)
		return
	}

	pat, err := s.app.patClient.GetPatientByIdWithResponse(r.Context(), updatedAppt.PatientId)
	if err != nil {
		slog.ErrorContext(
			r.Context(),
			"Can't fetch pat",
			"id",
			updatedAppt.DoctorId,
			"error",
			err.Error(),
		)
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, mapAppointmentToDoctorAppointment(updatedAppt, *pat.JSON200))
}

// GetDoctorAppointmentById implements api.ServerInterface.
func (s AppointmentServer) GetDoctorAppointmentById(
	w http.ResponseWriter,
	r *http.Request,
	doctorId api.DoctorId,
	appointmentId api.AppointmentId,
) {
	appt, err := s.app.GetAppointmentById(r.Context(), appointmentId)
	if err != nil {
		handleAppError(w, err, "Appointment", appointmentId)
		return
	}

	// Verify the appointment belongs to the requested doctor
	if appt.DoctorId != doctorId {
		slog.WarnContext(
			r.Context(),
			"Doctor ID mismatch",
			"requestedDoctorId",
			doctorId,
			"actualDoctorId",
			appt.DoctorId,
			"appointmentId",
			appointmentId,
		)
		encodeError(w, notFoundId("Appointment for doctor", appointmentId))
		return
	}

	pat, err := s.app.patClient.GetPatientByIdWithResponse(r.Context(), appt.PatientId)
	if err != nil {
		slog.ErrorContext(
			r.Context(),
			"Can't fetch pat",
			"id",
			appt.DoctorId,
			"error",
			err.Error(),
		)
		encodeError(w, internalServerError())
		return
	}

	apiAppt := mapAppointmentToDoctorAppointment(appt, *pat.JSON200)
	encode(w, http.StatusOK, apiAppt)
}

// GetDoctorAppointments implements api.ServerInterface.
func (s AppointmentServer) GetDoctorAppointments(
	w http.ResponseWriter,
	r *http.Request,
	doctorId api.DoctorId,
	params api.GetDoctorAppointmentsParams,
) {
	appts, err := s.app.GetDoctorAppointments(r.Context(), doctorId, params)
	if err != nil {
		handleAppError(w, err, "Appointments for doctor", doctorId)
		return
	}

	apiAppts := make([]api.DoctorAppointment, len(appts))
	for i, appt := range appts {
		pat, err := s.app.patClient.GetPatientByIdWithResponse(r.Context(), appt.PatientId)
		if err != nil {
			slog.ErrorContext(
				r.Context(),
				"Can't fetch pat",
				"id",
				appt.DoctorId,
				"error",
				err.Error(),
			)
			encodeError(w, internalServerError())
			return
		}
		apiAppts[i] = mapAppointmentToDoctorAppointment(appt, *pat.JSON200)
	}

	if params.Status != nil {
		filteredAppts := make([]api.DoctorAppointment, 0)
		for _, appt := range apiAppts {
			if appt.Status == *params.Status {
				filteredAppts = append(filteredAppts, appt)
			}
		}
		apiAppts = filteredAppts
	}

	res := api.DoctorAppointmentsResponse{Appointments: apiAppts}
	encode(w, http.StatusOK, res)
}

// GetPatientAppointmentById implements api.ServerInterface.
func (s AppointmentServer) GetPatientAppointmentById(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
	appointmentId api.AppointmentId,
) {
	appt, err := s.app.GetAppointmentById(r.Context(), appointmentId)
	if err != nil {
		handleAppError(w, err, "Appointment", appointmentId)
		return
	}

	// Verify the appointment belongs to the requested patient
	if appt.PatientId != patientId {
		slog.WarnContext(
			r.Context(),
			"Patient ID mismatch",
			"requestedPatientId",
			patientId,
			"actualPatientId",
			appt.PatientId,
			"appointmentId",
			appointmentId,
		)
		encodeError(
			w,
			notFoundId("Appointment for patient", appointmentId),
		) // Treat as not found for this patient
		return
	}

	doc, err := s.app.docClient.GetDoctorByIdWithResponse(r.Context(), appt.DoctorId)
	if err != nil {
		slog.ErrorContext(
			r.Context(),
			"Can't fetch doc",
			"id",
			appt.DoctorId,
			"error",
			err.Error(),
		)
		encodeError(w, internalServerError())
		return
	}

	apiAppt := mapAppointmentToPatientAppointment(appt, *doc.JSON200)
	encode(w, http.StatusOK, apiAppt)
}

// GetPatientAppointments implements api.ServerInterface.
func (s AppointmentServer) GetPatientAppointments(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
	params api.GetPatientAppointmentsParams,
) {
	appts, err := s.app.GetPatientAppointments(r.Context(), patientId, params)
	if err != nil {
		handleAppError(w, err, "Appointments for patient", patientId)
		return
	}

	apiAppts := make([]api.PatientAppointment, len(appts))
	for i, appt := range appts {
		doc, err := s.app.docClient.GetDoctorByIdWithResponse(r.Context(), appt.DoctorId)
		if err != nil {
			slog.ErrorContext(
				r.Context(),
				"Can't fetch doc",
				"id",
				appt.DoctorId,
				"error",
				err.Error(),
			)
			encodeError(w, internalServerError())
			return
		}
		apiAppts[i] = mapAppointmentToPatientAppointment(appt, *doc.JSON200)
	}

	// Filter by status if provided
	if params.Status != nil {
		filteredAppts := make([]api.PatientAppointment, 0)
		for _, appt := range apiAppts {
			if appt.Status == *params.Status {
				filteredAppts = append(filteredAppts, appt)
			}
		}
		apiAppts = filteredAppts
	}

	res := api.PatientAppointmentsResponse{Appointments: apiAppts}
	encode(w, http.StatusOK, res)
}

func handleAppError(w http.ResponseWriter, err error, resource string, id uuid.UUID) {
	var apiErr *ApiError
	switch {
	case errors.Is(err, ErrAppointmentNotFound):
		apiErr = notFoundId(resource, id)
	case errors.Is(err, ErrDoctorNotAvailable):
		apiErr = &ApiError{api.ErrorDetail{
			Code:   DoctorUnavailableCode,
			Title:  "Doctor Unavailable",
			Detail: err.Error(),
			Status: http.StatusConflict, // 409 Conflict
		}}
	case errors.Is(err, ErrAppointmentBadState):
		apiErr = &ApiError{api.ErrorDetail{
			Code:   AppointmentBadStateCode,
			Title:  "Invalid Appointment State",
			Detail: err.Error(),
			Status: http.StatusConflict, // 409 Conflict
		}}
	case errors.Is(err, ErrInvalidDecisionAction), errors.Is(err, ErrMissingDenialReason):
		apiErr = &ApiError{api.ErrorDetail{
			Code:   AppointmentValidationCode,
			Title:  "Invalid Request Data",
			Detail: err.Error(),
			Status: http.StatusBadRequest, // 400 Bad Request
		}}
	case errors.Is(err, ErrIdMismatch):
		apiErr = &ApiError{api.ErrorDetail{
			Code:   AppointmentValidationCode,
			Title:  "ID Mismatch",
			Detail: err.Error(),
			Status: http.StatusBadRequest, // 400 Bad Request
		}}
	default:
		slog.Error("Unhandled application error", "error", err)
		apiErr = internalServerError()
	}
	encodeError(w, apiErr)
}
