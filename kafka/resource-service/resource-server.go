package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/IBM/sarama"
	"github.com/go-chi/httplog/v2"
	"github.com/google/uuid"

	"github.com/Nesquiko/aass/common/server"
	commonapi "github.com/Nesquiko/aass/common/server/api"
	"github.com/Nesquiko/aass/resource-service/api"
	appointmentapi "github.com/Nesquiko/aass/resource-service/appointment-api"
)

type resourceServer struct {
	db            mongoResourcesDb
	kafka         sarama.Client
	kafkaProducer sarama.SyncProducer
}

const ResourceReservedTopic = "resource-reserved"

func newResourceServer(
	db mongoResourcesDb,
	logger *httplog.Logger,
	opts commonapi.ChiServerOptions,
) http.Handler {
	kafkaClient, err := server.InitKafka(ResourceReservedTopic)
	if err != nil {
		slog.Error("Error while creating kafka client", "error", err.Error())
		os.Exit(1)
	}

	kafkaProducer, err := sarama.NewSyncProducerFromClient(kafkaClient)
	if err != nil {
		kafkaClient.Close()
		slog.Error("Error creating Kafka sync producer", "error", err)
		os.Exit(1)
	}

	srv := resourceServer{
		db:            db,
		kafka:         kafkaClient,
		kafkaProducer: kafkaProducer,
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

	go server.NewConsumer(
		kafkaClient,
		srv.appointmentScheduledConsumer,
		[]string{"appointment-scheduled"},
	)

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
	err := s.reserveAppointmentResources(ctx, appointmentId, req)
	if err != nil {
		server.EncodeError(w, handlErr(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s resourceServer) reserveAppointmentResources(
	ctx context.Context,
	appointmentId api.AppointmentId,
	req api.ReserveAppointmentResourcesJSONBody,
) error {
	reservationStart := req.Start
	// Assuming 1 hour duration for reservation based on appointment start time
	reservationEnd := reservationStart.Add(time.Hour)

	// --- Reserve Equipment ---
	if req.EquipmentId != nil {
		resource, err := s.db.ResourceById(ctx, *req.EquipmentId)
		if err != nil {
			slog.Error(
				"error finding equipment resource",
				"error", err.Error(), "where", "ReserveAppointmentResources",
				"appointmentId", appointmentId.String(), "equipmentId", req.EquipmentId.String(),
			)
			return err
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
			return err
		}
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
			return err
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
			return err
		}
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
			return err
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
			return err
		}
	}

	return nil
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

func (s resourceServer) appointmentScheduledConsumer(value []byte) {
	ctx := context.Background()
	var appt appointmentapi.Appointment
	dec := json.NewDecoder(bytes.NewReader(value))
	err := dec.Decode(&appt)
	if err != nil {
		slog.Error("Consumer error decoding value", "error", err.Error())
		return
	}

	req := api.ReserveAppointmentResourcesJSONBody{
		Start: appt.AppointmentDateTime,
	}
	if appt.Equipment != nil && len(*appt.Equipment) != 0 {
		req.EquipmentId = &((*appt.Equipment)[0].Id)
	}
	if appt.Medicine != nil && len(*appt.Medicine) != 0 {
		req.MedicineId = &((*appt.Medicine)[0].Id)
	}
	if appt.Facilities != nil && len(*appt.Facilities) != 0 {
		req.FacilityId = &((*appt.Facilities)[0].Id)
	}

	err = s.reserveAppointmentResources(context.Background(), appt.Id, req)
	if err != nil {
		slog.Error("Consumer error reserving resources", "error", err.Error())
		return
	}

	reserved := ReservedResources{
		AppointmentId: appt.Id,
	}

	if req.EquipmentId != nil {
		res, _ := s.db.ResourceById(ctx, *req.EquipmentId)
		reserved.Equipment = &ReservedResource{
			Id:   res.Id,
			Name: res.Name,
			Type: res.Type,
		}
	}

	if req.MedicineId != nil {
		res, _ := s.db.ResourceById(ctx, *req.MedicineId)
		reserved.Medicine = &ReservedResource{
			Id:   res.Id,
			Name: res.Name,
			Type: res.Type,
		}
	}

	if req.FacilityId != nil {
		res, _ := s.db.ResourceById(ctx, *req.FacilityId)
		reserved.Facility = &ReservedResource{
			Id:   res.Id,
			Name: res.Name,
			Type: res.Type,
		}
	}

	slog.Info("Emitting reserved event", "reserved", reserved)
	eventValue, marshalErr := mapResourcesToKafkaMessageValue(reserved)
	if marshalErr != nil {
		slog.Error("Failed to marshal appointment event", "error", marshalErr)
	} else {
		msg := &sarama.ProducerMessage{
			Topic: ResourceReservedTopic,
			Key:   sarama.StringEncoder(appt.Id.String()),
			Value: sarama.ByteEncoder(eventValue),
		}

		partition, offset, sendErr := s.kafkaProducer.SendMessage(msg)
		if sendErr != nil {
			slog.Error("Failed to send reserved resources event to Kafka", "error", sendErr)
		} else {
			slog.Info("Successfully sent appointment scheduled event to Kafka",
				"topic", ResourceReservedTopic,
				"partition", partition,
				"offset", offset,
			)
		}
	}
}

type ReservedResource struct {
	Id   uuid.UUID    `json:"id"`
	Name string       `json:"name"`
	Type ResourceType `json:"type"`
}

type ReservedResources struct {
	AppointmentId uuid.UUID `json:"appointmentId"`

	Equipment *ReservedResource `json:"equipment,omitempty"`
	Facility  *ReservedResource `json:"facility,omitempty"`
	Medicine  *ReservedResource `json:"medicine,omitempty"`
}

func mapResourcesToKafkaMessageValue(re ReservedResources) ([]byte, error) {
	value, err := json.Marshal(re)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal appointment to JSON: %w", err)
	}
	return value, nil
}
