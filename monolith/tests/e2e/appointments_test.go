//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	netUrl "net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"

	"github.com/Nesquiko/wac/pkg/api"
	"github.com/Nesquiko/wac/pkg/server"
)

func TestRescheduleAppointment(t *testing.T) {
	t.Parallel()

	patientEmail := fmt.Sprintf("test.reschedule.appt.%s@patient.com", uuid.NewString())
	patient := mustCreatePatient(t, newPatient(patientEmail))

	doctorEmail := fmt.Sprintf("test.reschedule.appt.%s@doctor.com", uuid.NewString())
	doctor := mustCreateDoctor(t, newDoctor(doctorEmail))

	appointmentTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	newAppointmentReq := api.NewAppointmentRequest{
		PatientId:           patient.Id,
		DoctorId:            doctor.Id,
		AppointmentDateTime: appointmentTime,
	}
	createdAppointment := mustCreateAppointment(t, newAppointmentReq)
	appointmentId := *createdAppointment.Id

	newDateTime := appointmentTime.Add(2 * 24 * time.Hour)
	rescheduleReq := api.AppointmentReschedule{
		NewAppointmentDateTime: newDateTime,
	}
	rescheduleReqBody, err := json.Marshal(rescheduleReq)
	require.NoError(t, err, "Failed to marshal reschedule request")

	url := fmt.Sprintf("%s/appointments/%s", ServerUrl, appointmentId)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(rescheduleReqBody))
	require.NoError(t, err, "Failed to create HTTP PATCH request")
	req.Header.Set("Content-Type", server.ApplicationJSON)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "http.Patch failed for RescheduleAppointment")
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode, "Expected '200 OK' status code")

	var fetchedAppointment api.PatientAppointment
	err = json.NewDecoder(res.Body).Decode(&fetchedAppointment)
	require.NoError(t, err, "Failed to decode fetched appointment")

	assert := assert.New(t)
	assert.Equal(appointmentId, *fetchedAppointment.Id, "Appointment ID mismatch")
	assert.True(
		newDateTime.Equal(fetchedAppointment.AppointmentDateTime),
		"Appointment date time mismatch",
	)
	assert.Equal(
		api.Requested,
		fetchedAppointment.Status,
		"Appointment status should be 'requested'",
	)
}

func TestDoctorsTimeslots(t *testing.T) {
	t.Parallel()

	doctorEmail := fmt.Sprintf("test.doctor.timeslots.%s@doctor.com", uuid.NewString())
	doctor := mustCreateDoctor(t, newDoctor(doctorEmail))

	patientEmail := fmt.Sprintf("test.doctor.timeslots.%s@patient.com", uuid.NewString())
	patient := mustCreatePatient(t, newPatient(patientEmail))

	date := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)

	appointmentTime := time.Date(
		date.Year(),
		date.Month(),
		date.Day(),
		10,
		0,
		0,
		0,
		time.UTC,
	)
	newAppointmentReq := api.NewAppointmentRequest{
		PatientId:           patient.Id,
		DoctorId:            doctor.Id,
		AppointmentDateTime: appointmentTime,
	}
	mustCreateAppointment(t, newAppointmentReq)

	url := fmt.Sprintf(
		"%s/doctors/%s/timeslots?date=%s",
		ServerUrl,
		doctor.Id,
		netUrl.QueryEscape(date.Format("2006-01-02")),
	)

	res, err := http.Get(url)
	require.NoError(t, err, "http.Get failed for DoctorsTimeslots")
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode, "Expected '200 OK' status code")

	var doctorTimeslots api.DoctorTimeslots
	err = json.NewDecoder(res.Body).Decode(&doctorTimeslots)
	require.NoError(t, err, "Failed to decode doctor timeslots")

	assert := assert.New(t)
	require.Len(t, doctorTimeslots.Slots, 7)

	for i, slot := range doctorTimeslots.Slots {
		hour := 8 + i
		expectedTime := time.Date(date.Year(), date.Month(), date.Day(), hour, 0, 0, 0, time.UTC).
			Format("15:04")
		assert.Equal(expectedTime, slot.Time, "Time mismatch for hour %d", hour)

		if hour == 10 {
			assert.Equal(
				api.Unavailable,
				slot.Status,
				"Expected 'unavailable' status for hour %d",
				hour,
			)
		} else {
			assert.Equal(api.Available, slot.Status, "Expected 'available' status for hour %d", hour)
		}
	}
}

func TestPatientsCalendar(t *testing.T) {
	t.Parallel()

	patientEmail := fmt.Sprintf("test.patient.calendar.%s@patient.com", uuid.NewString())
	patient := mustCreatePatient(t, newPatient(patientEmail))

	doctorEmail := fmt.Sprintf("test.patient.calendar.%s@doctor.com", uuid.NewString())
	doctor := mustCreateDoctor(t, newDoctor(doctorEmail))

	appointmentIds := make(map[uuid.UUID]bool)
	appointmentTimes := make(map[uuid.UUID]time.Time)
	startDate := time.Date(2025, time.April, 5, 0, 0, 0, 0, time.UTC)
	for i := range 10 {
		appointmentTime := startDate.Add(time.Duration(i) * 24 * time.Hour)
		newAppointmentReq := api.NewAppointmentRequest{
			PatientId:           patient.Id,
			DoctorId:            doctor.Id,
			AppointmentDateTime: appointmentTime,
		}
		createdAppointment := mustCreateAppointment(t, newAppointmentReq)
		appointmentIds[*createdAppointment.Id] = true
		appointmentTimes[*createdAppointment.Id] = appointmentTime
	}

	conditionIds := make(map[uuid.UUID]bool)
	conditionStartDates := make(map[uuid.UUID]time.Time)
	for i := range 10 {
		conditionStartDate := startDate.Add(time.Duration(i) * 24 * time.Hour)
		newConditionReq := api.NewCondition{
			PatientId: patient.Id,
			Start:     conditionStartDate,
			End:       asPtr(conditionStartDate.AddDate(1, 0, 0)),
			Name:      fmt.Sprintf("Condition %d", i),
		}
		createdCondition := mustCreateCondition(t, newConditionReq)
		conditionIds[*createdCondition.Id] = true
		conditionStartDates[*createdCondition.Id] = conditionStartDate
	}

	prescriptionIds := make(map[uuid.UUID]bool)
	prescriptionDates := make(map[uuid.UUID]time.Time)
	for i := range 10 {
		prescriptionDate := startDate.Add(time.Duration(i) * 24 * time.Hour)
		newPrescriptionReq := api.NewPrescription{
			PatientId: patient.Id,
			Start:     prescriptionDate,
			End:       prescriptionDate.AddDate(1, 0, 0),
			Name:      "Medication",
		}
		createdPrescription := mustCreatePrescription(t, newPrescriptionReq)
		prescriptionIds[*createdPrescription.Id] = true
		prescriptionDates[*createdPrescription.Id] = prescriptionDate
	}

	fromDate := startDate.AddDate(0, 0, 3)
	toDate := startDate.AddDate(0, 0, 6)

	url := fmt.Sprintf(
		"%s/patients/%s/calendar?from=%s&to=%s",
		ServerUrl,
		patient.Id,
		netUrl.QueryEscape(fromDate.Format("2006-01-02")),
		netUrl.QueryEscape(toDate.Format("2006-01-02")),
	)

	res, err := http.Get(url)
	require.NoError(t, err, "http.Get failed for PatientsCalendar")
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode, "Expected '200 OK' status code")

	var patientCalendar api.PatientsCalendar
	err = json.NewDecoder(res.Body).Decode(&patientCalendar)
	require.NoError(t, err, "Failed to decode patient calendar")

	assert := assert.New(t)
	assert.NotNil(patientCalendar.Appointments)
	assert.Len(*patientCalendar.Appointments, 4)

	for _, appt := range *patientCalendar.Appointments {
		assert.True(appointmentIds[appt.Id], "Unexpected appointment ID")
		assert.True(
			appointmentTimes[appt.Id].Equal(appt.AppointmentDateTime),
			"Appointment date time mismatch",
		)
		assert.True(
			fromDate.Before(appt.AppointmentDateTime) || fromDate.Equal(appt.AppointmentDateTime),
		)
		assert.True(
			toDate.After(appt.AppointmentDateTime) || toDate.Equal(appt.AppointmentDateTime),
		)
	}

	assert.NotNil(patientCalendar.Conditions)
	assert.Len(*patientCalendar.Conditions, 7)

	for _, cond := range *patientCalendar.Conditions {
		assert.True(conditionIds[*cond.Id], "Unexpected condition ID")
		assert.True(
			conditionStartDates[*cond.Id].Equal(cond.Start),
			"Condition start date mismatch",
		)
		assert.True(
			!(toDate.Before(cond.Start) || fromDate.After(*cond.End)),
			"Condition date range does not intersect with from-to range",
		)

	}

	assert.NotNil(patientCalendar.Prescriptions)
	assert.Len(*patientCalendar.Prescriptions, 7)

	for _, presc := range *patientCalendar.Prescriptions {
		assert.True(prescriptionIds[*presc.Id], "Unexpected prescription ID")
		assert.True(prescriptionDates[*presc.Id].Equal(presc.Start), "Prescription date mismatch")
		assert.True(
			!(toDate.Before(presc.Start) || fromDate.After(presc.End)),
			"Prescription date range does not intersect with from-to range",
		)
	}
}

func TestDoctorsCalendar(t *testing.T) {
	t.Parallel()

	doctorEmail := fmt.Sprintf("test.doctor.calendar.%s@doctor.com", uuid.NewString())
	doctor := mustCreateDoctor(t, newDoctor(doctorEmail))

	patientEmail := fmt.Sprintf("test.doctor.calendar.%s@patient.com", uuid.NewString())
	patient := mustCreatePatient(t, newPatient(patientEmail))

	appointmentIds := make(map[uuid.UUID]bool)
	appointmentTimes := make(map[uuid.UUID]time.Time)
	startDate := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)
	for i := range 10 {
		appointmentTime := startDate.Add(time.Duration(i) * 24 * time.Hour)
		newAppointmentReq := api.NewAppointmentRequest{
			PatientId:           patient.Id,
			DoctorId:            doctor.Id,
			AppointmentDateTime: appointmentTime,
		}
		createdAppointment := mustCreateAppointment(t, newAppointmentReq)
		appointmentIds[*createdAppointment.Id] = true
		appointmentTimes[*createdAppointment.Id] = appointmentTime
	}

	fromDate := startDate.Add(3 * 24 * time.Hour)
	toDate := startDate.Add(6 * 24 * time.Hour)

	url := fmt.Sprintf(
		"%s/doctors/%s/calendar?from=%s&to=%s",
		ServerUrl,
		doctor.Id,
		netUrl.QueryEscape(fromDate.Format("2006-01-02")),
		netUrl.QueryEscape(toDate.Format("2006-01-02")),
	)

	res, err := http.Get(url)
	require.NoError(t, err, "http.Get failed for DoctorsCalendar")
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode, "Expected '200 OK' status code")

	var doctorCalendar api.DoctorCalendar
	err = json.NewDecoder(res.Body).Decode(&doctorCalendar)
	require.NoError(t, err, "Failed to decode doctor calendar")

	assert := assert.New(t)
	assert.NotNil(doctorCalendar.Appointments)
	assert.Len(*doctorCalendar.Appointments, 4)

	for _, appt := range *doctorCalendar.Appointments {
		assert.True(appointmentIds[appt.Id], "Unexpected appointment ID")
		assert.True(
			appointmentTimes[appt.Id].Equal(appt.AppointmentDateTime),
			"Appointment date time mismatch",
		)
		assert.True(
			fromDate.Before(appt.AppointmentDateTime) || fromDate.Equal(appt.AppointmentDateTime),
		)
		assert.True(
			toDate.After(appt.AppointmentDateTime) || toDate.Equal(appt.AppointmentDateTime),
		)
	}
}

func TestDecideAppointmentApprove(t *testing.T) {
	t.Parallel()

	patientEmail := fmt.Sprintf("test.decide.appt.approve.%s@patient.com", uuid.NewString())
	patient := mustCreatePatient(t, newPatient(patientEmail))

	doctorEmail := fmt.Sprintf("test.decide.appt.approve.%s@doctor.com", uuid.NewString())
	doctor := mustCreateDoctor(t, newDoctor(doctorEmail))

	appointmentTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	newAppointmentReq := api.NewAppointmentRequest{
		PatientId:           patient.Id,
		DoctorId:            doctor.Id,
		AppointmentDateTime: appointmentTime,
		Type:                asPtr(api.RegularCheck),
	}
	createdAppointment := mustCreateAppointment(t, newAppointmentReq)
	appointmentId := *createdAppointment.Id

	resourceName := "Test Resource"
	resourceType := api.ResourceTypeEquipment
	resource := mustCreateResource(t, api.NewResource{Name: resourceName, Type: resourceType})

	decision := api.AppointmentDecision{
		Action: api.Accept,
		Equipment: &[]api.Equipment{
			{
				Id:   *resource.Id,
				Name: resource.Name,
			},
		},
	}
	decisionReqBody, err := json.Marshal(decision)
	require.NoError(t, err, "Failed to marshal decision request body")

	url := fmt.Sprintf("%s/appointments/%s", ServerUrl, appointmentId)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(decisionReqBody))
	require.NoError(t, err, "Failed to create HTTP POST request")
	req.Header.Set("Content-Type", server.ApplicationJSON)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "http.Post failed for DecideAppointment")
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode, "Expected '200 OK' status code")

	var fetchedAppointment api.DoctorAppointment
	err = json.NewDecoder(res.Body).Decode(&fetchedAppointment)
	require.NoError(t, err, "Failed to decode fetched appointment")

	assert := assert.New(t)
	assert.Equal(
		api.Scheduled,
		fetchedAppointment.Status,
		"Appointment status should be 'scheduled'",
	)
	assert.NotNil(fetchedAppointment.Equipment)
	assert.Len(*fetchedAppointment.Equipment, 1)
	assert.Equal(*resource.Id, (*fetchedAppointment.Equipment)[0].Id)
	assert.Equal(resource.Name, (*fetchedAppointment.Equipment)[0].Name)
}

func TestDecideAppointmentReject(t *testing.T) {
	t.Parallel()

	patientEmail := fmt.Sprintf("test.decide.appt.reject.%s@patient.com", uuid.NewString())
	patient := mustCreatePatient(t, newPatient(patientEmail))

	doctorEmail := fmt.Sprintf("test.decide.appt.reject.%s@doctor.com", uuid.NewString())
	doctor := mustCreateDoctor(t, newDoctor(doctorEmail))

	appointmentTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	newAppointmentReq := api.NewAppointmentRequest{
		PatientId:           patient.Id,
		DoctorId:            doctor.Id,
		AppointmentDateTime: appointmentTime,
		Type:                asPtr(api.RegularCheck),
	}
	createdAppointment := mustCreateAppointment(t, newAppointmentReq)
	appointmentId := *createdAppointment.Id

	rejectionReason := "Test rejection reason"
	decision := api.AppointmentDecision{
		Action: api.Reject,
		Reason: &rejectionReason,
	}
	decisionReqBody, err := json.Marshal(decision)
	require.NoError(t, err, "Failed to marshal decision request body")

	url := fmt.Sprintf("%s/appointments/%s", ServerUrl, appointmentId)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(decisionReqBody))
	require.NoError(t, err, "Failed to create HTTP POST request")
	req.Header.Set("Content-Type", server.ApplicationJSON)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "http.Post failed for DecideAppointment")
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode, "Expected '200 OK' status code")

	var fetchedAppointment api.DoctorAppointment
	err = json.NewDecoder(res.Body).Decode(&fetchedAppointment)
	require.NoError(t, err, "Failed to decode fetched appointment")

	assert := assert.New(t)
	assert.Equal(api.Denied, fetchedAppointment.Status, "Appointment status should be 'denied'")
	assert.Equal(rejectionReason, *fetchedAppointment.Reason, "Rejection reason mismatch")
}

func TestDoctorsAppointmentById(t *testing.T) {
	t.Parallel()

	patientEmail := fmt.Sprintf("test.doctor.appt.by.id.%s@patient.com", uuid.NewString())
	patient := mustCreatePatient(t, newPatient(patientEmail))

	doctorEmail := fmt.Sprintf("test.doctor.appt.by.id.%s@doctor.com", uuid.NewString())
	doctor := mustCreateDoctor(t, newDoctor(doctorEmail))

	appointmentTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	newAppointmentReq := api.NewAppointmentRequest{
		PatientId:           patient.Id,
		DoctorId:            doctor.Id,
		AppointmentDateTime: appointmentTime,
		Type:                asPtr(api.RegularCheck),
	}
	createdAppointment := mustCreateAppointment(t, newAppointmentReq)
	appointmentId := *createdAppointment.Id

	resourceName := "Test Resource"
	resourceType := api.ResourceTypeEquipment
	resource := mustCreateResource(t, api.NewResource{Name: resourceName, Type: resourceType})
	mustCreateReservation(
		t,
		api.ResourceReservation{
			AppointmentId: appointmentId,
			Start:         appointmentTime,
			End:           appointmentTime.Add(time.Hour),
		},
		*resource.Id,
	)

	url := fmt.Sprintf("%s/doctors/%s/appointment/%s", ServerUrl, doctor.Id, appointmentId)
	res, err := http.Get(url)
	require.NoError(t, err, "http.Get failed for DoctorsAppointmentById")
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode, "Expected '200 OK' status code")

	var fetchedAppointment api.DoctorAppointment
	err = json.NewDecoder(res.Body).Decode(&fetchedAppointment)
	require.NoError(t, err, "Failed to decode fetched appointment")

	assert := assert.New(t)
	assert.True(
		appointmentTime.Equal(fetchedAppointment.AppointmentDateTime),
		"Appointment date time mismatch",
	)
	assert.Nil(
		fetchedAppointment.CancellationReason,
		"Cancellation reason should be nil for a new appointment",
	)
	assert.Equal(patient.Id, fetchedAppointment.Patient.Id, "Patient ID mismatch")
	assert.Equal(appointmentId, *fetchedAppointment.Id, "Appointment ID mismatch")
	assert.Equal(
		newAppointmentReq.Reason,
		fetchedAppointment.Reason,
		"Appointment reason mismatch",
	)
	assert.Equal(
		api.Requested,
		fetchedAppointment.Status,
		"Appointment status should be 'requested'",
	)
	assert.Equal(*newAppointmentReq.Type, fetchedAppointment.Type, "Appointment type mismatch")

	require.NotNil(t, fetchedAppointment.Equipment)
	assert.Len(*fetchedAppointment.Equipment, 1)
	assert.Equal(*resource.Id, (*fetchedAppointment.Equipment)[0].Id)
	assert.Equal(resource.Name, (*fetchedAppointment.Equipment)[0].Name)
}

func TestPatientsAppointmentById(t *testing.T) {
	t.Parallel()

	patientEmail := fmt.Sprintf("test.patient.appt.by.id.%s@patient.com", uuid.NewString())
	patient := mustCreatePatient(t, newPatient(patientEmail))

	doctorEmail := fmt.Sprintf("test.patient.appt.by.id.%s@doctor.com", uuid.NewString())
	doctor := mustCreateDoctor(t, newDoctor(doctorEmail))

	appointmentTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	newAppointmentReq := api.NewAppointmentRequest{
		PatientId:           patient.Id,
		DoctorId:            doctor.Id,
		AppointmentDateTime: appointmentTime,
		Type:                asPtr(api.RegularCheck),
	}
	createdAppointment := mustCreateAppointment(t, newAppointmentReq)
	appointmentId := *createdAppointment.Id

	url := fmt.Sprintf("%s/patients/%s/appointment/%s", ServerUrl, patient.Id, appointmentId)
	res, err := http.Get(url)
	require.NoError(t, err, "http.Get failed for PatientsAppointmentById")
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode, "Expected '200 OK' status code")

	var fetchedAppointment api.PatientAppointment
	err = json.NewDecoder(res.Body).Decode(&fetchedAppointment)
	require.NoError(t, err, "Failed to decode fetched appointment")

	assert := assert.New(t)
	assert.True(
		appointmentTime.Equal(fetchedAppointment.AppointmentDateTime),
		"Appointment date time mismatch",
	)
	assert.Nil(
		fetchedAppointment.CancellationReason,
		"Cancellation reason should be nil for a new appointment",
	)
	assert.Nil(fetchedAppointment.Condition, "Condition should be nil if not provided")
	assert.Equal(doctor.Id, fetchedAppointment.Doctor.Id, "Doctor ID mismatch")
	assert.Equal(appointmentId, *fetchedAppointment.Id, "Appointment ID mismatch")
	assert.Equal(
		newAppointmentReq.Reason,
		fetchedAppointment.Reason,
		"Appointment reason mismatch",
	)
	assert.Equal(
		api.Requested,
		fetchedAppointment.Status,
		"Appointment status should be 'scheduled'",
	)
	assert.Equal(*newAppointmentReq.Type, fetchedAppointment.Type, "Appointment type mismatch")
}

func TestCancelAppointment(t *testing.T) {
	t.Parallel()

	patientEmail := fmt.Sprintf("test.cancel.appt.%s@patient.com", uuid.NewString())
	patient := mustCreatePatient(t, newPatient(patientEmail))

	doctorEmail := fmt.Sprintf("test.cancel.appt.%s@doctor.com", uuid.NewString())
	doctor := mustCreateDoctor(t, newDoctor(doctorEmail))

	appointmentTime := time.Now().Add(24 * time.Hour).Truncate(time.Hour)
	newAppointmentReq := api.NewAppointmentRequest{
		PatientId:           patient.Id,
		DoctorId:            doctor.Id,
		AppointmentDateTime: appointmentTime,
	}
	createdAppointment := mustCreateAppointment(t, newAppointmentReq)
	appointmentId := *createdAppointment.Id

	cancellationReason := "Test cancellation reason"
	cancellationReqBody, err := json.Marshal(
		api.AppointmentCancellation{Reason: &cancellationReason},
	)
	require.NoError(t, err, "Failed to marshal cancellation request body")

	url := fmt.Sprintf("%s/appointments/%s", ServerUrl, appointmentId)
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(cancellationReqBody))
	require.NoError(t, err, "Failed to create HTTP DELETE request")
	req.Header.Set("Content-Type", server.ApplicationJSON)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "http.Delete failed for CancelAppointment")
	defer res.Body.Close()

	require.Equal(t, http.StatusNoContent, res.StatusCode, "Expected '204 No Content' status code")

	getUrl := fmt.Sprintf("%s/patients/%s/appointment/%s", ServerUrl, patient.Id, appointmentId)
	getRes, err := http.Get(getUrl)
	require.NoError(t, err, "Failed to fetch appointment after cancellation")
	defer getRes.Body.Close()

	require.Equal(
		t,
		http.StatusOK,
		getRes.StatusCode,
		"Expected '200 OK' status code for fetched appointment",
	)

	var fetchedAppointment api.PatientAppointment
	err = json.NewDecoder(getRes.Body).Decode(&fetchedAppointment)
	require.NoError(t, err, "Failed to decode fetched appointment")

	assert.Equal(
		t,
		api.Cancelled,
		fetchedAppointment.Status,
		"Appointment status should be 'cancelled'",
	)
	assert.Equal(
		t,
		cancellationReason,
		*fetchedAppointment.CancellationReason,
		"Cancellation reason mismatch",
	)
}

func mustCreateAppointment(t *testing.T, request api.NewAppointmentRequest) api.PatientAppointment {
	t.Helper()
	require := require.New(t)

	reqBodyBytes, err := json.Marshal(request)
	require.NoError(err, "mustCreateAppointment: Failed to marshal request")

	url := fmt.Sprintf("%s/appointments", ServerUrl)
	res, err := http.Post(url, server.ApplicationJSON, bytes.NewBuffer(reqBodyBytes))
	require.NoError(err, "mustCreateAppointment: http.Post failed")
	defer res.Body.Close()

	bodyBytes, readErr := io.ReadAll(res.Body)
	require.NoError(readErr, "mustCreateAppointment: Failed to read response body")
	res.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	require.Equal(
		http.StatusCreated,
		res.StatusCode,
		"mustCreateAppointment: Expected '201 Created'. Body: %s",
		string(bodyBytes),
	)

	var createdAppointment api.PatientAppointment
	err = json.NewDecoder(res.Body).Decode(&createdAppointment)
	require.NoError(err, "mustCreateAppointment: Failed to decode response")
	require.NotNil(createdAppointment.Id, "mustCreateAppointment: Response ID is nil")
	require.NotEmpty(*createdAppointment.Id, "mustCreateAppointment: Response ID is empty UUID")

	return createdAppointment
}

func mustCreatePrescription(t *testing.T, request api.NewPrescription) api.PrescriptionDisplay {
	t.Helper()
	require := require.New(t)

	reqBodyBytes, err := json.Marshal(request)
	require.NoError(err, "mustCreatePrescription: Failed to marshal request")

	url := fmt.Sprintf("%s/prescriptions", ServerUrl)
	res, err := http.Post(url, server.ApplicationJSON, bytes.NewBuffer(reqBodyBytes))
	require.NoError(err, "mustCreatePrescription: http.Post failed")
	defer res.Body.Close()

	bodyBytes, readErr := io.ReadAll(res.Body)
	require.NoError(readErr, "mustCreatePrescription: Failed to read response body")
	res.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	require.Equal(
		http.StatusCreated,
		res.StatusCode,
		"mustCreatePrescription: Expected '201 Created'. Body: %s",
		string(bodyBytes),
	)

	var createdPrescription api.PrescriptionDisplay
	err = json.NewDecoder(res.Body).Decode(&createdPrescription)
	require.NoError(err, "mustCreatePrescription: Failed to decode response")
	require.NotNil(createdPrescription.Id, "mustCreatePrescription: Response ID is nil")
	require.NotEmpty(*createdPrescription.Id, "mustCreatePrescription: Response ID is empty UUID")

	return createdPrescription
}
