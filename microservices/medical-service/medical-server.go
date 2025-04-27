package main

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/httplog/v2"

	"github.com/Nesquiko/aass/common/server"
	commonapi "github.com/Nesquiko/aass/common/server/api"
	"github.com/Nesquiko/aass/medical-service/api"
	appointmentapi "github.com/Nesquiko/aass/medical-service/appointment-api"
)

type medicalServer struct {
	db      mongoMedicalDb
	apptApi *appointmentapi.ClientWithResponses
}

func newMedicalServer(
	db mongoMedicalDb,
	logger *httplog.Logger,
	opts commonapi.ChiServerOptions,
) http.Handler {
	apptClient, _ := appointmentapi.NewClientWithResponses("http://appointment-service:8080/")
	srv := medicalServer{db: db, apptApi: apptClient}

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

// ConditionDetail implements api.ServerInterface.
func (m medicalServer) ConditionDetail(
	w http.ResponseWriter,
	r *http.Request,
	conditionId api.ConditionId,
) {
	cond, err := m.db.ConditionById(r.Context(), conditionId)
	if errors.Is(err, ErrNotFound) {
		server.EncodeError(w, server.NotFoundId("Condition", conditionId))
		return
	} else if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "ConditionDetail")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	res, err := m.apptApi.AppointmentsByConditionIdWithResponse(r.Context(), cond.Id)
	if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"ConditionDetail fetch appts",
		)
		server.EncodeError(w, server.InternalServerError())
		return
	} else if res.JSON200 == nil {
		slog.Error("no appointments response")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	appts := []appointmentapi.AppointmentDisplay{}
	if res.JSON200.Appointments != nil {
		appts = *res.JSON200.Appointments
	}

	server.Encode(w, http.StatusOK, dataCondToCond(cond, appts))
}

// ConditionsInDateRange implements api.ServerInterface.
func (m medicalServer) ConditionsInDateRange(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
	params api.ConditionsInDateRangeParams,
) {
	var to *time.Time = nil
	if params.To != nil {
		to = &params.To.Time
	}

	conditions, err := m.db.FindConditionsByPatientId(
		r.Context(),
		patientId,
		params.From.Time,
		to,
	)
	if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"ConditionsInDateRange",
			"patientId",
			patientId.String(),
			"from",
			params.From,
			"to",
			params.To,
		)
		server.EncodeError(w, server.InternalServerError())
		return
	}

	server.Encode(
		w,
		http.StatusOK,
		api.Conditions{Conditions: server.Map(conditions, dataCondToCondDisplay)},
	)
}

// CreatePatientCondition implements api.ServerInterface.
func (m medicalServer) CreatePatientCondition(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := server.Decode[api.NewCondition](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	cond, err := m.db.CreateCondition(r.Context(), newCondToDataCond(req))
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "CreatePatientCondition")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	server.Encode(w, http.StatusCreated, dataCondToCondDisplay(cond))
}

func (m medicalServer) CreatePrescription(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := server.Decode[api.NewPrescription](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	presc, err := m.db.CreatePrescription(r.Context(), newPrescToDataPresc(req))
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "CreatePrescription")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	var prescAppt *appointmentapi.Appointment = nil
	if presc.AppointmentId != nil {
		appt, err := m.apptApi.AppointmentByIdWithResponse(r.Context(), *presc.AppointmentId)
		if err != nil {
			slog.Error(server.UnexpectedError, "error", err.Error(), "where", "CreatePrescription")
			server.EncodeError(w, server.InternalServerError())
			return
		}

		prescAppt = appt.JSON200
	}

	server.Encode(w, http.StatusCreated, dataPrescToPresc(presc, prescAppt))
}

// DeletePrescription implements api.ServerInterface.
func (m medicalServer) DeletePrescription(
	w http.ResponseWriter,
	r *http.Request,
	prescriptionId api.PrescriptionId,
) {
	err := m.db.DeletePrescription(r.Context(), prescriptionId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			server.EncodeError(w, server.NotFoundId("Prescription", prescriptionId))
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
		server.EncodeError(w, server.InternalServerError())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PrescriptionDetail implements api.ServerInterface.
func (m medicalServer) PrescriptionDetail(
	w http.ResponseWriter,
	r *http.Request,
	prescriptionId api.PrescriptionId,
) {
	prescription, err := m.db.PrescriptionById(r.Context(), prescriptionId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			server.EncodeError(w, server.NotFoundId("Prescription", prescriptionId))
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
		server.EncodeError(w, server.InternalServerError())
		return
	}

	var prescAppt *appointmentapi.Appointment = nil
	if prescription.AppointmentId != nil {
		appt, err := m.apptApi.AppointmentByIdWithResponse(r.Context(), *prescription.AppointmentId)
		if err != nil {
			slog.Error(server.UnexpectedError, "error", err.Error(), "where", "PrescriptionDetail")
			server.EncodeError(w, server.InternalServerError())
			return
		}

		prescAppt = appt.JSON200
	}

	server.Encode(w, http.StatusOK, dataPrescToPresc(prescription, prescAppt))
}

// PrescriptionsInDateRange implements api.ServerInterface.
func (m medicalServer) PrescriptionsInDateRange(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
	params api.PrescriptionsInDateRangeParams,
) {
	var to *time.Time = nil
	if params.To != nil {
		to = &params.To.Time
	}

	prescriptions, err := m.db.FindPrescriptionsByPatientId(
		r.Context(),
		patientId,
		params.From.Time,
		to,
	)
	if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"PrescriptionsInDateRange",
			"patientId",
			patientId.String(),
			"from",
			params.From,
			"to",
			params.To,
		)
		server.EncodeError(w, server.InternalServerError())
		return
	}

	server.Encode(
		w,
		http.StatusOK,
		api.Prescriptions{Prescriptions: server.Map(prescriptions, dataPrescToPrescDisplay)},
	)
}

// UpdateCondition implements api.ServerInterface.
func (m medicalServer) UpdateCondition(
	w http.ResponseWriter,
	r *http.Request,
	conditionId api.ConditionId,
) {
	req, decodeErr := server.Decode[api.UpdateCondition](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	existingCondition, err := m.db.ConditionById(r.Context(), conditionId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			server.EncodeError(w, server.NotFoundId("Condition", conditionId))
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
		server.EncodeError(w, server.InternalServerError())
		return
	}

	updated := false
	if req.End != nil {
		if req.End.IsNull() && existingCondition.End != nil {
			existingCondition.End = nil
			updated = true
		} else if req.End.IsSpecified() && (existingCondition.End == nil || !existingCondition.End.Equal(req.End.MustGet())) {
			existingCondition.End = server.AsPtr(req.End.MustGet())
			updated = true
		}
	}
	if req.Name != nil {
		if existingCondition.Name != *req.Name {
			existingCondition.Name = *req.Name
			updated = true
		}
	}
	if req.PatientId != nil {
		if existingCondition.PatientId != *req.PatientId {
			existingCondition.PatientId = *req.PatientId
			updated = true
		}
	}
	if req.Start != nil {
		if !existingCondition.Start.Equal(*req.Start) {
			existingCondition.Start = *req.Start
			updated = true
		}
	}

	var finalConditionData Condition
	if updated {
		updatedDbResult, err := m.db.UpdateCondition(
			r.Context(),
			conditionId,
			existingCondition,
		)
		if err != nil {
			slog.Error(server.UnexpectedError, "error", err.Error(),
				"where",
				"UpdateCondition update",
				"conditionId",
				conditionId.String(),
			)
			server.EncodeError(w, server.InternalServerError())
			return
		}
		finalConditionData = updatedDbResult
	} else {
		finalConditionData = existingCondition
	}

	res, err := m.apptApi.AppointmentsByConditionIdWithResponse(r.Context(), finalConditionData.Id)
	if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"ConditionDetail fetch appts",
		)
		server.EncodeError(w, server.InternalServerError())
		return
	} else if res.JSON200 == nil {
		slog.Error("no appointments response")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	appts := []appointmentapi.AppointmentDisplay{}
	if res.JSON200.Appointments != nil {
		appts = *res.JSON200.Appointments
	}

	server.Encode(w, http.StatusOK, dataCondToCond(finalConditionData, appts))
}

// UpdatePrescription implements api.ServerInterface.
func (m medicalServer) UpdatePrescription(
	w http.ResponseWriter,
	r *http.Request,
	prescriptionId api.PrescriptionId,
) {
	req, decodeErr := server.Decode[api.UpdatePrescription](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	existingPrescription, err := m.db.PrescriptionById(r.Context(), prescriptionId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			server.EncodeError(w, server.NotFoundId("Prescription", prescriptionId))
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
		server.EncodeError(w, server.InternalServerError())
		return
	}

	updated := false
	if req.AppointmentId != nil {
		if existingPrescription.AppointmentId == nil ||
			*existingPrescription.AppointmentId != *req.AppointmentId {
			existingPrescription.AppointmentId = req.AppointmentId
			updated = true
		}
	}
	if req.DoctorsNote != nil {
		if existingPrescription.DoctorsNote == nil ||
			*existingPrescription.DoctorsNote != *req.DoctorsNote {
			existingPrescription.DoctorsNote = req.DoctorsNote
			updated = true
		}
	}
	if req.End != nil {
		if !existingPrescription.End.Equal(*req.End) {
			existingPrescription.End = *req.End
			updated = true
		}
	}
	if req.Name != nil {
		if existingPrescription.Name != *req.Name {
			existingPrescription.Name = *req.Name
			updated = true
		}
	}
	if req.PatientId != nil {
		if existingPrescription.PatientId != *req.PatientId {
			existingPrescription.PatientId = *req.PatientId
			updated = true
		}
	}
	if req.Start != nil {
		if !existingPrescription.Start.Equal(*req.Start) {
			existingPrescription.Start = *req.Start
			updated = true
		}
	}

	var updatedDbPrescription Prescription
	if updated {
		updatedDbPrescription, err = m.db.UpdatePrescription(
			r.Context(),
			prescriptionId,
			existingPrescription,
		)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				server.EncodeError(w, server.NotFoundId("Prescription", prescriptionId))
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
			server.EncodeError(w, server.InternalServerError())
			return
		}
	} else {
		updatedDbPrescription = existingPrescription
	}

	var prescAppt *appointmentapi.Appointment = nil
	if updatedDbPrescription.AppointmentId != nil {
		appt, err := m.apptApi.AppointmentByIdWithResponse(
			r.Context(),
			*updatedDbPrescription.AppointmentId,
		)
		if err != nil {
			slog.Error(server.UnexpectedError, "error", err.Error(), "where", "CreatePrescription")
			server.EncodeError(w, server.InternalServerError())
			return
		}

		prescAppt = appt.JSON200
	}

	server.Encode(w, http.StatusOK, dataPrescToPresc(updatedDbPrescription, prescAppt))
}
