package main

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/httplog/v2"

	"github.com/Nesquiko/aass/common/server"
	commonapi "github.com/Nesquiko/aass/common/server/api"
	"github.com/Nesquiko/aass/resource-service/api"
)

type resourceServer struct{ db mongoResourcesDb }

func newResourceServer(
	db mongoResourcesDb,
	logger *httplog.Logger,
	opts commonapi.ChiServerOptions,
) http.Handler {
	srv := resourceServer{db: db}

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
	reservationEnd := reservationStart.Add(time.Hour)

	if req.EquipmentId != nil {
		resource, err := s.db.ResourceById(ctx, *req.EquipmentId)
		if err != nil {
			slog.Error(
				"error in equipment",
				"error",
				err.Error(),
				"where",
				"ReserveAppointmentResources",
				"appointmentId",
				appointmentId.String(),
			)

			server.EncodeError(w, handlErr(err))
			return
		}

		_, err = s.db.CreateReservation(
			ctx,
			appointmentId,
			*req.EquipmentId,
			resource.Name,
			resource.Type,
			reservationStart,
			reservationEnd,
		)
		if err != nil {
			server.EncodeError(w, handlErr(err))
			return
		}
	}

	if req.FacilityId != nil {
		resource, err := s.db.ResourceById(ctx, *req.FacilityId)
		if err != nil {
			slog.Error(
				"error in facility",
				"error",
				err.Error(),
				"where",
				"ReserveAppointmentResources",
				"appointmentId",
				appointmentId.String(),
			)

			server.EncodeError(w, handlErr(err))
			return
		}

		_, err = s.db.CreateReservation(
			ctx,
			appointmentId,
			*req.FacilityId,
			resource.Name,
			resource.Type,
			reservationStart,
			reservationEnd,
		)
		if err != nil {
			server.EncodeError(w, handlErr(err))
			return
		}
	}

	if req.MedicineId != nil {
		resource, err := s.db.ResourceById(ctx, *req.MedicineId)
		if err != nil {
			slog.Error(
				"error in medicine",
				"error",
				err.Error(),
				"where",
				"ReserveAppointmentResources",
				"appointmentId",
				appointmentId.String(),
			)

			server.EncodeError(w, handlErr(err))
			return
		}
		_, err = s.db.CreateReservation(
			ctx,
			appointmentId,
			*req.MedicineId,
			resource.Name,
			resource.Type,
			reservationStart,
			reservationEnd,
		)
		if err != nil {
			server.EncodeError(w, handlErr(err))
			return
		}
	}

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
