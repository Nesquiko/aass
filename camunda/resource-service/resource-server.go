package main

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/httplog/v2"
	"github.com/google/uuid"

	"github.com/Nesquiko/aass/common/server"
	commonapi "github.com/Nesquiko/aass/common/server/api"
	"github.com/Nesquiko/aass/resource-service/api"
	appointmentapi "github.com/Nesquiko/aass/resource-service/appointment-api"
)

type resourceServer struct {
	db             mongoResourcesDb
	appointmentApi *appointmentapi.ClientWithResponses
}

func newResourceServer(
	db mongoResourcesDb,
	logger *httplog.Logger,
	opts commonapi.ChiServerOptions,
) http.Handler {
	apptClient, _ := appointmentapi.NewClientWithResponses("http://appointment-service:8080/")
	srv := resourceServer{db: db, appointmentApi: apptClient}

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

func (s resourceServer) CreateResource(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := server.Decode[api.NewResource](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	resource, err := s.db.CreateResource(r.Context(), req.Name, ResourceType(req.Type))
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "CreateResource")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	res := api.NewResource{
		Id:   &resource.Id,
		Name: resource.Name,
		Type: api.ResourceType(resource.Type),
	}

	server.Encode(w, http.StatusCreated, res)
}

func (s resourceServer) GetAvailableResources(
	w http.ResponseWriter,
	r *http.Request,
	params api.GetAvailableResourcesParams,
) {
	resources, err := s.db.FindAvailableResourcesAtTime(r.Context(), params.DateTime)
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "GetAvailableResources")
		server.EncodeError(w, server.InternalServerError())
		return
	}

	server.Encode(w, http.StatusOK, dataResourcesToApiResources(resources))
}

func (s resourceServer) ReserveAppointmentResources(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	req, decodeErr := server.Decode[api.ReserveAppointmentResourcesJSONBody](w, r)
	if decodeErr != nil {
		server.EncodeError(w, decodeErr)
		return
	}

	ctx := r.Context()
	reservationStart := req.Start
	// Assuming 1 hour duration for reservation based on appointment start time
	reservationEnd := reservationStart.Add(time.Hour)

	// Keep track of successfully reserved resource IDs
	var reservedFacilityId *uuid.UUID
	var reservedEquipmentId *uuid.UUID
	var reservedMedicineId *uuid.UUID

	// --- Reserve Equipment ---
	if req.EquipmentId != nil {
		resource, err := s.db.ResourceById(ctx, *req.EquipmentId)
		if err != nil {
			slog.Error(
				"error finding equipment resource",
				"error", err.Error(), "where", "ReserveAppointmentResources",
				"appointmentId", appointmentId.String(), "equipmentId", req.EquipmentId.String(),
			)
			server.EncodeError(w, handlErr(err))
			return
		}

		_, err = s.db.CreateReservation(
			ctx, appointmentId, *req.EquipmentId, resource.Name,
			resource.Type, reservationStart, reservationEnd,
		)
		if err != nil {
			slog.Error(
				"error creating equipment reservation",
				"error", err.Error(), "where", "ReserveAppointmentResources",
				"appointmentId", appointmentId.String(), "equipmentId", req.EquipmentId.String(),
			)
			server.EncodeError(w, handlErr(err)) // Handles ErrResourceUnavailable
			return
		}
		reservedEquipmentId = req.EquipmentId // Mark as successfully reserved
	}

	// --- Reserve Facility ---
	if req.FacilityId != nil {
		resource, err := s.db.ResourceById(ctx, *req.FacilityId)
		if err != nil {
			slog.Error(
				"error finding facility resource",
				"error", err.Error(), "where", "ReserveAppointmentResources",
				"appointmentId", appointmentId.String(), "facilityId", req.FacilityId.String(),
			)
			server.EncodeError(w, handlErr(err))
			return
		}

		_, err = s.db.CreateReservation(
			ctx, appointmentId, *req.FacilityId, resource.Name,
			resource.Type, reservationStart, reservationEnd,
		)
		if err != nil {
			slog.Error(
				"error creating facility reservation",
				"error", err.Error(), "where", "ReserveAppointmentResources",
				"appointmentId", appointmentId.String(), "facilityId", req.FacilityId.String(),
			)
			server.EncodeError(w, handlErr(err))
			return
		}
		reservedFacilityId = req.FacilityId // Mark as successfully reserved
	}

	// --- Reserve Medicine ---
	if req.MedicineId != nil {
		resource, err := s.db.ResourceById(ctx, *req.MedicineId)
		if err != nil {
			slog.Error(
				"error finding medicine resource",
				"error", err.Error(), "where", "ReserveAppointmentResources",
				"appointmentId", appointmentId.String(), "medicineId", req.MedicineId.String(),
			)
			server.EncodeError(w, handlErr(err))
			return
		}
		_, err = s.db.CreateReservation(
			ctx, appointmentId, *req.MedicineId, resource.Name,
			resource.Type, reservationStart, reservationEnd,
		)
		if err != nil {
			slog.Error(
				"error creating medicine reservation",
				"error", err.Error(), "where", "ReserveAppointmentResources",
				"appointmentId", appointmentId.String(), "medicineId", req.MedicineId.String(),
			)
			server.EncodeError(w, handlErr(err))
			return
		}
		reservedMedicineId = req.MedicineId // Mark as successfully reserved
	}

	// --- Update Appointment Service ---
	// Only call if at least one resource was requested and presumably reserved
	if req.FacilityId != nil || req.EquipmentId != nil || req.MedicineId != nil {
		updateBody := appointmentapi.AppointmentResourceUpdate{
			FacilityId:  reservedFacilityId,
			EquipmentId: reservedEquipmentId,
			MedicineId:  reservedMedicineId,
		}

		slog.Info(
			"Attempting to update appointment resources",
			"appointmentId", appointmentId.String(),
			"facilityId", reservedFacilityId,
			"equipmentId", reservedEquipmentId,
			"medicineId", reservedMedicineId,
		)

		apptUpdateResp, apptUpdateErr := s.appointmentApi.UpdateAppointmentResourcesWithResponse(
			ctx,
			appointmentId,
			updateBody,
		)

		if apptUpdateErr != nil {
			// Log failure but don't necessarily fail the resource reservation itself
			slog.Error(
				"failed to call update appointment resources endpoint",
				"error", apptUpdateErr.Error(), "where", "ReserveAppointmentResources",
				"appointmentId", appointmentId.String(),
			)
		} else if apptUpdateResp.StatusCode() != http.StatusOK {
			// Log failure but don't necessarily fail the resource reservation itself
			slog.Error(
				"failed to update appointment resources",
				"status", apptUpdateResp.StatusCode(), "body", string(apptUpdateResp.Body),
				"where", "ReserveAppointmentResources", "appointmentId", appointmentId.String(),
			)
		} else {
			slog.Info(
				"Successfully updated appointment resources",
				"appointmentId", appointmentId.String(),
			)
		}
	} else {
		slog.Info(
			"No resources requested for reservation, skipping appointment update call",
			"appointmentId", appointmentId.String(),
		)
	}

	// If we reached here, all requested reservations were successful locally.
	w.WriteHeader(http.StatusNoContent)
}

func handlErr(err error) *server.ApiError {
	if errors.Is(err, ErrNotFound) {
		apiErr := &server.ApiError{
			ErrorDetail: commonapi.ErrorDetail{
				Code:   "resource.or.appointment.not-found",
				Title:  "Not Found",
				Detail: err.Error(),
				Status: http.StatusNotFound,
			},
		}
		return apiErr
	} else if errors.Is(err, ErrResourceUnavailable) {
		apiErr := &server.ApiError{
			ErrorDetail: commonapi.ErrorDetail{
				Code:   "resource.unavailable",
				Title:  "Conflict",
				Detail: err.Error(),
				Status: http.StatusConflict,
			},
		}
		return apiErr
	}
	return server.InternalServerError()
}

func (s resourceServer) GetResourceById(
	w http.ResponseWriter,
	r *http.Request,
	resourceId api.ResourceId,
) {
	resource, err := s.db.ResourceById(r.Context(), resourceId)
	if err != nil {
		server.EncodeError(w, handlErr(err))
		return
	}

	server.Encode(w, http.StatusOK, resource)
}
