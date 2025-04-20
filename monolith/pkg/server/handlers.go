package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Nesquiko/wac/pkg/api"
	"github.com/Nesquiko/wac/pkg/app"
)

// CancelAppointment implements api.ServerInterface.
func (s Server) CancelAppointment(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	req, decodeErr := Decode[api.AppointmentCancellation](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	err := s.app.CancelAppointment(r.Context(), appointmentId, req)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "CancelAppointment")
		encodeError(w, internalServerError())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ConditionDetail implements api.ServerInterface.
func (s Server) ConditionDetail(
	w http.ResponseWriter,
	r *http.Request,
	conditionId api.ConditionId,
) {
	cond, err := s.app.ConditionById(r.Context(), conditionId)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "GetAvailableResources")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, cond)
}

// DecideAppointment implements api.ServerInterface.
func (s Server) DecideAppointment(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	req, decodeErr := Decode[api.AppointmentDecision](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	doctorAppt, err := s.app.DecideAppointment(r.Context(), appointmentId, req)
	if err != nil {
		if errors.Is(err, app.ErrResourceUnavailable) {
			apiErr := &ApiError{
				ErrorDetail: api.ErrorDetail{
					Code:   "resources.unavailable",
					Title:  "Conflict",
					Detail: "Requested resources are unavailable",
					Status: http.StatusConflict,
				},
			}
			encodeError(w, apiErr)
			return
		}
		slog.Error(UnexpectedError, "error", err.Error(), "where", "DecideAppointment")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, doctorAppt)
}

// DoctorsCalendar implements api.ServerInterface.
func (s Server) DoctorsCalendar(
	w http.ResponseWriter,
	r *http.Request,
	doctorId api.DoctorId,
	params api.DoctorsCalendarParams,
) {
	calendar, err := s.app.DoctorsCalendar(r.Context(), doctorId, params.From, params.To)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "DoctorsCalendar")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, calendar)
}

// DoctorsTimeslots implements api.ServerInterface.
func (s Server) DoctorsTimeslots(
	w http.ResponseWriter,
	r *http.Request,
	doctorId api.DoctorId,
	params api.DoctorsTimeslotsParams,
) {
	slots, err := s.app.DoctorTimeSlots(r.Context(), doctorId, params.Date.Time)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "DoctorsCalendar")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, slots)
}

// GetAvailableResources implements api.ServerInterface.
func (s Server) GetAvailableResources(
	w http.ResponseWriter,
	r *http.Request,
	params api.GetAvailableResourcesParams,
) {
	resources, err := s.app.AvailableResources(r.Context(), params.DateTime)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "GetAvailableResources")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, resources)
}

// GetDoctorById implements api.ServerInterface.
func (s Server) GetDoctorById(w http.ResponseWriter, r *http.Request, doctorId api.DoctorId) {
	doctor, err := s.app.DoctorById(r.Context(), doctorId)
	if err != nil {
		if errors.Is(err, app.ErrNotFound) {
			apiErr := &ApiError{
				ErrorDetail: api.ErrorDetail{
					Code:   "doctor.not-found",
					Title:  "Not Found",
					Detail: fmt.Sprintf("Doctor with ID %q not found.", doctorId),
					Status: http.StatusNotFound,
				},
			}
			encodeError(w, apiErr)
			return
		}
		slog.Error(UnexpectedError, "error", err.Error(), "where", "GetDoctorById")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, doctor)
}

// DoctorsAppointment implements api.ServerInterface.
func (s Server) DoctorsAppointment(
	w http.ResponseWriter,
	r *http.Request,
	doctorId api.DoctorId,
	appointmentId api.AppointmentId,
) {
	appt, err := s.app.DoctorsAppointmentById(r.Context(), doctorId, appointmentId)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "DoctorsAppointment")
		encodeError(w, internalServerError())
		return
	}
	encode(w, http.StatusOK, appt)
}

// GetPatientById implements api.ServerInterface.
func (s Server) GetPatientById(w http.ResponseWriter, r *http.Request, patientId api.PatientId) {
	patient, err := s.app.PatientById(r.Context(), patientId)
	if err != nil {
		if errors.Is(err, app.ErrNotFound) {
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

// PatientsAppointment implements api.ServerInterface.
func (s Server) PatientsAppointment(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
	appointmentId api.AppointmentId,
) {
	appt, err := s.app.PatientsAppointmentById(r.Context(), patientId, appointmentId)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "PatientsAppointment")
		encodeError(w, internalServerError())
		return
	}
	encode(w, http.StatusOK, appt)
}

// PatientsCalendar implements api.ServerInterface.
func (s Server) PatientsCalendar(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
	params api.PatientsCalendarParams,
) {
	calendar, err := s.app.PatientsCalendar(r.Context(), patientId, params.From, params.To)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "PatientsCalendar")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, calendar)
}

// PatientsMedicalHistoryFiles implements api.ServerInterface.
func (s Server) PatientsMedicalHistoryFiles(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
	params api.PatientsMedicalHistoryFilesParams,
) {
	files, err := s.app.PatientMedicalHistoryFiles(
		r.Context(),
		patientId,
		params.Page,
		params.PageSize,
	)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "PatientMedicalHistoryFiles")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, files)
}

// RescheduleAppointment implements api.ServerInterface.
func (s Server) RescheduleAppointment(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	req, decodeErr := Decode[api.AppointmentReschedule](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	appt, err := s.app.RescheduleAppointment(r.Context(), appointmentId, req.NewAppointmentDateTime)
	if err != nil {
		if errors.Is(err, app.ErrDoctorUnavailable) {
			apiErr := &ApiError{
				ErrorDetail: api.ErrorDetail{
					Code:   "doctor.unavailable",
					Title:  "Conflict",
					Detail: "Doctor is unavailable in requested reschedule time",
					Status: http.StatusConflict,
				},
			}
			encodeError(w, apiErr)
			return
		}
		slog.Error(UnexpectedError, "error", err.Error(), "where", "RescheduleAppointment")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, appt)
}

// CreatePatientCondition implements api.ServerInterface.
func (s Server) CreatePatientCondition(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := Decode[api.NewCondition](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	cond, err := s.app.CreatePatientCondition(r.Context(), req)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "CreatePatientCondition")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusCreated, cond)
}

// CreatePatientMedicine implements api.ServerInterface.
func (s Server) CreatePrescription(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := Decode[api.NewPrescription](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	presc, err := s.app.CreatePatientPrescription(r.Context(), req)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "CreatePrescription")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusCreated, presc)
}

// RequestAppointment implements api.ServerInterface.
func (s Server) RequestAppointment(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := Decode[api.NewAppointmentRequest](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	appt, err := s.app.CreateAppointment(r.Context(), req)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "RequestAppointment")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusCreated, appt)
}

// CreateResource implements api.ServerInterface.
func (s Server) CreateResource(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := Decode[api.NewResource](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	resource, err := s.app.CreateResource(r.Context(), req)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "CreateResource")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusCreated, resource)
}

// ReserveResource implements api.ServerInterface.
func (s Server) ReserveResource(w http.ResponseWriter, r *http.Request, resourceId api.ResourceId) {
	req, decodeErr := Decode[api.ResourceReservation](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	err := s.app.ReserveResource(r.Context(), resourceId, req)
	if err != nil {
		slog.Error(UnexpectedError, "error", err.Error(), "where", "ReserveResource")
		encodeError(w, internalServerError())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AvailableDoctors implements api.ServerInterface.
func (s Server) AvailableDoctors(
	w http.ResponseWriter,
	r *http.Request,
	params api.AvailableDoctorsParams,
) {
	doctors, err := s.app.AvailableDoctors(r.Context(), params.DateTime)
	if err != nil {
		slog.Error(
			UnexpectedError,
			"error",
			err.Error(),
			"where",
			"AvailableDoctors",
			"dateTime",
			params.DateTime.String(),
		)
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, doctors)
}

// PrescriptionDetail implements api.ServerInterface.
func (s Server) PrescriptionDetail(
	w http.ResponseWriter,
	r *http.Request,
	prescriptionId api.PrescriptionId,
) {
	prescription, err := s.app.PrescriptionById(r.Context(), prescriptionId)
	if err != nil {
		if errors.Is(err, app.ErrNotFound) {
			encodeError(w, notFoundId("Prescription", prescriptionId))
			return
		}
		slog.Error(
			UnexpectedError,
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

func (s Server) ReserveAppointmentResources(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId, // Parsed from the path
) {
	req, decodeErr := Decode[api.ReserveAppointmentResourcesJSONBody](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	updatedAppointment, err := s.app.ReserveAppointmentResources(
		r.Context(),
		appointmentId,
		req,
	)
	if err != nil {
		if errors.Is(err, app.ErrNotFound) {
			apiErr := &ApiError{
				ErrorDetail: api.ErrorDetail{
					Code:   "resource.or.appointment.not-found",
					Title:  "Not Found",
					Detail: err.Error(),
					Status: http.StatusNotFound,
				},
			}
			encodeError(w, apiErr)
			return
		} else if errors.Is(err, app.ErrResourceUnavailable) {
			apiErr := &ApiError{
				ErrorDetail: api.ErrorDetail{
					Code:   "resource.unavailable",
					Title:  "Conflict",
					Detail: err.Error(),
					Status: http.StatusConflict,
				},
			}
			encodeError(w, apiErr)
			return
		}

		slog.Error(
			UnexpectedError,
			"error",
			err.Error(),
			"where",
			"ReserveAppointmentResources",
			"appointmentId",
			appointmentId.String(),
		)
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, updatedAppointment)
}

// UpdateCondition implements api.ServerInterface.
func (s Server) UpdateCondition(
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
		if errors.Is(err, app.ErrNotFound) {
			encodeError(w, notFoundId("Condition", conditionId))
			return
		}
		slog.Error(
			UnexpectedError,
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

func (s Server) UpdatePrescription(
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
		if errors.Is(err, app.ErrNotFound) {
			encodeError(w, notFoundId("Prescription", prescriptionId))
			return
		}
		slog.Error(
			UnexpectedError,
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

// GetDoctors implements api.ServerInterface.
func (s Server) GetDoctors(w http.ResponseWriter, r *http.Request) {
	doctors, err := s.app.GetAllDoctors(r.Context())
	if err != nil {
		slog.Error(
			UnexpectedError,
			"error",
			err.Error(),
			"where",
			"GetAllDoctors",
		)
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, api.Doctors{Doctors: doctors})
}

// ConditionsInDate implements api.ServerInterface.
func (s Server) ConditionsInDate(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
	params api.ConditionsInDateParams,
) {
	conditions, err := s.app.PatientConditionsOnDate(
		r.Context(),
		patientId,
		params.Date.Time,
	)
	if err != nil {
		slog.Error(
			UnexpectedError,
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

// DeletePrescription implements api.ServerInterface.
func (s Server) DeletePrescription(
	w http.ResponseWriter,
	r *http.Request,
	prescriptionId api.PrescriptionId,
) {
	err := s.app.DeletePrescription(r.Context(), prescriptionId)
	if err != nil {
		if errors.Is(err, app.ErrNotFound) {
			encodeError(w, notFoundId("Prescription", prescriptionId))
			return
		}

		slog.Error(
			UnexpectedError,
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
