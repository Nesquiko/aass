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

func dataPrescToPrescDisplay(p Prescription) api.PrescriptionDisplay {
	return api.PrescriptionDisplay{
		Id:            &p.Id,
		Name:          p.Name,
		Start:         p.Start,
		End:           p.End,
		AppointmentId: p.AppointmentId,
	}
}

func dataPrescSliceToPrescDisplaySlice(prescs []Prescription) []api.PrescriptionDisplay {
	apiPrescs := make([]api.PrescriptionDisplay, len(prescs))
	for i, p := range prescs {
		apiPrescs[i] = dataPrescToPrescDisplay(p)
	}
	return apiPrescs
}
