package pkg

import (
	"github.com/google/uuid"

	"github.com/Nesquiko/aass/condition-service/api"
)

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

func newCondToDataCond(c api.NewCondition) Condition {
	return Condition{
		PatientId: c.PatientId,
		Name:      c.Name,
		Start:     c.Start,
		End:       c.End,
	}
}

func dataCondToCondDisplay(c Condition) api.ConditionDisplay {
	return api.ConditionDisplay{
		Id:    &c.Id,
		Name:  c.Name,
		Start: c.Start,
		End:   c.End,
	}
}

func dataCondToCond(c Condition) api.Condition {
	return api.Condition{
		Id:              &c.Id,
		Name:            c.Name,
		Start:           c.Start,
		End:             c.End,
		Appointments:    []api.AppointmentDisplay{},
		AppointmentsIds: &[]uuid.UUID{},
	}
}
