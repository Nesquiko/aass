package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	camunda_client_go "github.com/citilinkru/camunda-client-go/v3"
	"github.com/citilinkru/camunda-client-go/v3/processor"
	"github.com/google/uuid"

	resourceapi "github.com/Nesquiko/aass/camunda-worker/resources-api"
	"github.com/Nesquiko/aass/common/server"
)

const (
	camundaRestURL = "http://camunda-platform:8080/engine-rest" // Your Camunda REST endpoint URL
	workerID       = "resource-reservation-worker"
	topicName      = "appointment-reserve-resources"

	lockDuration              = 5 * time.Second // How long the task is locked for this worker
	maxTasks                  = 10              // How many tasks to fetch at once
	maxParallelTaskPerHandler = 100             // Max concurrent handler executions
	asyncResponseTimeout      = 5000            // Async timeout for long polling (milliseconds)
	fetchInterval             = 5 * time.Second // How often to poll for new tasks if none are available

	dateTimeFormat = time.RFC3339
)

type ReserveResourcesRequestBody struct {
	AppointmentDateTime time.Time  `json:"appointmentDateTime"`
	AppointmentId       uuid.UUID  `json:"appointmentId"`
	FacilityId          *uuid.UUID `json:"facilityId,omitempty"`
	EquipmentId         *uuid.UUID `json:"equipmentId,omitempty"`
	MedicineId          *uuid.UUID `json:"medicineId,omitempty"`
}

func main() {
	slog.Info("Starting Camunda External Task Worker", "workerId", workerID, "topic", topicName)

	client := camunda_client_go.NewClient(camunda_client_go.ClientOptions{
		EndpointUrl: camundaRestURL,
		ApiUser:     "demo",
		ApiPassword: "demo",
		Timeout:     time.Second * 15,
	})

	proc := processor.NewProcessor(client, &processor.Options{
		WorkerId:                  workerID,
		LockDuration:              lockDuration,
		MaxTasks:                  maxTasks,
		MaxParallelTaskPerHandler: maxParallelTaskPerHandler,
		AsyncResponseTimeout:      server.AsPtr(asyncResponseTimeout),
		LongPollingTimeout:        fetchInterval,
	}, func(err error) {
		slog.Error("Camunda Processor Error", "error", err)
	})

	resourceClient, _ := resourceapi.NewClientWithResponses("http://resource-service:8080/")

	proc.AddHandler(
		[]*camunda_client_go.QueryFetchAndLockTopic{
			{TopicName: topicName, LockDuration: int(lockDuration.Milliseconds())},
		},
		func(ctx *processor.Context) error {
			return handleReserveResources(ctx, resourceClient)
		},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigint
		slog.Info("Received shutdown signal, shutting down worker...")
		cancel()
	}()

	slog.Info("Worker handlers added. Polling for tasks...")

	<-ctx.Done()

	slog.Info("Context cancelled. Graceful shutdown proceeding...")
}

func handleReserveResources(
	ctx *processor.Context,
	resourceClient *resourceapi.ClientWithResponses,
) error {
	slog.Info("Processing task",
		"taskId", ctx.Task.Id,
		"topicName", ctx.Task.TopicName,
		"businessKey", ctx.Task.BusinessKey,
	)

	appointmentIdVar, ok := ctx.Task.Variables["appointmentId"]
	if !ok || appointmentIdVar.Value == nil {
		slog.Error("Missing 'appointmentId' variable", "taskId", ctx.Task.Id)
		return ctx.HandleFailure(processor.QueryHandleFailure{
			ErrorMessage: server.AsPtr("Missing 'appointmentId' variable"),
			Retries:      server.AsPtr(0),
		})
	}
	appointmentIdStr, ok := appointmentIdVar.Value.(string)
	if !ok {
		slog.Error(
			"Invalid type for 'appointmentId'",
			"taskId",
			ctx.Task.Id,
			"type",
			fmt.Sprintf("%T", appointmentIdVar.Value),
		)
		return ctx.HandleFailure(processor.QueryHandleFailure{
			ErrorMessage: server.AsPtr("Invalid type for 'appointmentId', expected string UUID"),
			Retries:      server.AsPtr(0),
		})
	}

	apptUUID, err := uuid.Parse(appointmentIdStr)
	if err != nil {
		slog.Error(
			"Failed to parse 'appointmentId' as UUID",
			"taskId",
			ctx.Task.Id,
			"value",
			appointmentIdStr,
			"error",
			err,
		)
		return ctx.HandleFailure(processor.QueryHandleFailure{
			ErrorMessage: server.AsPtr(fmt.Sprintf("Invalid format for 'appointmentId': %v", err)),
			Retries:      server.AsPtr(0),
		})
	}

	appointmentDateTimeVar, ok := ctx.Task.Variables["appointmentDateTime"]
	if !ok || appointmentDateTimeVar.Value == nil {
		slog.Error("Missing 'appointmentDateTime' variable", "taskId", ctx.Task.Id)
		return ctx.HandleFailure(processor.QueryHandleFailure{
			ErrorMessage: server.AsPtr("Missing 'appointmentDateTime' variable"),
			Retries:      server.AsPtr(0),
		})
	}
	appointmentTime := time.Time{}
	switch v := appointmentDateTimeVar.Value.(type) {
	case time.Time:
		appointmentTime = v
	case string:
		parsedTime, parseErr := time.Parse(time.RFC3339, v)
		if parseErr != nil {
			slog.Error("Failed to parse 'appointmentDateTime' string", "taskId", ctx.Task.Id, "value", v, "error", parseErr)
			return ctx.HandleFailure(processor.QueryHandleFailure{
				ErrorMessage: server.AsPtr(fmt.Sprintf("Failed to parse 'appointmentDateTime' string: %v", parseErr)),
				Retries:      server.AsPtr(0),
			})
		}
		appointmentTime = parsedTime
	default:
		slog.Error("Invalid type for 'appointmentDateTime'", "taskId", ctx.Task.Id, "type", fmt.Sprintf("%T", appointmentDateTimeVar.Value))
		return ctx.HandleFailure(processor.QueryHandleFailure{
			ErrorMessage: server.AsPtr("Invalid type for 'appointmentDateTime', expected time.Time or string"),
			Retries:      server.AsPtr(0),
		})
	}

	var facilityId *uuid.UUID
	if facVar, ok := ctx.Task.Variables["facilityId"]; ok && facVar.Value != nil {
		if facIdStr, ok := facVar.Value.(string); ok && facIdStr != "" {
			facUUID, err := uuid.Parse(facIdStr)
			if err != nil {
				slog.Error(
					"Failed to parse 'facilityId' as UUID",
					"taskId",
					ctx.Task.Id,
					"value",
					facIdStr,
					"error",
					err,
				)
				return ctx.HandleFailure(processor.QueryHandleFailure{
					ErrorMessage: server.AsPtr(
						fmt.Sprintf("Invalid format for 'facilityId': %v", err),
					),
					Retries: server.AsPtr(0),
				})
			}
			facilityId = &facUUID
		}
	}

	var equipmentId *uuid.UUID
	if eqpVar, ok := ctx.Task.Variables["equipmentId"]; ok && eqpVar.Value != nil {
		if eqpIdStr, ok := eqpVar.Value.(string); ok && eqpIdStr != "" {
			eqpUUID, err := uuid.Parse(eqpIdStr)
			if err != nil {
				slog.Error(
					"Failed to parse 'equipmentId' as UUID",
					"taskId",
					ctx.Task.Id,
					"value",
					eqpIdStr,
					"error",
					err,
				)
				return ctx.HandleFailure(processor.QueryHandleFailure{
					ErrorMessage: server.AsPtr(
						fmt.Sprintf("Invalid format for 'equipmentId': %v", err),
					),
					Retries: server.AsPtr(0),
				})
			}
			equipmentId = &eqpUUID
		}
	}

	var medicineId *uuid.UUID
	if medVar, ok := ctx.Task.Variables["medicineId"]; ok && medVar.Value != nil {
		if medIdStr, ok := medVar.Value.(string); ok && medIdStr != "" {
			medUUID, err := uuid.Parse(medIdStr)
			if err != nil {
				slog.Error(
					"Failed to parse 'medicineId' as UUID",
					"taskId",
					ctx.Task.Id,
					"value",
					medIdStr,
					"error",
					err,
				)
				return ctx.HandleFailure(processor.QueryHandleFailure{
					ErrorMessage: server.AsPtr(
						fmt.Sprintf("Invalid format for 'medicineId': %v", err),
					),
					Retries: server.AsPtr(0),
				})
			}
			medicineId = &medUUID
		}
	}

	request := resourceapi.ReserveAppointmentResourcesJSONRequestBody{
		EquipmentId: equipmentId,
		FacilityId:  facilityId,
		MedicineId:  medicineId,
		Start:       appointmentTime,
	}
	res, err := resourceClient.ReserveAppointmentResourcesWithResponse(
		context.Background(),
		apptUUID,
		request,
	)
	if err != nil {
		slog.Error(
			server.UnexpectedError,
			"error",
			err.Error(),
			"where",
			"DecideAppointment reserve resources api call",
		)
		return ctx.HandleFailure(processor.QueryHandleFailure{
			ErrorMessage: server.AsPtr(fmt.Sprintf("Failed to send reservation: %v", err)),
			Retries:      server.AsPtr(0),
		})
	}

	if res.StatusCode() == http.StatusNoContent {
		err = ctx.Complete(processor.QueryComplete{})
		if err != nil {
			return ctx.HandleFailure(processor.QueryHandleFailure{
				ErrorMessage: server.AsPtr(fmt.Sprintf("failed to complete Camunda task: %v", err)),
				Retries:      server.AsPtr(0),
			})
		}
		return nil
	}

	return ctx.HandleFailure(processor.QueryHandleFailure{
		ErrorMessage: server.AsPtr(fmt.Sprintf("Failed to reserve resources: %v", err)),
		Retries:      server.AsPtr(0),
	})
}
