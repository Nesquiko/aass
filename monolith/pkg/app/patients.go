package app

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/Nesquiko/wac/pkg/api"
	"github.com/Nesquiko/wac/pkg/data"
)

func (a monolithApp) CreatePatient(
	ctx context.Context,
	p api.PatientRegistration,
) (api.Patient, error) {
	patient := patientRegToDataPatient(p)

	patient, err := a.db.CreatePatient(ctx, patient)
	if errors.Is(err, data.ErrDuplicateEmail) {
		return api.Patient{}, fmt.Errorf("CreatePatient duplicate emall: %w", ErrDuplicateEmail)
	} else if err != nil {
		return api.Patient{}, fmt.Errorf("CreatePatient: %w", err)
	}

	return dataPatientToApiPatient(patient), nil
}

func (a monolithApp) PatientById(ctx context.Context, id uuid.UUID) (api.Patient, error) {
	patient, err := a.db.PatientById(ctx, id)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return api.Patient{}, fmt.Errorf("PatientById: %w", ErrNotFound)
		}
		return api.Patient{}, fmt.Errorf("PatientById: %w", err)
	}

	return dataPatientToApiPatient(patient), nil
}

func (a monolithApp) PatientByEmail(ctx context.Context, email string) (api.Patient, error) {
	patient, err := a.db.PatientByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, data.ErrNotFound) {
			return api.Patient{}, fmt.Errorf("PatientByEmail: %w", ErrNotFound)
		}
		return api.Patient{}, fmt.Errorf("PatientByEmail: %w", err)
	}

	return dataPatientToApiPatient(patient), nil
}

func (a monolithApp) PatientsCalendar(
	ctx context.Context,
	patientId uuid.UUID,
	from api.From,
	to *api.To,
) (api.PatientsCalendar, error) {
	var toTime *time.Time = nil
	if to != nil {
		toTime = &to.Time
	}

	appts, err := a.db.AppointmentsByPatientId(ctx, patientId, from.Time, toTime)
	if err != nil {
		return api.PatientsCalendar{}, fmt.Errorf("PatientsCalendar appointments: %w", err)
	}

	conds, err := a.db.FindConditionsByPatientId(ctx, patientId, from.Time, toTime)
	if err != nil {
		return api.PatientsCalendar{}, fmt.Errorf("PatientsCalendar conditions: %w", err)
	}

	prescriptions, err := a.db.FindPrescriptionsByPatientId(ctx, patientId, from.Time, toTime)
	if err != nil {
		return api.PatientsCalendar{}, fmt.Errorf("PatientsCalendar prescriptions: %w", err)
	}

	calendar := api.PatientsCalendar{}
	var doctor *data.Doctor = nil
	if len(appts) != 0 {
		calendar.Appointments = asPtr(make([]api.AppointmentDisplay, len(appts)))

		d, err := a.db.DoctorById(ctx, appts[0].DoctorId)
		if err != nil {
			return api.PatientsCalendar{}, fmt.Errorf("PatientsCalendar doc find: %w", err)
		}
		doctor = &d
	}

	if len(conds) != 0 {
		calendar.Conditions = asPtr(make([]api.ConditionDisplay, len(conds)))
	}
	if len(prescriptions) != 0 {
		calendar.Prescriptions = asPtr(make([]api.PrescriptionDisplay, len(prescriptions)))
	}

	for i, appt := range appts {
		patient, err := a.db.PatientById(ctx, appt.PatientId)
		if err != nil {
			return api.PatientsCalendar{}, fmt.Errorf("PatientsCalendar patient find: %w", err)
		}
		(*calendar.Appointments)[i] = dataApptToApptDisplay(appt, patient, *doctor)
	}

	for i, cond := range conds {
		(*calendar.Conditions)[i] = dataCondToCondDisplay(cond)
	}

	for i, presc := range prescriptions {
		(*calendar.Prescriptions)[i] = dataPrescToPrescDisplay(presc)
	}

	return calendar, nil
}

func (a monolithApp) PatientMedicalHistoryFiles(
	ctx context.Context,
	patientId uuid.UUID,
	page int,
	pageSize int,
) (api.MedicalHistoryFileList, error) {
	files := make([]string, 0)
	for i := range pageSize {
		fileType := dummyFileTypes[rand.Intn(len(dummyFileTypes))]
		date := time.Now().AddDate(0, 0, -rand.Intn(365)).Format("2006-01-02")
		files = append(files, fmt.Sprintf("%s_%s_%d.pdf", fileType, date, i))
	}

	pagination := api.Pagination{
		Page:     page,
		PageSize: pageSize,
		Total:    38,
	}

	return api.MedicalHistoryFileList{
		Files:      files,
		Pagination: pagination,
	}, nil
}

var dummyFileTypes = []string{
	"lab_result",
	"medical_report",
	"prescription",
	"discharge_summary",
	"consultation_note",
	"radiology_report",
	"pathology_report",
	"medical_certificate",
	"referral_letter",
	"follow_up_note",
	"progress_note",
	"medication_list",
	"allergy_report",
	"immunization_record",
	"vital_signs_chart",
	"medical_history",
	"family_medical_history",
	"surgical_report",
	"anesthesia_report",
	"recovery_room_note",
	"physical_examination_report",
	"diagnostic_test_result",
	"treatment_plan",
	"care_plan",
	"discharge_plan",
}
