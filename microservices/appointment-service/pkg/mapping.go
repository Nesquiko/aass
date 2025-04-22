package pkg

import (
	"github.com/Nesquiko/aass/appointment-service/api"
	doctorapi "github.com/Nesquiko/aass/appointment-service/api/doctor-api"
	patientapi "github.com/Nesquiko/aass/appointment-service/api/patient-api"
)

// --- mapAppointmentToPatientAppointment function (Corrected) ---
func mapAppointmentToPatientAppointment(
	appt Appointment,
	doc doctorapi.Doctor,
) api.PatientAppointment {
	// Directly map fields from pkg.Appointment to api.PatientAppointment
	// Provide zero-value for the required Doctor field.
	apiAppt := api.PatientAppointment{
		Id:                  &appt.Id,
		AppointmentDateTime: appt.AppointmentDateTime,
		Status:              api.AppointmentStatus(appt.Status),
		Type:                api.AppointmentType(appt.Type),
		Reason:              appt.Reason,
		CancellationReason:  appt.CancellationReason,
		DenialReason:        appt.DenialReason,
		Doctor: api.Doctor{
			Id:             *doc.Id,
			Email:          doc.Email,
			FirstName:      doc.FirstName,
			LastName:       doc.LastName,
			Role:           (*api.DoctorRole)(doc.Role),
			Specialization: api.SpecializationEnum(doc.Specialization),
		},
	}

	if appt.CancelledBy != nil {
		role := api.UserRole(*appt.CancelledBy)
		if role == api.UserRoleDoctor || role == api.UserRolePatient {
			apiAppt.CanceledBy = &role
		}
	}

	return apiAppt
}

func mapAppointmentToDoctorAppointment(
	appt Appointment,
	patient patientapi.Patient,
) api.DoctorAppointment {
	apiAppt := api.DoctorAppointment{
		Id:                  &appt.Id,
		AppointmentDateTime: appt.AppointmentDateTime,
		Status:              api.AppointmentStatus(appt.Status),
		Type:                api.AppointmentType(appt.Type),
		Reason:              appt.Reason,
		CancellationReason:  appt.CancellationReason,
		DenialReason:        appt.DenialReason,
		Patient: api.Patient{
			Id:        *patient.Id,
			Email:     patient.Email,
			FirstName: patient.FirstName,
			LastName:  patient.LastName,
			Role:      (*api.PatientRole)(patient.Role),
		},
	}

	if appt.CancelledBy != nil {
		role := api.UserRole(*appt.CancelledBy)
		if role == api.UserRoleDoctor || role == api.UserRolePatient {
			apiAppt.CanceledBy = &role
		}
	}
	return apiAppt
}
