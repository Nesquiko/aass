package main

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

func dataResourcesToApiResources(resources struct {
	Medicines  []Resource
	Facilities []Resource
	Equipment  []Resource
},
) api.AvailableResources {
	available := api.AvailableResources{
		Equipment:  make([]api.Equipment, len(resources.Equipment)),
		Facilities: make([]api.Facility, len(resources.Facilities)),
		Medicine:   make([]api.Medicine, len(resources.Medicines)),
	}

	for i, res := range resources.Equipment {
		available.Equipment[i] = resourceToEquipment(res)
	}

	for i, res := range resources.Facilities {
		available.Facilities[i] = resourceToFacility(res)
	}

	for i, res := range resources.Medicines {
		available.Medicine[i] = resourceToMedicine(res)
	}

	return available
}
