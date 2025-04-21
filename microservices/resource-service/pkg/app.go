package pkg

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Nesquiko/aass/resource-service/api"
)

type ResourceApp struct {
	db *MongoResourceDb
}

func (a ResourceApp) CreateResource(
	ctx context.Context,
	resource api.NewResource,
) (api.Resource, error) {
	res, err := a.db.CreateResource(ctx, resource.Name, ResourceType(resource.Type))
	if err != nil {
		return api.Resource{}, fmt.Errorf("CreateResource: %w", err)
	}

	return api.Resource{
		Id:   &res.Id,
		Name: res.Name,
		Type: api.ResourceType(res.Type),
	}, nil
}

func (a ResourceApp) ReserveResource(
	ctx context.Context,
	resourceId uuid.UUID,
	appointmentId uuid.UUID,
	res api.ReserveAppointmentResourcesJSONBody,
) error {
	resource, err := a.db.ResourceById(ctx, resourceId)
	if err != nil {
		return fmt.Errorf("ReserveResource can't find resource by id %q: %w", resourceId, err)
	}
	_, err = a.db.CreateReservation(
		ctx,
		appointmentId,
		resourceId,
		resource.Name,
		resource.Type,
		res.Start,
		res.End,
	)
	if err != nil {
		return fmt.Errorf("ReserveResource can't reserve resource: %w", err)
	}
	return nil
}

func (a ResourceApp) AvailableResources(
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

func (a ResourceApp) ReserveAppointmentResources(
	ctx context.Context,
	appointmentId uuid.UUID,
	payload api.ReserveAppointmentResourcesJSONBody,
) error {
	reservationStart := payload.Start
	reservationEnd := reservationStart.Add(time.Hour)

	for _, res := range payload.Resources {
		resource, err := a.db.ResourceById(ctx, res.Id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return fmt.Errorf(
					"ReserveAppointmentResources: equipment resource %s %w",
					res.Id,
					ErrNotFound,
				)
			}
			return fmt.Errorf(
				"ReserveAppointmentResources fetch equipment resource %s failed: %w",
				res.Id,
				err,
			)
		}

		_, err = a.db.CreateReservation(
			ctx,
			appointmentId,
			res.Id,
			resource.Name,
			resource.Type,
			reservationStart,
			reservationEnd,
		)
		if err != nil {
			if errors.Is(err, ErrResourceUnavailable) {
				return fmt.Errorf(
					"ReserveAppointmentResources: equipment %s %w",
					res.Id,
					ErrResourceUnavailable,
				)
			}
			return fmt.Errorf(
				"ReserveAppointmentResources create equipment reservation for %s failed: %w",
				res.Id,
				err,
			)
		}

	}

	return nil
}
