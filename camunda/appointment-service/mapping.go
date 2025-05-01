package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Nesquiko/aass/appointment-service/api"
	resourceapi "github.com/Nesquiko/aass/appointment-service/resources-api"
	"github.com/Nesquiko/aass/common/server"
)

func (a appointmentServer) mapDataApptToApiAppt(
	ctx context.Context,
	apptData Appointment,
) (api.Appointment, *server.ApiError) {
	patientResp, patientErr := a.userApi.GetPatientByIdWithResponse(ctx, apptData.PatientId)
	if patientErr != nil || patientResp.JSON200 == nil {
		slog.Error(
			"failed to get patient details for mapping",
			"error",
			patientErr,
			"status",
			patientResp.StatusCode(),
		)
		return api.Appointment{}, server.InternalServerError()
	}

	doctorResp, doctorErr := a.userApi.GetDoctorByIdWithResponse(ctx, apptData.DoctorId)
	if doctorErr != nil || doctorResp.JSON200 == nil {
		slog.Error(
			"failed to get doctor details for mapping",
			"error",
			doctorErr,
			"status",
			doctorResp.StatusCode(),
		)
		return api.Appointment{}, server.InternalServerError()
	}

	var conditionDisplay *api.ConditionDisplay = nil
	if apptData.ConditionId != nil {
		condResp, condErr := a.medicalApi.ConditionDetailWithResponse(ctx, *apptData.ConditionId)
		if condErr != nil || condResp.JSON200 == nil {
			// Log error but don't fail the whole request if condition is missing
			slog.Warn(
				"failed to get condition details for mapping",
				"error",
				condErr,
				"status",
				condResp.StatusCode(),
				"conditionId",
				apptData.ConditionId.String(),
			)
		} else {
			conditionDisplay = &api.ConditionDisplay{
				Id:              condResp.JSON200.Id,
				Name:            condResp.JSON200.Name,
				Start:           condResp.JSON200.Start,
				End:             condResp.JSON200.End,
				AppointmentsIds: condResp.JSON200.AppointmentsIds,
			}
		}
	}

	var prescriptionsDisplay *[]api.PrescriptionDisplay = nil
	prescResp, prescErr := a.medicalApi.GetPrescriptionsByAppointmentIdWithResponse(
		ctx,
		apptData.Id,
	)
	if prescErr != nil || prescResp.JSON200 == nil {
		// Log error but don't fail the whole request if prescriptions are missing
		slog.Warn(
			"failed to get prescriptions for mapping",
			"error",
			prescErr,
			"status",
			prescResp.StatusCode(),
			"appointmentId",
			apptData.Id.String(),
		)
	} else if prescResp.JSON200.Prescriptions != nil {
		prescs := make([]api.PrescriptionDisplay, len(prescResp.JSON200.Prescriptions))
		for i, p := range prescResp.JSON200.Prescriptions {
			prescs[i] = api.PrescriptionDisplay{
				Id:            p.Id,
				Name:          p.Name,
				Start:         p.Start,
				End:           p.End,
				AppointmentId: p.AppointmentId,
			}
		}
		prescriptionsDisplay = &prescs
	}

	apiPatient := api.Patient{
		Id:        patientResp.JSON200.Id,
		FirstName: patientResp.JSON200.FirstName,
		LastName:  patientResp.JSON200.LastName,
		Email:     patientResp.JSON200.Email,
		Role:      api.UserRole(patientResp.JSON200.Role),
	}

	apiDoctor := api.Doctor{
		Id:             doctorResp.JSON200.Id,
		FirstName:      doctorResp.JSON200.FirstName,
		LastName:       doctorResp.JSON200.LastName,
		Email:          doctorResp.JSON200.Email,
		Role:           api.UserRole(doctorResp.JSON200.Role),
		Specialization: api.SpecializationEnum(doctorResp.JSON200.Specialization),
	}

	var facilities *[]api.Facility = nil
	if len(apptData.Facilities) > 0 {
		f := make([]api.Facility, len(apptData.Facilities))
		for i, res := range apptData.Facilities {
			r, _ := a.resourceApi.GetResourceByIdWithResponse(ctx, res.Id)
			if r.JSON200 == nil {
				slog.Error("nil facility resources by ID response", "id", res.Id)
				continue
			}

			f[i] = api.Facility{Id: res.Id, Name: r.JSON200.Name}
		}
		facilities = &f
	}

	var equipment *[]api.Equipment = nil
	if len(apptData.Equipment) > 0 {
		e := make([]api.Equipment, len(apptData.Equipment))
		for i, res := range apptData.Equipment {
			r, _ := a.resourceApi.GetResourceByIdWithResponse(ctx, res.Id)
			if r.JSON200 == nil {
				slog.Error("nil equipment resources by ID response", "id", res.Id)
				continue
			}
			e[i] = api.Equipment{Id: res.Id, Name: r.JSON200.Name}
		}
		equipment = &e
	}

	var medicine *[]api.Medicine = nil
	if len(apptData.Medicines) > 0 {
		m := make([]api.Medicine, len(apptData.Medicines))
		for i, res := range apptData.Medicines {
			r, _ := a.resourceApi.GetResourceByIdWithResponse(ctx, res.Id)
			if r.JSON200 == nil {
				slog.Error("nil medicine resources by ID response", "id", res.Id)
				continue
			}
			m[i] = api.Medicine{Id: res.Id, Name: r.JSON200.Name}
		}
		medicine = &m
	}

	var canceledBy *api.UserRole = nil
	if apptData.CancelledBy != nil {
		role := api.UserRole(*apptData.CancelledBy)
		canceledBy = &role
	}

	return api.Appointment{
		Id:                  apptData.Id,
		AppointmentDateTime: apptData.AppointmentDateTime,
		Type:                api.AppointmentType(apptData.Type),
		Condition:           conditionDisplay,
		Status:              api.AppointmentStatus(apptData.Status),
		Reason:              apptData.Reason,
		CancellationReason:  apptData.CancellationReason,
		CanceledBy:          canceledBy,
		DenialReason:        apptData.DenialReason,
		Prescriptions:       prescriptionsDisplay,
		Patient:             apiPatient,
		Doctor:              apiDoctor,
		Facilities:          facilities,
		Equipment:           equipment,
		Medicine:            medicine,
	}, nil
}

func (a appointmentServer) mapDataApptToApiApptDisplay(
	ctx context.Context,
	apptData Appointment,
) (api.AppointmentDisplay, *server.ApiError) {
	patientResp, patientErr := a.userApi.GetPatientByIdWithResponse(ctx, apptData.PatientId)
	if patientErr != nil || patientResp.JSON200 == nil {
		slog.Error(
			"failed to get patient details for display mapping",
			"error",
			patientErr,
			"status",
			patientResp.StatusCode(),
		)
		return api.AppointmentDisplay{}, server.InternalServerError()
	}

	doctorResp, doctorErr := a.userApi.GetDoctorByIdWithResponse(ctx, apptData.DoctorId)
	if doctorErr != nil || doctorResp.JSON200 == nil {
		slog.Error(
			"failed to get doctor details for display mapping",
			"error",
			doctorErr,
			"status",
			doctorResp.StatusCode(),
		)
		return api.AppointmentDisplay{}, server.InternalServerError()
	}

	return api.AppointmentDisplay{
		Id:                  apptData.Id,
		AppointmentDateTime: apptData.AppointmentDateTime,
		DoctorName: fmt.Sprintf(
			"%s %s",
			doctorResp.JSON200.FirstName,
			doctorResp.JSON200.LastName,
		),
		PatientName: fmt.Sprintf(
			"%s %s",
			patientResp.JSON200.FirstName,
			patientResp.JSON200.LastName,
		),
		Status: api.AppointmentStatus(apptData.Status),
		Type:   api.AppointmentType(apptData.Type),
	}, nil
}
