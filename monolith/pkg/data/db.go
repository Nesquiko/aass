package data

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDuplicateEmail      = errors.New("email address already exists")
	ErrNotFound            = errors.New("resource not found")
	ErrDoctorUnavailable   = errors.New("doctor unavailable at the specified time")
	ErrResourceUnavailable = errors.New("resource is unavailable during the requested time slot")
)

type Db interface {
	Disconnect(ctx context.Context) error

	CreateAppointment(ctx context.Context, appointment Appointment) (Appointment, error)
	AppointmentById(ctx context.Context, id uuid.UUID) (Appointment, error)
	AppointmentsByDoctorId(
		ctx context.Context,
		doctorId uuid.UUID,
		from time.Time,
		to *time.Time,
	) ([]Appointment, error)
	AppointmentsByPatientId(
		ctx context.Context,
		patientId uuid.UUID,
		from time.Time,
		to *time.Time,
	) ([]Appointment, error)
	CancelAppointment(
		ctx context.Context,
		appointmentId uuid.UUID,
		by string,
		cancellationReason *string,
	) error
	DecideAppointment(
		ctx context.Context,
		appointmentId uuid.UUID,
		decision string,
		denyReason *string,
		resources []Resource,
	) (Appointment, error)
	AppointmentsByDoctorIdAndDate(
		ctx context.Context,
		doctorId uuid.UUID,
		date time.Time,
	) ([]Appointment, error)
	RescheduleAppointment(
		ctx context.Context,
		appointmentId uuid.UUID,
		newDateTime time.Time,
	) (Appointment, error)
	AppointmentsByConditionId(ctx context.Context, conditionId uuid.UUID) ([]Appointment, error)

	CreatePatient(ctx context.Context, patient Patient) (Patient, error)
	PatientById(ctx context.Context, id uuid.UUID) (Patient, error)
	PatientByEmail(ctx context.Context, email string) (Patient, error)

	CreateDoctor(ctx context.Context, doctor Doctor) (Doctor, error)
	DoctorById(ctx context.Context, id uuid.UUID) (Doctor, error)
	DoctorByEmail(ctx context.Context, email string) (Doctor, error)
	AvailableDoctors(ctx context.Context, dateTime time.Time) ([]Doctor, error)
	GetAllDoctors(ctx context.Context) ([]Doctor, error)

	CreateCondition(ctx context.Context, condition Condition) (Condition, error)
	ConditionById(ctx context.Context, id uuid.UUID) (Condition, error)
	FindConditionsByPatientId(
		ctx context.Context,
		patientId uuid.UUID,
		from time.Time,
		to *time.Time,
	) ([]Condition, error)
	UpdateCondition(ctx context.Context, id uuid.UUID, condition Condition) (Condition, error)
	FindConditionsByPatientIdAndDate(
		ctx context.Context,
		patientId uuid.UUID,
		date time.Time,
	) ([]Condition, error)

	CreatePrescription(ctx context.Context, prescription Prescription) (Prescription, error)
	PrescriptionById(ctx context.Context, id uuid.UUID) (Prescription, error)
	FindPrescriptionsByPatientId(
		ctx context.Context,
		patientId uuid.UUID,
		from time.Time,
		to *time.Time,
	) ([]Prescription, error)
	UpdatePrescription(
		ctx context.Context,
		id uuid.UUID,
		prescription Prescription,
	) (Prescription, error)
	PrescriptionByAppointmentId(
		ctx context.Context,
		appointmentId uuid.UUID,
	) ([]Prescription, error)
	DeletePrescription(ctx context.Context, id uuid.UUID) error

	CreateResource(ctx context.Context, name string, typ ResourceType) (Resource, error)
	ResourceById(ctx context.Context, id uuid.UUID) (Resource, error)
	FindAvailableResourcesAtTime(
		ctx context.Context,
		appointmentDate time.Time,
	) (struct {
		Medicines  []Resource
		Facilities []Resource
		Equipment  []Resource
	}, error)
	CreateReservation(
		ctx context.Context,
		appointmentId uuid.UUID,
		resourceId uuid.UUID,
		resourceName string,
		resourceType ResourceType,
		startTime time.Time,
		endTime time.Time,
	) (Reservation, error)
	ResourcesByAppointmentId(ctx context.Context, appointmentId uuid.UUID) ([]Resource, error)
}

type PaginationResult struct {
	Total    int64
	Page     int
	PageSize int
}
