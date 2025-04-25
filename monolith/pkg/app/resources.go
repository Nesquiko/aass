package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Nesquiko/wac/pkg/api"
	"github.com/Nesquiko/wac/pkg/data"
)

func (a MonolithApp) CreateResource(
	ctx context.Context,
	resource api.NewResource,
) (api.NewResource, error) {
	res, err := a.db.CreateResource(ctx, resource.Name, data.ResourceType(resource.Type))
	if err != nil {
		return api.NewResource{}, fmt.Errorf("CreateResource: %w", err)
	}

	return api.NewResource{
		Id:   &res.Id,
		Name: res.Name,
		Type: api.ResourceType(res.Type),
	}, nil
}

func (a MonolithApp) AvailableResources(
	ctx context.Context,
	dateTime time.Time,
) (api.AvailableResources, error) {
	resources, err := a.db.FindAvailableResourcesAtTime(ctx, dateTime)
	if err != nil {
		return api.AvailableResources{}, fmt.Errorf("AvailableResources: %w", err)
	}

	available := api.AvailableResources{
		Equipment:  make([]api.Equipment, len(resources.Equipment)),
		Facilities: make([]api.Facility, len(resources.Facilities)),
		Medicine:   make([]api.Medicine, len(resources.Medicines)),
	}

	for i, res := range resources.Equipment {
		available.Equipment[i] = resourceToEquipment(res)
	}

	for i, res := range resources.Facilities {
		available.Facilities[i] = resourceToFacility(res)
	}

	for i, res := range resources.Medicines {
		available.Medicine[i] = resourceToMedicine(res)
	}

	return available, nil
}

func (a MonolithApp) ReserveAppointmentResources(
	ctx context.Context,
	appointmentId uuid.UUID,
	payload api.ReserveAppointmentResourcesJSONBody,
) error {
	appointment, err := a.db.AppointmentById(ctx, appointmentId)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return fmt.Errorf(
				"ReserveAppointmentResources: appointment %w",
				ErrNotFound,
			)
		}
		return fmt.Errorf(
			"ReserveAppointmentResources fetch appointment failed: %w",
			err,
		)
	}

	reservationStart := payload.Start
	reservationEnd := appointment.EndTime

	if payload.EquipmentId != nil {
		resourceId := *payload.EquipmentId
		resource, err := a.db.ResourceById(ctx, resourceId)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				return fmt.Errorf(
					"ReserveAppointmentResources: equipment resource %s %w",
					resourceId,
					ErrNotFound,
				)
			}
			return fmt.Errorf(
				"ReserveAppointmentResources fetch equipment resource %s failed: %w",
				resourceId,
				err,
			)
		}
		if resource.Type != data.ResourceTypeEquipment {
			return fmt.Errorf(
				"ReserveAppointmentResources: resource %s is not equipment",
				resourceId,
			)
		}
		_, err = a.db.CreateReservation(
			ctx,
			appointmentId,
			resourceId,
			resource.Name,
			resource.Type,
			reservationStart,
			reservationEnd,
		)
		if err != nil {
			if errors.Is(err, data.ErrResourceUnavailable) {
				return fmt.Errorf(
					"ReserveAppointmentResources: equipment %s %w",
					resourceId,
					ErrResourceUnavailable,
				)
			}
			return fmt.Errorf(
				"ReserveAppointmentResources create equipment reservation for %s failed: %w",
				resourceId,
				err,
			)
		}
	}

	if payload.FacilityId != nil {
		resourceId := *payload.FacilityId
		resource, err := a.db.ResourceById(ctx, resourceId)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				return fmt.Errorf(
					"ReserveAppointmentResources: facility resource %s %w",
					resourceId,
					ErrNotFound,
				)
			}
			return fmt.Errorf(
				"ReserveAppointmentResources fetch facility resource %s failed: %w",
				resourceId,
				err,
			)
		}
		if resource.Type != data.ResourceTypeFacility {
			return fmt.Errorf(
				"ReserveAppointmentResources: resource %s is not a facility",
				resourceId,
			)
		}
		_, err = a.db.CreateReservation(
			ctx,
			appointmentId,
			resourceId,
			resource.Name,
			resource.Type,
			reservationStart,
			reservationEnd,
		)
		if err != nil {
			if errors.Is(err, data.ErrResourceUnavailable) {
				return fmt.Errorf(
					"ReserveAppointmentResources: facility %s %w",
					resourceId,
					ErrResourceUnavailable,
				)
			}
			return fmt.Errorf(
				"ReserveAppointmentResources create facility reservation for %s failed: %w",
				resourceId,
				err,
			)
		}
	}

	if payload.MedicineId != nil {
		resourceId := *payload.MedicineId
		resource, err := a.db.ResourceById(ctx, resourceId)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				return fmt.Errorf(
					"ReserveAppointmentResources: medicine resource %s %w",
					resourceId,
					ErrNotFound,
				)
			}
			return fmt.Errorf(
				"ReserveAppointmentResources fetch medicine resource %s failed: %w",
				resourceId,
				err,
			)
		}
		if resource.Type != data.ResourceTypeMedicine {
			return fmt.Errorf(
				"ReserveAppointmentResources: resource %s is not medicine",
				resourceId,
			)
		}
		_, err = a.db.CreateReservation(
			ctx,
			appointmentId,
			resourceId,
			resource.Name,
			resource.Type,
			reservationStart,
			reservationEnd,
		)
		if err != nil {
			if errors.Is(err, data.ErrResourceUnavailable) {
				return fmt.Errorf(
					"ReserveAppointmentResources: medicine %s %w",
					resourceId,
					ErrResourceUnavailable,
				)
			}
			return fmt.Errorf(
				"ReserveAppointmentResources create medicine reservation for %s failed: %w",
				resourceId,
				err,
			)
		}
	}

	return nil
}
