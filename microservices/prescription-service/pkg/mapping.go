package pkg

import "github.com/Nesquiko/aass/prescription-service/api"

func asPtr[T any](v T) *T {
	return &v
}

func Map[T, V any](ts []T, fn func(T) V) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func newPrescToDataPresc(p api.NewPrescription) Prescription {
	return Prescription{
		PatientId:     p.PatientId,
		Name:          p.Name,
		Start:         p.Start,
		End:           p.End,
		DoctorsNote:   p.DoctorsNote,
		AppointmentId: p.AppointmentId,
	}
}

func dataPrescToPresc(p Prescription) api.Prescription {
	presc := api.Prescription{
		Id:            &p.Id,
		Name:          p.Name,
		Start:         p.Start,
		End:           p.End,
		DoctorsNote:   p.DoctorsNote,
		AppointmentId: p.AppointmentId,
	}
	return presc
}
