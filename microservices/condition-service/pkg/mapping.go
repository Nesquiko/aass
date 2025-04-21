package pkg

import (
	"time"

	"github.com/google/uuid"
	"github.com/oapi-codegen/nullable"

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
	var end *time.Time = nil
	if c.End.IsSpecified() {
		end = asPtr(c.End.MustGet())
	}
	return Condition{
		PatientId: c.PatientId,
		Name:      c.Name,
		Start:     c.Start,
		End:       end,
	}
}

func dataCondToCondDisplay(c Condition) api.ConditionDisplay {
	end := nullable.NewNullNullable[time.Time]()
	if c.End != nil {
		end.Set(*c.End)
	}
	return api.ConditionDisplay{
		Id:    &c.Id,
		Name:  c.Name,
		Start: c.Start,
		End:   end,
	}
}

func dataCondToCond(c Condition) api.Condition {
	end := nullable.NewNullNullable[time.Time]()
	if c.End != nil {
		end.Set(*c.End)
	}
	return api.Condition{
		Id:              &c.Id,
		Name:            c.Name,
		Start:           c.Start,
		End:             end,
		Appointments:    &[]api.AppointmentDisplay{},
		AppointmentsIds: &[]uuid.UUID{},
	}
}
