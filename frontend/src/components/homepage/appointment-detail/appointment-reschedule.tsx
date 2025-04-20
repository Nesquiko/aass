import { Doctor, TimeSlot } from '../../../api/generated';
import { renderDateSelects } from '../../../utils/utils';
import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-appointment-reschedule',
  shadow: false,
})
export class AppointmentReschedule {
  @Prop() rescheduling: boolean = false;
  @Prop() reschedulingAppointmentDate: Date = null;
  @Prop() reschedulingAppointmentTime: string = '';
  @Prop() reschedulingAppointmentDoctor: Doctor = null;
  @Prop() reschedulingAppointmentReason: string = '';
  @Prop() reschedulingAvailableDoctors: Array<Doctor> = [];
  @Prop() reschedulingAvailableTimes: Array<TimeSlot> = [];

  @Prop() handleDateSelectInput: (
    type: 'start' | 'end',
    part: 'day' | 'month' | 'year',
    event: Event,
  ) => void;
  @Prop() handleDoctorSelectInput: (event: Event) => void;
  @Prop() handleTimeSelectInput: (event: Event) => void;
  @Prop() handleReschedulingAppointmentReasonChange: (event: Event) => void;

  @Prop() handleReschedule: () => void;

  render() {
    return (
      <div class="my-6 w-full max-w-md rounded-md bg-gray-200 px-4 py-3">
        <h4 class="mb-2 text-sm font-medium text-[#7357be]">Re-schedule appointment</h4>
        <div class="mb-3 flex w-full flex-col gap-y-3">
          {renderDateSelects(
            'start',
            'New appointment date',
            this.reschedulingAppointmentDate,
            this.handleDateSelectInput,
          )}

          <div class="flex w-full flex-col gap-y-1">
            <label class="text-sm font-medium text-gray-600">Doctor</label>
            <md-outlined-select
              label="Doctor"
              class="w-full"
              value={this.reschedulingAppointmentDoctor?.id}
              onInput={this.handleDoctorSelectInput}
            >
              {this.reschedulingAvailableDoctors.map((doctor: Doctor) => (
                <md-select-option value={doctor.id}>
                  <div slot="headline">
                    Dr. {doctor.firstName} {doctor.lastName}
                  </div>
                </md-select-option>
              ))}
            </md-outlined-select>
          </div>

          <div class="flex w-full flex-col gap-y-1">
            <label class="text-sm font-medium text-gray-600">Time</label>
            <md-outlined-select
              label="Time"
              class="w-full"
              value={this.reschedulingAppointmentTime}
              onInput={this.handleTimeSelectInput}
            >
              {this.reschedulingAvailableTimes.map((timeSlot: TimeSlot) => (
                <md-select-option value={timeSlot.time} disabled={timeSlot.status !== 'available'}>
                  <div slot="headline">{timeSlot.time}</div>
                </md-select-option>
              ))}
            </md-outlined-select>
          </div>

          <div class="flex w-full flex-col gap-y-1">
            <label class="text-sm font-medium text-gray-600">Reason for re-scheduling</label>
            <md-outlined-text-field
              label="Rescheduling reason (optional)"
              class="w-full"
              value={this.reschedulingAppointmentReason}
              onInput={this.handleReschedulingAppointmentReasonChange}
            />
          </div>
        </div>

        <div class="flex gap-x-2">
          <md-filled-button
            class={`flex-1 rounded-full bg-[#7357be]`}
            disabled={
              !this.reschedulingAppointmentDate ||
              !this.reschedulingAppointmentTime ||
              !this.reschedulingAppointmentDoctor
            }
            onClick={this.handleReschedule}
          >
            Re-schedule
          </md-filled-button>
          <md-outlined-button
            class="flex-1 rounded-full"
            onClick={() => (this.rescheduling = false)}
          >
            Cancel
          </md-outlined-button>
        </div>
      </div>
    );
  }
}
