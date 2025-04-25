package pkg

import "github.com/Nesquiko/aass/resource-service/api"

func resourceToEquipment(r Resource) api.Equipment {
	return api.Equipment{Id: r.Id, Name: r.Name}
}

func resourceToFacility(r Resource) api.Facility {
	return api.Facility{Id: r.Id, Name: r.Name}
}

func resourceToMedicine(r Resource) api.Medicine {
	return api.Medicine{Id: r.Id, Name: r.Name}
}

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
