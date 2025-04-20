import {
  AppointmentStatus,
  DoctorAppointment,
  instanceOfDoctorAppointment,
  PatientAppointment,
  PrescriptionDisplay,
} from '../../../api/generated';
import { renderDateSelects } from '../../../utils/utils';
import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-appointment-prescriptions',
  shadow: false,
})
export class AppointmentPrescriptions {
  @Prop() isDoctor: boolean;
  @Prop() appointment: PatientAppointment | DoctorAppointment;
  @Prop() prescriptionsExpanded: boolean;
  @Prop() handleSelectPrescription: (prescription: PrescriptionDisplay) => void;

  @Prop() addingPrescription: boolean;
  @Prop() setAddingPrescription: (addingPrescription: boolean) => void;
  @Prop() addingPrescriptionName: string;
  @Prop() addingPrescriptionStart: Date;
  @Prop() addingPrescriptionEnd: Date;
  @Prop() addingPrescriptionDoctorsNote: string;
  @Prop() handleAddPrescriptionNameChange: (event: Event) => void;
  @Prop() handleAddPrescriptionDateChange: (
    type: 'start' | 'end',
    part: 'day' | 'month' | 'year',
    event: Event,
  ) => void;
  @Prop() handleAddPrescriptionDoctorsNoteChange: (event: Event) => void;
  @Prop() handleAddPrescription: () => void;

  @Prop() editingPrescription: PrescriptionDisplay;
  @Prop() setEditingPrescription: (editingPrescription: PrescriptionDisplay) => void;
  @Prop() editingPrescriptionNewName: string;
  @Prop() editingPrescriptionNewStart: Date;
  @Prop() editingPrescriptionNewEnd: Date;
  @Prop() editingPrescriptionNewDoctorsNote: string;
  @Prop() handleUpdatePrescriptionNameChange: (event: Event) => void;
  @Prop() handleUpdatePrescriptionDateChange: (
    type: 'start' | 'end',
    part: 'day' | 'month' | 'year',
    event: Event,
  ) => void;
  @Prop() handleUpdatePrescriptionDoctorsNoteChange: (event: Event) => void;
  @Prop() handleUpdatePrescription: () => void;

  @Prop() deletingPrescription: PrescriptionDisplay;
  @Prop() setDeletingPrescription: (deletingPrescription: PrescriptionDisplay) => void;
  @Prop() handleDeletePrescription: () => void;

  private toggleAddPrescriptionForm = (event: Event) => {
    event.stopPropagation();

    if (this.addingPrescription) {
      this.setAddingPrescription(false);
      return;
    }

    this.setAddingPrescription(!this.addingPrescription);
    this.setEditingPrescription(null);
    this.setDeletingPrescription(null);
  };

  private toggleEditPrescriptionForm = (event: Event, prescription: PrescriptionDisplay) => {
    event.stopPropagation();

    if (this.editingPrescription && this.editingPrescription.id === prescription.id) {
      this.setEditingPrescription(null);
      return;
    }

    this.editingPrescriptionNewName = prescription.name;
    this.editingPrescriptionNewStart = prescription.start ? new Date(prescription.start) : null;
    this.editingPrescriptionNewEnd = prescription.end ? new Date(prescription.end) : null;
    // this.editingPrescriptionNewDoctorsNote = prescription.doctorsNote ?? '';
    this.setEditingPrescription(prescription);
    this.setAddingPrescription(false);
    this.setDeletingPrescription(null);
  };

  private toggleDeletePrescriptionForm = (event: Event, prescription: PrescriptionDisplay) => {
    event.stopPropagation();

    if (this.deletingPrescription && this.deletingPrescription.id === prescription.id) {
      this.setDeletingPrescription(null);
      return;
    }

    this.setDeletingPrescription(prescription);
    this.setAddingPrescription(false);
    this.setEditingPrescription(null);
  };

  render() {
    return (
      <div class="relative mb-6 w-full max-w-md rounded-md bg-gray-200 px-4 py-3">
        <div class="mb-2 flex flex-row items-center justify-between">
          <div class="flex flex-row items-center gap-x-2 text-gray-500">
            <md-icon style={{ fontSize: '16px' }}>medication</md-icon>
            Prescriptions
            {instanceOfDoctorAppointment(this.appointment) &&
              (this.appointment.status === AppointmentStatus.Completed ||
                this.appointment.status === AppointmentStatus.Scheduled) && (
                <md-icon-button
                  title="Add a prescription"
                  class="ml-2"
                  style={{ width: '24px', height: '24px' }}
                  onClick={(event: Event) => this.toggleAddPrescriptionForm(event)}
                >
                  <md-icon style={{ fontSize: '16px' }}>add</md-icon>
                </md-icon-button>
              )}
          </div>
          {this.appointment.prescriptions && this.appointment.prescriptions.length > 0 && (
            <md-icon-button
              title={this.prescriptionsExpanded ? 'Collapse' : 'Expand'}
              onClick={() => (this.prescriptionsExpanded = !this.prescriptionsExpanded)}
              style={{ width: '24px', height: '24px' }}
            >
              <md-icon style={{ fontSize: '16px' }}>
                {this.prescriptionsExpanded ? 'expand_less' : 'expand_more'}
              </md-icon>
            </md-icon-button>
          )}
        </div>

        {/* Prescription List */}
        {(!this.appointment.prescriptions || this.appointment.prescriptions.length <= 0) &&
        !this.addingPrescription &&
        !this.editingPrescription ? (
          <div class="text-sm font-medium text-gray-600">
            No prescriptions for this appointment.
          </div>
        ) : this.prescriptionsExpanded ? (
          <div class="max-h-28 w-full overflow-y-auto rounded-md bg-gray-200">
            {this.appointment.prescriptions.map((prescription: PrescriptionDisplay) => (
              <div
                key={prescription.id}
                class="mr-2 flex cursor-pointer items-center justify-between rounded px-2 py-1 hover:bg-gray-300"
                onClick={() => this.handleSelectPrescription(prescription)}
              >
                <div class="flex items-center font-medium text-gray-600">
                  <md-icon class="mr-2" style={{ fontSize: '14px' }}>
                    medication
                  </md-icon>
                  {prescription.name}
                </div>
                {instanceOfDoctorAppointment(this.appointment) &&
                  ['completed', 'scheduled'].includes(this.appointment.status) && (
                    <div class="flex flex-row items-center gap-x-2">
                      {/* Edit prescription */}
                      <md-icon-button
                        title="Edit prescription"
                        onClick={(event: Event) =>
                          this.toggleEditPrescriptionForm(event, prescription)
                        }
                        style={{ width: '24px', height: '24px' }}
                      >
                        <md-icon style={{ fontSize: '16px' }}>edit</md-icon>
                      </md-icon-button>

                      {/* Delete prescription */}
                      <md-icon-button
                        title="Delete prescription"
                        onClick={(event: Event) =>
                          this.toggleDeletePrescriptionForm(event, prescription)
                        }
                        style={{ width: '24px', height: '24px' }}
                      >
                        <md-icon style={{ fontSize: '16px' }}>delete</md-icon>
                      </md-icon-button>
                    </div>
                  )}
              </div>
            ))}
          </div>
        ) : (
          !this.addingPrescription &&
          !this.editingPrescription && (
            <div class="ml-2 text-sm font-medium text-gray-600">
              {this.appointment.prescriptions?.length ?? 0} prescription
              {this.appointment.prescriptions?.length !== 1 ? 's' : ''}
            </div>
          )
        )}

        {/* Add Prescription */}
        {this.isDoctor &&
          instanceOfDoctorAppointment(this.appointment) &&
          this.addingPrescription && (
            <div class="mt-3 w-full pt-3">
              <h4 class="mb-2 w-full text-center font-medium text-[#7357be]">
                Add a new prescription
              </h4>
              <div class="mb-3 flex w-full flex-col gap-y-3">
                <div class="flex w-full flex-col gap-y-1">
                  <label class="text-sm font-medium text-gray-600">Prescription name</label>
                  <md-outlined-text-field
                    required={true}
                    label="Prescription name"
                    class="w-full"
                    value={this.addingPrescriptionName}
                    onInput={(e: Event) => this.handleAddPrescriptionNameChange(e)}
                  />
                </div>

                {renderDateSelects(
                  'start',
                  'Prescription start date',
                  this.addingPrescriptionStart,
                  this.handleAddPrescriptionDateChange,
                )}

                {renderDateSelects(
                  'end',
                  'Prescription end date',
                  this.addingPrescriptionEnd,
                  this.handleAddPrescriptionDateChange,
                )}

                <div class="flex w-full flex-col gap-y-1">
                  <label class="text-sm font-medium text-gray-600">Doctor's note</label>
                  <md-outlined-text-field
                    type="textarea"
                    rows={2}
                    label="Doctor's note (optional)"
                    class="w-full"
                    value={this.addingPrescriptionDoctorsNote}
                    onInput={(e: Event) => this.handleAddPrescriptionDoctorsNoteChange(e)}
                  />
                </div>
              </div>

              <div class="flex gap-x-2">
                <md-filled-button
                  class={`flex-1 rounded-full bg-[#7357be]`}
                  disabled={
                    !this.addingPrescriptionName ||
                    !this.addingPrescriptionStart ||
                    !this.addingPrescriptionEnd
                  }
                  onClick={this.handleAddPrescription}
                >
                  Add Prescription
                </md-filled-button>
                <md-outlined-button
                  class="flex-1 rounded-full"
                  onClick={() => this.setAddingPrescription(false)}
                >
                  Cancel
                </md-outlined-button>
              </div>
            </div>
          )}

        {/* Edit Prescription */}
        {this.isDoctor &&
          instanceOfDoctorAppointment(this.appointment) &&
          this.editingPrescription && (
            <div class="mt-3 w-full pt-3">
              <h4 class="mb-2 w-full text-center font-medium text-[#7357be]">Edit prescription</h4>
              <div class="mb-3 flex w-full flex-col gap-y-3">
                <div class="flex w-full flex-col gap-y-1">
                  <label class="text-sm font-medium text-gray-600">Prescription name</label>
                  <md-outlined-text-field
                    required={true}
                    label="Prescription name"
                    class="w-full"
                    value={this.editingPrescriptionNewName}
                    onInput={(e: Event) => this.handleUpdatePrescriptionNameChange(e)}
                  />
                </div>

                {renderDateSelects(
                  'start',
                  'Prescription start date',
                  this.editingPrescriptionNewStart,
                  this.handleUpdatePrescriptionDateChange,
                )}

                {renderDateSelects(
                  'end',
                  'Prescription end date',
                  this.editingPrescriptionNewEnd,
                  this.handleUpdatePrescriptionDateChange,
                )}
                <div class="flex w-full flex-col gap-y-1">
                  <label class="text-sm font-medium text-gray-600">
                    Prescription doctor's note
                  </label>
                  <md-outlined-text-field
                    type="textarea"
                    rows={2}
                    label="Doctor's note (optional)"
                    class="w-full"
                    value={this.editingPrescriptionNewDoctorsNote}
                    onInput={(e: Event) => this.handleUpdatePrescriptionDoctorsNoteChange(e)}
                  />
                </div>
              </div>

              <div class="flex gap-x-2">
                <md-filled-button
                  class={`flex-1 rounded-full bg-[#7357be]`}
                  disabled={
                    !this.editingPrescriptionNewName ||
                    !this.editingPrescriptionNewStart ||
                    !this.editingPrescriptionNewEnd
                  }
                  onClick={this.handleUpdatePrescription}
                >
                  Save Changes
                </md-filled-button>
                <md-outlined-button
                  class="flex-1 rounded-full"
                  onClick={() => this.setEditingPrescription(null)}
                >
                  Cancel
                </md-outlined-button>
              </div>
            </div>
          )}

        {/* Delete Prescription */}
        {this.isDoctor &&
          instanceOfDoctorAppointment(this.appointment) &&
          this.deletingPrescription && (
            <div class="mt-3 w-full pt-3">
              <h4 class="mb-2 w-full text-center font-medium text-[#7357be]">
                Delete prescription
              </h4>
              <p class="mb-3 w-full text-center text-sm text-gray-600">
                Are you sure you want to delete this prescription?
              </p>

              <div class="flex gap-x-2">
                <md-filled-button
                  class={`flex-1 rounded-full bg-[#7357be]`}
                  onClick={this.handleDeletePrescription}
                >
                  Confirm delete
                </md-filled-button>
                <md-outlined-button
                  class="flex-1 rounded-full"
                  onClick={() => this.setDeletingPrescription(null)}
                >
                  Cancel
                </md-outlined-button>
              </div>
            </div>
          )}
      </div>
    );
  }
}
