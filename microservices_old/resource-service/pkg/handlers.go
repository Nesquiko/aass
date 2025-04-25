package pkg

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Nesquiko/aass/common/server"
	"github.com/Nesquiko/aass/resource-service/api"
)

// CreateResource implements api.ServerInterface.
func (s ResourceServer) CreateResource(w http.ResponseWriter, r *http.Request) {
	req, decodeErr := Decode[api.NewResource](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	resource, err := s.app.CreateResource(r.Context(), req)
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "CreateResource")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusCreated, resource)
}

// DeleteAppointmentReservations implements api.ServerInterface.
func (ResourceServer) DeleteAppointmentReservations(
	w http.ResponseWriter,
	r *http.Request,
	params api.DeleteAppointmentReservationsParams,
) {
	panic("unimplemented not used on FE")
}

// GetAvailableResources implements api.ServerInterface.
func (s ResourceServer) GetAvailableResources(
	w http.ResponseWriter,
	r *http.Request,
	params api.GetAvailableResourcesParams,
) {
	resources, err := s.app.AvailableResources(r.Context(), params.DateTime)
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "GetAvailableResources")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, resources)
}

// GetResourceById implements api.ServerInterface.
func (s ResourceServer) GetResourceById(
	w http.ResponseWriter,
	r *http.Request,
	resourceId api.ResourceId,
) {
	resource, err := s.app.db.ResourceById(r.Context(), resourceId)
	if err != nil {
		slog.Error(server.UnexpectedError, "error", err.Error(), "where", "GetResourceById")
		encodeError(w, internalServerError())
		return
	}

	encode(w, http.StatusOK, resource)
}

// ReserveAppointmentResources implements api.ServerInterface.
func (s ResourceServer) ReserveAppointmentResources(
	w http.ResponseWriter,
	r *http.Request,
	appointmentId api.AppointmentId,
) {
	req, decodeErr := Decode[api.ReserveAppointmentResourcesJSONBody](w, r)
	if decodeErr != nil {
		encodeError(w, decodeErr)
		return
	}

	err := s.app.ReserveAppointmentResources(
		r.Context(),
		appointmentId,
		req,
	)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
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
		} else if errors.Is(err, ErrResourceUnavailable) {
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
			server.UnexpectedError,
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

	w.WriteHeader(http.StatusNoContent)
}
