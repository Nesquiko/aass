package pkg

import (
	"log/slog"

	"github.com/Nesquiko/aass/appointment-service/api"
)

// --- mapAppointmentToPatientAppointment function (Corrected) ---
func mapAppointmentToPatientAppointment(appt Appointment) api.PatientAppointment {
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
		Doctor:              api.Doctor{}, // Provide zero-value for required Doctor
		// Condition: // Omitted for MVP - map if needed
		// Prescriptions: // Omitted for MVP - map if needed
	}

	// Handle optional CanceledBy field
	if appt.CancelledBy != nil {
		role := api.UserRole(*appt.CancelledBy)
		if role == api.UserRoleDoctor || role == api.UserRolePatient {
			apiAppt.CanceledBy = &role
		} else {
			slog.Warn("Unknown role found for cancelledBy", "role", *appt.CancelledBy, "appointmentId", appt.Id)
		}
	}

	// Map Condition if available and needed for the view
	// if appt.ConditionId != nil {
	//    apiAppt.Condition = &api.ConditionDisplay{ Id: *appt.ConditionId, Name: "Fetched Condition Name" } // Placeholder
	// }

	// Map Prescriptions if available and needed for the view
	// apiAppt.Prescriptions = &[]api.PrescriptionDisplay{} // Placeholder

	return apiAppt
}

// --- mapAppointmentsToPatientAppointmentSlice function (Corrected) ---
func mapAppointmentsToPatientAppointmentSlice(appts []Appointment) []api.PatientAppointment {
	apiAppts := make([]api.PatientAppointment, len(appts))
	for i, appt := range appts {
		apiAppts[i] = mapAppointmentToPatientAppointment(appt) // Use the corrected mapper
	}
	return apiAppts
}

// --- mapAppointmentToDoctorAppointment function (Corrected) ---
func mapAppointmentToDoctorAppointment(appt Appointment) api.DoctorAppointment {
	// Directly map fields from pkg.Appointment to api.DoctorAppointment
	// Provide zero-value for the required Patient field.
	// Omit resource fields for MVP.
	apiAppt := api.DoctorAppointment{
		Id:                  &appt.Id,
		AppointmentDateTime: appt.AppointmentDateTime,
		Status:              api.AppointmentStatus(appt.Status),
		Type:                api.AppointmentType(appt.Type),
		Reason:              appt.Reason,
		CancellationReason:  appt.CancellationReason,
		DenialReason:        appt.DenialReason,
		Patient:             api.Patient{}, // Provide zero-value for required Patient
		// Condition: // Omitted for MVP - map if needed
		// Prescriptions: // Omitted for MVP - map if needed
		// Facilities: // Omitted for MVP
		// Equipment:  // Omitted for MVP
		// Medicine:   // Omitted for MVP
	}

	// Handle optional CanceledBy field
	if appt.CancelledBy != nil {
		role := api.UserRole(*appt.CancelledBy)
		if role == api.UserRoleDoctor || role == api.UserRolePatient {
			apiAppt.CanceledBy = &role
		} else {
			slog.Warn("Unknown role found for cancelledBy", "role", *appt.CancelledBy, "appointmentId", appt.Id)
		}
	}

	// Map Condition if available and needed for the view
	// if appt.ConditionId != nil {
	//    apiAppt.Condition = &api.ConditionDisplay{ Id: *appt.ConditionId, Name: "Fetched Condition Name" } // Placeholder
	// }

	// Map Prescriptions if available and needed for the view
	// apiAppt.Prescriptions = &[]api.PrescriptionDisplay{} // Placeholder

	// Map Resources if available and needed for the view
	// apiAppt.Facilities = &[]api.Facility{} // Placeholder
	// apiAppt.Equipment = &[]api.Equipment{} // Placeholder
	// apiAppt.Medicine = &[]api.Medicine{} // Placeholder

	return apiAppt
}

// --- mapAppointmentsToDoctorAppointmentSlice function (Corrected) ---
func mapAppointmentsToDoctorAppointmentSlice(appts []Appointment) []api.DoctorAppointment {
	apiAppts := make([]api.DoctorAppointment, len(appts))
	for i, appt := range appts {
		apiAppts[i] = mapAppointmentToDoctorAppointment(appt) // Use the corrected mapper
	}
	return apiAppts
}
