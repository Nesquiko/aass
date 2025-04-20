import {
  DoctorAppointment,
  Equipment,
  Facility,
  instanceOfDoctorAppointment,
  Medicine,
  PatientAppointment,
} from '../../../api/generated';
import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-appointment-resources',
  shadow: false,
})
export class AppointmentResources {
  @Prop() isDoctor: boolean;
  @Prop() appointment: PatientAppointment | DoctorAppointment;
  @Prop() editingResources: boolean;
  @Prop() setEditingResources: (editingResources: boolean) => void;

  @Prop() availableFacilities: Array<Facility>;
  @Prop() selectedFacility: Facility;
  @Prop() handleSelectFacility: (event: Event) => void;

  @Prop() availableEquipment: Array<Equipment>;
  @Prop() selectedEquipment: Equipment;
  @Prop() handleSelectEquipment: (event: Event) => void;

  @Prop() availableMedicine: Array<Medicine>;
  @Prop() selectedMedicine: Medicine;
  @Prop() handleSelectMedicine: (event: Event) => void;

  @Prop() handleSaveResources: () => void;

  render() {
    return (
      <div class="relative mb-6 w-full max-w-md rounded-md bg-gray-200 px-4 py-3">
        <div class="mb-1 flex flex-row items-center justify-start gap-x-2 text-gray-500">
          <md-icon style={{ fontSize: '16px' }}>library_add</md-icon>
          Resources
        </div>

        {['scheduled', 'requested'].includes(this.appointment.status) && (
          <md-icon-button
            title="Edit resources"
            class="absolute top-1 left-32"
            onClick={() => this.setEditingResources(!this.editingResources)}
          >
            <md-icon style={{ fontSize: '16px' }}>edit</md-icon>
          </md-icon-button>
        )}

        <div class="flex w-full flex-row items-center justify-between">
          <div class="flex flex-row items-center gap-x-2 text-gray-500">
            <md-icon style={{ fontSize: '16px' }}>meeting_room</md-icon>
            Facility
          </div>
          <span class="font-medium text-gray-600">
            {instanceOfDoctorAppointment(this.appointment) &&
              (this.appointment.equipment?.[0].name ?? '')}
          </span>
        </div>
        <div class="flex w-full flex-row items-center justify-between">
          <div class="flex flex-row items-center gap-x-2 text-gray-500">
            <md-icon style={{ fontSize: '16px' }}>service_toolbox</md-icon>
            Equipment
          </div>
          <span class="font-medium text-gray-600">
            {instanceOfDoctorAppointment(this.appointment) &&
              (this.appointment.equipment?.[0].name ?? '')}
          </span>
        </div>
        <div class="flex w-full flex-row items-center justify-between">
          <div class="flex flex-row items-center gap-x-2 text-gray-500">
            <md-icon style={{ fontSize: '16px' }}>vaccines</md-icon>
            Medicine
          </div>
          <span class="font-medium text-gray-600">
            {instanceOfDoctorAppointment(this.appointment) &&
              (this.appointment.medicine?.[0].name ?? '')}
          </span>
        </div>

        {this.isDoctor &&
          instanceOfDoctorAppointment(this.appointment) &&
          ['scheduled', 'requested'].includes(this.appointment.status) &&
          this.editingResources && (
            <div class="mt-3 w-full">
              <h4 class="mb-2 w-full text-center font-medium text-[#7357be]">Edit resources</h4>
              <div class="mb-3 flex w-full flex-col gap-y-3">
                <md-outlined-select
                  label="Facility"
                  class="w-full"
                  value={this.selectedFacility}
                  onInput={(e: Event) => this.handleSelectFacility(e)}
                >
                  <md-select-option value={null}>
                    <div slot="headline">No facility</div>
                  </md-select-option>
                  {this.availableFacilities.map((facility: Facility) => (
                    <md-select-option value={facility.id}>
                      <div slot="headline">{facility.name}</div>
                    </md-select-option>
                  ))}
                </md-outlined-select>

                <md-outlined-select
                  label="Equipment"
                  class="w-full"
                  value={this.selectedEquipment}
                  onInput={(e: Event) => this.handleSelectEquipment(e)}
                >
                  <md-select-option value={null}>
                    <div slot="headline">No equipment</div>
                  </md-select-option>
                  {this.availableEquipment.map((equipment: Equipment) => (
                    <md-select-option value={equipment.id}>
                      <div slot="headline">{equipment.name}</div>
                    </md-select-option>
                  ))}
                </md-outlined-select>

                <md-outlined-select
                  label="Medicine"
                  class="w-full"
                  value={this.selectedMedicine}
                  onInput={(e: Event) => this.handleSelectMedicine(e)}
                >
                  <md-select-option value={null}>
                    <div slot="headline">No medicine</div>
                  </md-select-option>
                  {this.availableMedicine.map((medicine: Medicine) => (
                    <md-select-option value={medicine.id}>
                      <div slot="headline">{medicine.name}</div>
                    </md-select-option>
                  ))}
                </md-outlined-select>
              </div>

              {this.appointment.status === 'scheduled' && (
                <div class="flex flex-row items-center justify-between gap-x-2">
                  <md-filled-button
                    class={`w-1/2 rounded-full bg-[#7357be]`}
                    onClick={this.handleSaveResources}
                  >
                    Save resources
                  </md-filled-button>

                  <md-outlined-button
                    class={`w-1/2 rounded-full`}
                    onClick={() => this.setEditingResources(false)}
                  >
                    Cancel
                  </md-outlined-button>
                </div>
              )}
            </div>
          )}
      </div>
    );
  }
}
