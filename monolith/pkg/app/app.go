package app

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/Nesquiko/wac/pkg/api"
	"github.com/Nesquiko/wac/pkg/data"
)

var (
	ErrDuplicateEmail      = errors.New("email address already exists")
	ErrNotFound            = errors.New("resource not found")
	ErrDoctorUnavailable   = errors.New("doctor unavailable at the specified time")
	ErrResourceUnavailable = errors.New("resource is unavailable during the requested time slot")
)

type App interface {
	CreateAppointment(
		ctx context.Context,
		appt api.NewAppointmentRequest,
	) (api.PatientAppointment, error)
	CancelAppointment(
		ctx context.Context,
		appointmentId uuid.UUID,
		req api.AppointmentCancellation,
	) error
	PatientsAppointmentById(
		ctx context.Context,
		patientId uuid.UUID,
		appointmentId uuid.UUID,
	) (api.PatientAppointment, error)
	DoctorsAppointmentById(
		ctx context.Context,
		doctorId uuid.UUID,
		appointmentId uuid.UUID,
	) (api.DoctorAppointment, error)
	DecideAppointment(
		ctx context.Context,
		appointmentId uuid.UUID,
		decision api.AppointmentDecision,
	) (api.DoctorAppointment, error)
	RescheduleAppointment(
		ctx context.Context,
		appointmentId api.AppointmentId,
		newDateTime time.Time,
	) (api.PatientAppointment, error)

	CreatePatient(ctx context.Context, p api.PatientRegistration) (api.Patient, error)
	PatientById(ctx context.Context, id uuid.UUID) (api.Patient, error)
	PatientByEmail(ctx context.Context, email string) (api.Patient, error)
	PatientsCalendar(
		ctx context.Context,
		patientId uuid.UUID,
		from api.From,
		to *api.To,
	) (api.PatientsCalendar, error)
	PatientMedicalHistoryFiles(
		ctx context.Context,
		patientId uuid.UUID,
		page int,
		pageSize int,
	) (api.MedicalHistoryFileList, error)

	CreateDoctor(ctx context.Context, d api.DoctorRegistration) (api.Doctor, error)
	DoctorById(ctx context.Context, id uuid.UUID) (api.Doctor, error)
	DoctorByEmail(ctx context.Context, email string) (api.Doctor, error)
	DoctorsCalendar(
		ctx context.Context,
		doctorId uuid.UUID,
		from api.From,
		to *api.To,
	) (api.DoctorCalendar, error)
	DoctorTimeSlots(
		ctx context.Context,
		doctorId uuid.UUID,
		date time.Time,
	) (api.DoctorTimeslots, error)
	AvailableDoctors(ctx context.Context, dateTime time.Time) ([]api.Doctor, error)
	GetAllDoctors(ctx context.Context) ([]api.Doctor, error)

	CreatePatientCondition(ctx context.Context, cond api.NewCondition) (api.ConditionDisplay, error)
	ConditionById(ctx context.Context, id uuid.UUID) (api.Condition, error)
	UpdatePatientCondition(
		ctx context.Context,
		conditionId uuid.UUID,
		updateData api.UpdateCondition,
	) (api.Condition, error)
	PatientConditionsOnDate(
		ctx context.Context,
		patientId uuid.UUID,
		date time.Time,
	) ([]api.ConditionDisplay, error)

	CreatePatientPrescription(
		ctx context.Context,
		pres api.NewPrescription,
	) (api.Prescription, error)
	UpdatePatientPrescription(
		ctx context.Context,
		prescriptionId uuid.UUID,
		updateData api.UpdatePrescription,
	) (api.Prescription, error)
	PrescriptionById(ctx context.Context, prescriptionId uuid.UUID) (api.Prescription, error)
	DeletePrescription(ctx context.Context, id uuid.UUID) error

	CreateResource(ctx context.Context, resource api.NewResource) (api.NewResource, error)
	ReserveResource(
		ctx context.Context,
		resourceId uuid.UUID,
		reservation api.ResourceReservation,
	) error
	AvailableResources(ctx context.Context, dateTime time.Time) (api.AvailableResources, error)
	ReserveAppointmentResources(
		ctx context.Context,
		appointmentId uuid.UUID,
		payload api.ReserveAppointmentResourcesJSONBody,
	) (api.DoctorAppointment, error)
}

func New(db data.Db) App {
	return monolithApp{db}
}

type monolithApp struct {
	db data.Db
}
