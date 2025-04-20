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

func (a monolithApp) CreateResource(
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

func (a monolithApp) ReserveResource(
	ctx context.Context,
	resourceId uuid.UUID,
	res api.ResourceReservation,
) error {
	resource, err := a.db.ResourceById(ctx, resourceId)
	if err != nil {
		return fmt.Errorf("ReserveResource can't find resource by id %q: %w", resourceId, err)
	}
	_, err = a.db.CreateReservation(
		ctx,
		res.AppointmentId,
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

func (a monolithApp) AvailableResources(
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

func (a monolithApp) ReserveAppointmentResources(
	ctx context.Context,
	appointmentId uuid.UUID,
	payload api.ReserveAppointmentResourcesJSONBody,
) (api.DoctorAppointment, error) {
	appointment, err := a.db.AppointmentById(ctx, appointmentId)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return api.DoctorAppointment{}, fmt.Errorf(
				"ReserveAppointmentResources: appointment %w",
				ErrNotFound,
			)
		}
		return api.DoctorAppointment{}, fmt.Errorf(
			"ReserveAppointmentResources fetch appointment failed: %w",
			err,
		)
	}

	reservationStart := payload.Start
	reservationEnd := appointment.EndTime

	if payload.Equipment != nil {
		resourceId := *payload.Equipment
		resource, err := a.db.ResourceById(ctx, resourceId)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				return api.DoctorAppointment{}, fmt.Errorf(
					"ReserveAppointmentResources: equipment resource %s %w",
					resourceId,
					ErrNotFound,
				)
			}
			return api.DoctorAppointment{}, fmt.Errorf(
				"ReserveAppointmentResources fetch equipment resource %s failed: %w",
				resourceId,
				err,
			)
		}
		if resource.Type != data.ResourceTypeEquipment {
			return api.DoctorAppointment{}, fmt.Errorf(
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
				return api.DoctorAppointment{}, fmt.Errorf(
					"ReserveAppointmentResources: equipment %s %w",
					resourceId,
					ErrResourceUnavailable,
				)
			}
			return api.DoctorAppointment{}, fmt.Errorf(
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
				return api.DoctorAppointment{}, fmt.Errorf(
					"ReserveAppointmentResources: facility resource %s %w",
					resourceId,
					ErrNotFound,
				)
			}
			return api.DoctorAppointment{}, fmt.Errorf(
				"ReserveAppointmentResources fetch facility resource %s failed: %w",
				resourceId,
				err,
			)
		}
		if resource.Type != data.ResourceTypeFacility {
			return api.DoctorAppointment{}, fmt.Errorf(
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
				return api.DoctorAppointment{}, fmt.Errorf(
					"ReserveAppointmentResources: facility %s %w",
					resourceId,
					ErrResourceUnavailable,
				)
			}
			return api.DoctorAppointment{}, fmt.Errorf(
				"ReserveAppointmentResources create facility reservation for %s failed: %w",
				resourceId,
				err,
			)
		}
	}

	if payload.Medicine != nil {
		resourceId := *payload.Medicine
		resource, err := a.db.ResourceById(ctx, resourceId)
		if err != nil {
			if errors.Is(err, data.ErrNotFound) {
				return api.DoctorAppointment{}, fmt.Errorf(
					"ReserveAppointmentResources: medicine resource %s %w",
					resourceId,
					ErrNotFound,
				)
			}
			return api.DoctorAppointment{}, fmt.Errorf(
				"ReserveAppointmentResources fetch medicine resource %s failed: %w",
				resourceId,
				err,
			)
		}
		if resource.Type != data.ResourceTypeMedicine {
			return api.DoctorAppointment{}, fmt.Errorf(
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
				return api.DoctorAppointment{}, fmt.Errorf(
					"ReserveAppointmentResources: medicine %s %w",
					resourceId,
					ErrResourceUnavailable,
				)
			}
			return api.DoctorAppointment{}, fmt.Errorf(
				"ReserveAppointmentResources create medicine reservation for %s failed: %w",
				resourceId,
				err,
			)
		}
	}

	patient, err := a.db.PatientById(ctx, appointment.PatientId)
	if err != nil {
		return api.DoctorAppointment{}, fmt.Errorf(
			"ReserveAppointmentResources fetch patient %s failed: %w",
			appointment.PatientId,
			err,
		)
	}

	var cond *data.Condition
	if appointment.ConditionId != nil {
		c, err := a.db.ConditionById(ctx, *appointment.ConditionId)
		if err != nil {
			return api.DoctorAppointment{}, fmt.Errorf(
				"ReserveAppointmentResources fetch condition %s failed: %w",
				*appointment.ConditionId,
				err,
			)
		}
		cond = &c
	}

	allReservedResources, err := a.db.ResourcesByAppointmentId(
		ctx,
		appointmentId,
	)
	if err != nil {
		return api.DoctorAppointment{}, fmt.Errorf(
			"ReserveAppointmentResources fetch final resources failed: %w",
			err,
		)
	}

	var facilities, equipment, medicine []data.Resource
	for _, resource := range allReservedResources {
		switch resource.Type {
		case data.ResourceTypeFacility:
			facilities = append(facilities, resource)
		case data.ResourceTypeEquipment:
			equipment = append(equipment, resource)
		case data.ResourceTypeMedicine:
			medicine = append(medicine, resource)
		}
	}

	prescriptions, err := a.db.PrescriptionByAppointmentId(ctx, appointmentId)
	if err != nil {
		return api.DoctorAppointment{}, fmt.Errorf(
			"ReserveAppointmentResources fetch prescriptions: %w",
			err,
		)
	}

	doctorAppointment := dataApptToDoctorAppt(
		appointment,
		patient,
		cond,
		facilities,
		equipment,
		medicine,
		prescriptions,
	)

	return doctorAppointment, nil
}
