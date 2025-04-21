package pkg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"
	"github.com/oapi-codegen/runtime/types"
	"golang.org/x/sync/errgroup"

	"github.com/Nesquiko/aass/patient-service/api"
	apptapi "github.com/Nesquiko/aass/patient-service/api/appt-api"
	condapi "github.com/Nesquiko/aass/patient-service/api/cond-api"
	presapi "github.com/Nesquiko/aass/patient-service/api/pres-api"
)

var ErrDownstreamService = errors.New("failed to fetch data from a downstream service")

type PatientApp struct {
	db         *MongoPatientDb
	apptClient *apptapi.ClientWithResponses
	condClient *condapi.ClientWithResponses
	presClient *presapi.ClientWithResponses
}

func (a PatientApp) CreatePatient(
	ctx context.Context,
	p api.NewPatientRequest,
) (api.Patient, error) {
	patient := patientRegToDataPatient(p)

	patient, err := a.db.CreatePatient(ctx, patient)
	if errors.Is(err, ErrDuplicateEmail) {
		return api.Patient{}, fmt.Errorf("CreatePatient duplicate emall: %w", ErrDuplicateEmail)
	} else if err != nil {
		return api.Patient{}, fmt.Errorf("CreatePatient: %w", err)
	}

	return dataPatientToApiPatient(patient), nil
}

func (a PatientApp) PatientById(ctx context.Context, id uuid.UUID) (api.Patient, error) {
	patient, err := a.db.FindPatientById(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return api.Patient{}, fmt.Errorf("PatientById: %w", ErrNotFound)
		}
		return api.Patient{}, fmt.Errorf("PatientById: %w", err)
	}

	return dataPatientToApiPatient(patient), nil
}

func (a PatientApp) PatientByEmail(ctx context.Context, email string) (api.Patient, error) {
	patient, err := a.db.FindPatientByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return api.Patient{}, fmt.Errorf("PatientByEmail: %w", ErrNotFound)
		}
		return api.Patient{}, fmt.Errorf("PatientByEmail: %w", err)
	}

	return dataPatientToApiPatient(patient), nil
}

func (a PatientApp) GetPatientCalendar(
	ctx context.Context,
	patientId uuid.UUID,
	from time.Time,
	to time.Time,
) (api.PatientCalendarView, error) {
	// 1. Verify patient exists (optional, but good practice)
	_, err := a.db.FindPatientById(ctx, patientId)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return api.PatientCalendarView{}, ErrNotFound // Return specific error
		}
		slog.ErrorContext(
			ctx,
			"Failed patient check for calendar",
			"patientId",
			patientId,
			"error",
			err,
		)
		return api.PatientCalendarView{}, fmt.Errorf("failed patient check: %w", err)
	}

	// 2. Fetch data concurrently using errgroup
	var calendarView api.PatientCalendarView
	g, childCtx := errgroup.WithContext(ctx)

	// Fetch Appointments
	g.Go(func() error {
		slog.InfoContext(childCtx, "Fetching appointments for calendar", "patientId", patientId)
		params := &apptapi.GetPatientAppointmentsParams{
			From: &types.Date{Time: from},
			To:   &types.Date{Time: to},
		}
		resp, err := a.apptClient.GetPatientAppointmentsWithResponse(childCtx, patientId, params)
		if err != nil {
			slog.ErrorContext(
				childCtx,
				"Appointment client call failed",
				"patientId",
				patientId,
				"error",
				err,
			)
			return fmt.Errorf(
				"%w: appointment service request failed: %v",
				ErrDownstreamService,
				err,
			)
		}

		if resp.StatusCode() != http.StatusOK {
			if resp.StatusCode() == http.StatusNotFound {
				slog.InfoContext(
					childCtx,
					"No appointments found for patient",
					"patientId",
					patientId,
					"status",
					resp.StatusCode(),
				)
				calendarView.Appointments = &[]api.AppointmentDisplay{} // Ensure empty slice, not nil
				return nil
			}
			errMsg := fmt.Sprintf("appointment service returned status %d", resp.StatusCode())
			if resp.ApplicationproblemJSON500 != nil {
				errMsg = fmt.Sprintf("%s: %s", errMsg, resp.ApplicationproblemJSON500.Detail)
			} else if len(resp.Body) > 0 {
				errMsg = fmt.Sprintf("%s - Body: %s", errMsg, string(resp.Body))
			}
			slog.ErrorContext(
				childCtx,
				"Appointment service error response",
				"patientId",
				patientId,
				"status",
				resp.StatusCode(),
				"response",
				errMsg,
			)
			return fmt.Errorf("%w: %s", ErrDownstreamService, errMsg)
		}

		if resp.JSON200 == nil || resp.JSON200.Appointments == nil {
			slog.InfoContext(
				childCtx,
				"Appointments response body or appointments list is nil",
				"patientId",
				patientId,
			)
			calendarView.Appointments = &[]api.AppointmentDisplay{}
			return nil
		}

		// Map results
		apiAppointments := make([]api.AppointmentDisplay, 0, len(resp.JSON200.Appointments))
		for _, appt := range resp.JSON200.Appointments {
			// Basic mapping - assumes PatientAppointment has necessary fields
			// In real scenario, might need more complex mapping or enrichment
			apiAppt := api.AppointmentDisplay{
				Id:                  *appt.Id, // Dereference pointer
				AppointmentDateTime: appt.AppointmentDateTime,
				Status:              api.AppointmentStatus(appt.Status), // Cast status
				Type:                api.AppointmentType(appt.Type),     // Cast type
				// DoctorName needs enrichment - assuming it's present in PatientAppointment for now
				DoctorName: fmt.Sprintf("%s %s", appt.Doctor.FirstName, appt.Doctor.LastName),
			}
			apiAppointments = append(apiAppointments, apiAppt)
		}
		calendarView.Appointments = &apiAppointments
		slog.InfoContext(
			childCtx,
			"Successfully fetched appointments",
			"patientId",
			patientId,
			"count",
			len(apiAppointments),
		)
		return nil
	})

	// Fetch Conditions
	g.Go(func() error {
		slog.InfoContext(childCtx, "Fetching conditions for calendar", "patientId", patientId)
		// Assuming Condition Service has a similar endpoint structure
		params := &condapi.GetConditionsByPatientAndRangeParams{ // Adjust based on actual generated client params
			PatientId: patientId,
			From:      types.Date{Time: from},
			To:        types.Date{Time: to},
		}
		resp, err := a.condClient.GetConditionsByPatientAndRangeWithResponse(childCtx, params)
		if err != nil {
			slog.ErrorContext(
				childCtx,
				"Condition client call failed",
				"patientId",
				patientId,
				"error",
				err,
			)
			return fmt.Errorf("%w: condition service request failed: %v", ErrDownstreamService, err)
		}

		if resp.StatusCode() != http.StatusOK {
			if resp.StatusCode() == http.StatusNotFound {
				slog.InfoContext(
					childCtx,
					"No conditions found for patient",
					"patientId",
					patientId,
					"status",
					resp.StatusCode(),
				)
				calendarView.Conditions = &[]api.ConditionDisplay{}
				return nil
			}
			errMsg := fmt.Sprintf("condition service returned status %d", resp.StatusCode())
			// Add more detail parsing if available from condapi response
			slog.ErrorContext(
				childCtx,
				"Condition service error response",
				"patientId",
				patientId,
				"status",
				resp.StatusCode(),
				"response",
				errMsg,
			)
			return fmt.Errorf("%w: %s", ErrDownstreamService, errMsg)
		}

		if resp.JSON200 == nil ||
			resp.JSON200.Conditions == nil { // Adjust based on actual response structure
			slog.InfoContext(
				childCtx,
				"Conditions response body or list is nil",
				"patientId",
				patientId,
			)
			calendarView.Conditions = &[]api.ConditionDisplay{}
			return nil
		}

		apiConditions := make([]api.ConditionDisplay, 0, len(resp.JSON200.Conditions))
		for _, cond := range resp.JSON200.Conditions {
			end := nullable.NewNullNullable[time.Time]()
			if cond.End != nil {
				end.Set(*cond.End)
			}
			apiCond := api.ConditionDisplay{
				Id:    *cond.Id,
				Name:  cond.Name,
				Start: cond.Start,
				End:   end,
			}
			apiConditions = append(apiConditions, apiCond)
		}
		calendarView.Conditions = &apiConditions
		slog.InfoContext(
			childCtx,
			"Successfully fetched conditions",
			"patientId",
			patientId,
			"count",
			len(apiConditions),
		)
		return nil
	})

	// Fetch Prescriptions
	g.Go(func() error {
		slog.InfoContext(childCtx, "Fetching prescriptions for calendar", "patientId", patientId)
		// Assuming Prescription Service has a similar endpoint structure
		params := &presapi.GetPrescriptionsByPatientAndRangeParams{
			PatientId: patientId, // Assuming param is pointer
			From:      types.Date{Time: from},
			To:        types.Date{Time: to},
		}
		resp, err := a.presClient.GetPrescriptionsByPatientAndRangeWithResponse(childCtx, params)
		if err != nil {
			slog.ErrorContext(
				childCtx,
				"Prescription client call failed",
				"patientId",
				patientId,
				"error",
				err,
			)
			return fmt.Errorf(
				"%w: prescription service request failed: %v",
				ErrDownstreamService,
				err,
			)
		}

		if resp.StatusCode() != http.StatusOK {
			if resp.StatusCode() == http.StatusNotFound {
				slog.InfoContext(
					childCtx,
					"No prescriptions found for patient",
					"patientId",
					patientId,
					"status",
					resp.StatusCode(),
				)
				calendarView.Prescriptions = &[]api.PrescriptionDisplay{}
				return nil
			}
			errMsg := fmt.Sprintf("prescription service returned status %d", resp.StatusCode())
			// Add more detail parsing if available from presapi response
			slog.ErrorContext(
				childCtx,
				"Prescription service error response",
				"patientId",
				patientId,
				"status",
				resp.StatusCode(),
				"response",
				errMsg,
			)
			return fmt.Errorf("%w: %s", ErrDownstreamService, errMsg)
		}

		if resp.JSON200 == nil ||
			resp.JSON200.Prescriptions == nil { // Adjust based on actual response structure
			slog.InfoContext(
				childCtx,
				"Prescriptions response body or list is nil",
				"patientId",
				patientId,
			)
			calendarView.Prescriptions = &[]api.PrescriptionDisplay{}
			return nil
		}

		// Map results
		apiPrescriptions := make([]api.PrescriptionDisplay, 0, len(resp.JSON200.Prescriptions))
		for _, pres := range resp.JSON200.Prescriptions {
			apiPres := api.PrescriptionDisplay{
				Id:    *pres.Id,
				Name:  pres.Name,
				Start: pres.Start,
				End:   pres.End,
			}
			apiPrescriptions = append(apiPrescriptions, apiPres)
		}
		calendarView.Prescriptions = &apiPrescriptions
		slog.InfoContext(
			childCtx,
			"Successfully fetched prescriptions",
			"patientId",
			patientId,
			"count",
			len(apiPrescriptions),
		)
		return nil
	})

	// Wait for all fetches to complete and check for errors
	if err := g.Wait(); err != nil {
		// Error already logged in goroutines
		return api.PatientCalendarView{}, err // Return the wrapped ErrDownstreamService or other error
	}

	// Ensure slices are not nil if empty
	if calendarView.Appointments == nil {
		calendarView.Appointments = &[]api.AppointmentDisplay{}
	}
	if calendarView.Conditions == nil {
		calendarView.Conditions = &[]api.ConditionDisplay{}
	}
	if calendarView.Prescriptions == nil {
		calendarView.Prescriptions = &[]api.PrescriptionDisplay{}
	}

	return calendarView, nil
}
