package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/httplog/v2"

	"github.com/Nesquiko/aass/appointment-service/api"
	medicalapi "github.com/Nesquiko/aass/appointment-service/medical-api"
	resourceapi "github.com/Nesquiko/aass/appointment-service/resources-api"
	userapi "github.com/Nesquiko/aass/appointment-service/user-api"
	"github.com/Nesquiko/aass/common/server"
	commonapi "github.com/Nesquiko/aass/common/server/api"
)

type appointmentServer struct {
	db          mongoAppointmentDb
	medicalApi  *medicalapi.ClientWithResponses
	resourceApi *resourceapi.ClientWithResponses
	userApi     *userapi.ClientWithResponses
}

func newAppointmentServer(
	db mongoAppointmentDb,
	logger *httplog.Logger,
	opts commonapi.ChiServerOptions,
) http.Handler {
	medicalClient, _ := medicalapi.NewClientWithResponses("http://medical-service:8080/")
	resourceClient, _ := resourceapi.NewClientWithResponses("http://resource-service:8080/")
	userClient, _ := userapi.NewClientWithResponses("http://user-service:8080/")
	srv := appointmentServer{
		db:          db,
		medicalApi:  medicalClient,
		resourceApi: resourceClient,
		userApi:     userClient,
	}

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

// AppointmentById implements api.ServerInterface.
func (a appointmentServer) AppointmentById(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	ctx := r.Context()
	apptData, err := a.db.AppointmentById(ctx, appointmentId)
	if errors.Is(err, ErrNotFound) {
		server.EncodeError(w, server.NotFoundId("Appointment", appointmentId))
		return
	} else if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "AppointmentById db")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	apiAppt, apiErr := a.mapDataApptToApiAppt(ctx, apptData)
	if apiErr != nil {
		server.EncodeError(w, apiErr)
		return
	}

	server.Encode(w, http.StatusOK, apiAppt)
}

// AppointmentsByConditionId implements api.ServerInterface.
func (a appointmentServer) AppointmentsByConditionId(
	w http.ResponseWriter,
	r *http.Request,
	conditionId api.ConditionId,
) {
	ctx := r.Context()
	apptsData, err := a.db.AppointmentsByConditionId(ctx, conditionId)
	if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"AppointmentsByConditionId db",
		)
		server.EncodeError(w, server.InternalServerError())
		return
	}

	apiAppts := make([]api.AppointmentDisplay, 0, len(apptsData))
	for _, apptData := range apptsData {
		apiApptDisplay, apiErr := a.mapDataApptToApiApptDisplay(ctx, apptData)
		if apiErr != nil {
			server.EncodeError(w, apiErr)
			return
		}
		apiAppts = append(apiAppts, apiApptDisplay)
	}

	server.Encode(w, http.StatusOK, api.Appointments{Appointments: &apiAppts})
}

// CancelAppointment implements api.ServerInterface.
func (a appointmentServer) CancelAppointment(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	ctx := r.Context()
	req, decodeErr := server.Decode[api.AppointmentCancellation](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	err := a.db.CancelAppointment(ctx, appointmentId, string(req.By), req.Reason)
	if errors.Is(err, ErrNotFound) {
		server.EncodeError(w, server.NotFoundId("Appointment", appointmentId))
		return
	} else if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"CancelAppointment db",
		)
		server.EncodeError(w, server.InternalServerError())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DecideAppointment implements api.ServerInterface.
func (a appointmentServer) DecideAppointment(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	ctx := r.Context()
	req, decodeErr := server.Decode[api.AppointmentDecision](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	apptData, err := a.db.AppointmentById(ctx, appointmentId)
	if errors.Is(err, ErrNotFound) {
		server.EncodeError(w, server.NotFoundId("Appointment", appointmentId))
		return
	} else if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"DecideAppointment get appt db",
		)
		server.EncodeError(w, server.InternalServerError())
		return
	}

	var updatedApptData Appointment

	if req.Action == api.Accept {
		reserveReqBody := resourceapi.ReserveAppointmentResourcesJSONRequestBody{
			Start:       apptData.AppointmentDateTime,
			FacilityId:  req.Facility,
			EquipmentId: req.Equipment,
			MedicineId:  req.Medicine,
		}
		resResp, resErr := a.resourceApi.ReserveAppointmentResourcesWithResponse(
			ctx,
			appointmentId,
			reserveReqBody,
		)
		if resErr != nil {
			slog.Error(
				server.UnexpectedError,
				"error",
				resErr.Error(),
				"where",
				"DecideAppointment reserve resources api call",
			)
			server.EncodeError(w, server.InternalServerError())
			return
		}
		if resResp.StatusCode() != http.StatusNoContent {
			slog.Error(
				"resource reservation failed",
				"status",
				resResp.StatusCode(),
				"body",
				string(resResp.Body),
			)
			if resResp.ApplicationproblemJSON404 != nil {
				server.EncodeError(
					w,
					&server.ApiError{ErrorDetail: *resResp.ApplicationproblemJSON404},
				)
				return
			}
			server.EncodeError(w, server.InternalServerError())
			return
		}

		resources := make([]Resource, 0)
		if req.Facility != nil {
			resources = append(resources, Resource{Id: *req.Facility, Type: ResourceTypeFacility})
		}
		if req.Equipment != nil {
			resources = append(resources, Resource{Id: *req.Equipment, Type: ResourceTypeMedicine})
		}
		if req.Medicine != nil {
			resources = append(resources, Resource{Id: *req.Medicine, Type: ResourceTypeMedicine})
		}

		updatedApptData, err = a.db.DecideAppointment(
			ctx,
			appointmentId,
			string(req.Action),
			nil,
			resources,
		)
	} else if req.Action == api.Reject {
		if req.Reason == nil || *req.Reason == "" {
			server.EncodeError(w, &server.ApiError{
				ErrorDetail: commonapi.ErrorDetail{
					Code:   "validation.required",
					Title:  "Validation Error",
					Detail: "Reason is required when rejecting an appointment",
					Status: http.StatusBadRequest,
				},
			})
			return
		}
		updatedApptData, err = a.db.DecideAppointment(
			ctx,
			appointmentId,
			string(req.Action),
			req.Reason,
			nil,
		)
	} else {
		server.EncodeError(w, &server.ApiError{
			ErrorDetail: commonapi.ErrorDetail{
				Code:   "validation.invalid.action",
				Title:  "Validation Error",
				Detail: fmt.Sprintf("Invalid action: %s", req.Action),
				Status: http.StatusBadRequest,
			},
		})
		return
	}

	if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"DecideAppointment update db",
		)
		server.EncodeError(w, server.InternalServerError())
		return
	}

	apiAppt, apiErr := a.mapDataApptToApiAppt(ctx, updatedApptData)
	if apiErr != nil {
		server.EncodeError(w, apiErr)
		return
	}

	server.Encode(w, http.StatusOK, apiAppt)
}

// DoctorsCalendar implements api.ServerInterface.
func (a appointmentServer) DoctorsCalendar(
	w http.ResponseWriter,
	r *http.Request,
	doctorId api.DoctorId,
	params api.DoctorsCalendarParams,
) {
	ctx := r.Context()
	var to *time.Time = nil
	if params.To != nil {
		to = &params.To.Time
	}

	apptsData, err := a.db.AppointmentsByDoctorId(ctx, doctorId, params.From.Time, to)
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "DoctorsCalendar db")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	apiAppts := make([]api.AppointmentDisplay, 0, len(apptsData))
	for _, apptData := range apptsData {
		apiApptDisplay, apiErr := a.mapDataApptToApiApptDisplay(ctx, apptData)
		if apiErr != nil {
			server.EncodeError(w, apiErr)
			return
		}
		apiAppts = append(apiAppts, apiApptDisplay)
	}

	server.Encode(w, http.StatusOK, api.Appointments{Appointments: &apiAppts})
}

// DoctorsTimeslots implements api.ServerInterface.
func (a appointmentServer) DoctorsTimeslots(
	w http.ResponseWriter,
	r *http.Request,
	doctorId api.DoctorId,
	params api.DoctorsTimeslotsParams,
) {
	ctx := r.Context()
	date := params.Date.Time

	apptsData, err := a.db.AppointmentsByDoctorIdAndDate(ctx, doctorId, date)
	if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"DoctorsTimeslots db",
		)
		server.EncodeError(w, server.InternalServerError())
		return
	}

	loc := date.Location()
	startHour, endHour := 8, 14
	slotDuration := time.Hour

	allSlots := make([]api.TimeSlot, 0)
	currentTime := time.Date(date.Year(), date.Month(), date.Day(), startHour, 0, 0, 0, loc)
	endTimeLimit := time.Date(date.Year(), date.Month(), date.Day(), endHour, 0, 0, 0, loc)

	bookedTimes := make(map[time.Time]bool)
	for _, appt := range apptsData {
		if appt.Status != string(api.Cancelled) && appt.Status != string(api.Denied) {
			bookedTimes[appt.AppointmentDateTime] = true
		}
	}

	for currentTime.Before(endTimeLimit) {
		status := api.Available
		if bookedTimes[currentTime] {
			status = api.Unavailable
		}
		allSlots = append(allSlots, api.TimeSlot{
			Time:   currentTime.Format("15:04"),
			Status: status,
		})
		currentTime = currentTime.Add(slotDuration)
	}

	server.Encode(w, http.StatusOK, api.DoctorTimeslots{Slots: allSlots})
}

// PatientsCalendar implements api.ServerInterface.
func (a appointmentServer) PatientsCalendar(
	w http.ResponseWriter,
	r *http.Request,
	patientId api.PatientId,
	params api.PatientsCalendarParams,
) {
	ctx := r.Context()
	var to *time.Time = nil
	if params.To != nil {
		to = &params.To.Time
	}

	apptsData, err := a.db.AppointmentsByPatientId(ctx, patientId, params.From.Time, to)
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "PatientsCalendar db")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	apiAppts := make([]api.AppointmentDisplay, 0, len(apptsData))
	for _, apptData := range apptsData {
		apiApptDisplay, apiErr := a.mapDataApptToApiApptDisplay(ctx, apptData)
		if apiErr != nil {
			server.EncodeError(w, apiErr)
			return
		}
		apiAppts = append(apiAppts, apiApptDisplay)
	}

	server.Encode(w, http.StatusOK, api.Appointments{Appointments: &apiAppts})
}

// RequestAppointment implements api.ServerInterface.
func (a appointmentServer) RequestAppointment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, decodeErr := server.Decode[api.NewAppointmentRequest](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	res, patientErr := a.userApi.GetPatientByIdWithResponse(ctx, req.PatientId)
	if patientErr != nil || res.StatusCode() != http.StatusOK {
		slog.Error("failed to get patient", "error", patientErr, "status", res.StatusCode())
		if patientErr != nil && res.ApplicationproblemJSON404 != nil {
			server.EncodeError(w, server.NotFoundId("Patient", req.PatientId))
			return
		}
		server.EncodeError(w, server.InternalServerError())
		return
	}

	res2, doctorErr := a.userApi.GetDoctorByIdWithResponse(ctx, req.DoctorId)
	if doctorErr != nil || res2.StatusCode() != http.StatusOK {
		slog.Error("failed to get doctor", "error", doctorErr, "status", res2.StatusCode())
		if doctorErr != nil && res2.ApplicationproblemJSON404 != nil {
			server.EncodeError(w, server.NotFoundId("Doctor", req.DoctorId))
			return
		}
		server.EncodeError(w, server.InternalServerError())
		return
	}

	if req.ConditionId != nil {
		res3, conditionErr := a.medicalApi.ConditionDetailWithResponse(ctx, *req.ConditionId)
		if conditionErr != nil || res3.StatusCode() != http.StatusOK {
			slog.Error(
				"failed to get condition",
				"error",
				conditionErr,
				"status",
				res3.StatusCode(),
			)
			if conditionErr != nil && res3.JSON200 == nil {
				server.EncodeError(w, server.NotFoundId("Condition", *req.ConditionId))
				return
			}
			server.EncodeError(w, server.InternalServerError())
			return
		}
	}

	apptType := api.Consultation // Default type
	if req.Type != nil {
		apptType = *req.Type
	}

	endTime := req.AppointmentDateTime.Add(time.Hour)

	newApptData := Appointment{
		PatientId:           req.PatientId,
		DoctorId:            req.DoctorId,
		AppointmentDateTime: req.AppointmentDateTime,
		EndTime:             endTime,
		Type:                string(apptType),
		Status:              string(api.Requested),
		Reason:              req.Reason,
		ConditionId:         req.ConditionId,
	}

	createdApptData, err := a.db.CreateAppointment(ctx, newApptData)
	if errors.Is(err, ErrDoctorUnavailable) {
		server.EncodeError(w, &server.ApiError{
			ErrorDetail: commonapi.ErrorDetail{
				Code:   "doctor.unavailable",
				Title:  "Conflict",
				Detail: err.Error(),
				Status: http.StatusConflict,
			},
		})
		return
	} else if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"RequestAppointment db",
		)
		server.EncodeError(w, server.InternalServerError())
		return
	}

	apiAppt, apiErr := a.mapDataApptToApiAppt(ctx, createdApptData)
	if apiErr != nil {
		server.EncodeError(w, apiErr)
		return
	}

	server.Encode(w, http.StatusCreated, apiAppt)
}

// RescheduleAppointment implements api.ServerInterface.
func (a appointmentServer) RescheduleAppointment(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	ctx := r.Context()
	req, decodeErr := server.Decode[api.AppointmentReschedule](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	updatedApptData, err := a.db.RescheduleAppointment(
		ctx,
		appointmentId,
		req.NewAppointmentDateTime,
	)
	if errors.Is(err, ErrNotFound) {
		server.EncodeError(w, server.NotFoundId("Appointment", appointmentId))
		return
	} else if errors.Is(err, ErrDoctorUnavailable) {
		server.EncodeError(w, &server.ApiError{
			ErrorDetail: commonapi.ErrorDetail{
				Code:   "doctor.unavailable",
				Title:  "Conflict",
				Detail: err.Error(),
				Status: http.StatusConflict,
			},
		})
		return
	} else if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"RescheduleAppointment db",
		)
		server.EncodeError(w, server.InternalServerError())
		return
	}

	apiAppt, apiErr := a.mapDataApptToApiAppt(ctx, updatedApptData)
	if apiErr != nil {
		server.EncodeError(w, apiErr)
		return
	}

	server.Encode(w, http.StatusOK, apiAppt)
}

// UpdateAppointmentResources implements api.ServerInterface.
func (a appointmentServer) UpdateAppointmentResources(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	ctx := r.Context()
	req, decodeErr := server.Decode[api.AppointmentResourceUpdate](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	// 1. Fetch current appointment data
	currentApptData, err := a.db.AppointmentById(ctx, appointmentId)
	if errors.Is(err, ErrNotFound) {
		server.EncodeError(w, server.NotFoundId("Appointment", appointmentId))
		return
	} else if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"UpdateAppointmentResources get current appt",
		)
		server.EncodeError(w, server.InternalServerError())
		return
	}

	var finalFacilities []Resource
	var finalEquipment []Resource
	var finalMedicine []Resource

	// Facility
	if req.FacilityId != nil {
		if !req.FacilityId.IsNull() {
			id := req.FacilityId.MustGet()
			finalFacilities = []Resource{{Id: id, Type: ResourceTypeFacility}}
		} else {
			finalFacilities = []Resource{}
		}
	} else {
		finalFacilities = currentApptData.Facilities
	}

	// Equipment
	if req.EquipmentId != nil {
		if !req.EquipmentId.IsNull() {
			id := req.EquipmentId.MustGet()
			finalEquipment = []Resource{{Id: id, Type: ResourceTypeEquipment}}
		} else {
			finalEquipment = []Resource{}
		}
	} else {
		finalEquipment = currentApptData.Equipment
	}

	// Medicine
	if req.MedicineId != nil {
		if !req.MedicineId.IsNull() {
			id := req.MedicineId.MustGet()
			finalMedicine = []Resource{{Id: id, Type: ResourceTypeMedicine}}
		} else {
			finalMedicine = []Resource{}
		}
	} else {
		finalMedicine = currentApptData.Medicines
	}

	// 3. Update the appointment in the database
	updatedApptData, err := a.db.UpdateAppointmentResources(
		ctx,
		appointmentId,
		finalFacilities,
		finalEquipment,
		finalMedicine,
	)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			server.EncodeError(w, server.NotFoundId("Appointment", appointmentId))
			return
		}
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"UpdateAppointmentResources db update",
		)
		server.EncodeError(w, server.InternalServerError())
		return
	}

	apiAppt, apiErr := a.mapDataApptToApiAppt(ctx, updatedApptData)
	if apiErr != nil {
		server.EncodeError(w, apiErr)
		return
	}

	server.Encode(w, http.StatusOK, apiAppt)
}
