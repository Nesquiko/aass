import { Api } from '../../api/api';
import {
  AppointmentType,
  ConditionDisplay,
  Doctor,
  NewAppointmentRequest,
  TimeSlot,
  User,
} from '../../api/generated';
import { Navigate } from '../../utils/types';
import {
  formatAppointmentType,
  formatDate,
  formatSpecialization,
  getSelectedDateTimeObject,
  TODAY,
} from '../../utils/utils';
import { toastService } from '../services/toast-service';
import { Component, h, Prop, State } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-appointment-scheduler',
  shadow: false,
})
export class AppointmentScheduler {
  @Prop() navigate: Navigate;
  @Prop() api: Api;
  @Prop() user: User;
  @Prop() initialDate: Date = null;
  @Prop() conditionId: string = null;

  @State() selectedDate: Date = null;
  @State() selectedDoctor: Doctor = null;
  @State() selectedAppointmentType: AppointmentType = null;
  @State() selectedTime: string = null;
  @State() selectedCondition: ConditionDisplay = null;
  @State() appointmentReason: string = '';
  @State() currentViewMonth: number = TODAY.getMonth();
  @State() currentViewYear: number = TODAY.getFullYear();

  @State() private availableTimes: Array<TimeSlot> = [];
  @State() private availableDoctors: Array<Doctor> = [];
  @State() private activeConditions: Array<ConditionDisplay> = [];

  async componentWillLoad() {
    if (this.initialDate) {
      this.selectedDate = this.initialDate;
      this.currentViewMonth = this.initialDate.getMonth();
      this.currentViewYear = this.initialDate.getFullYear();
    }

    try {
      const docs = await this.api.doctors.getDoctors();
      this.availableDoctors = docs.doctors ?? [];
    } catch (err) {
      toastService.showError(err.message);
    }
  }

  private async loadAvailableTimes(doc: Doctor, d: Date) {
    try {
      const slots = await this.api.doctors.doctorsTimeslots({
        doctorId: doc.id,
        date: d,
      });
      this.availableTimes = slots.slots;
    } catch (err) {
      toastService.showError(err.message);
    }
  }

  private async loadActiveConditions(date: Date, patientId: string) {
    try {
      const conds = await this.api.conditions.conditionsInDateRange({
        from: date,
        to: date,
        patientId,
      });
      this.activeConditions = conds.conditions;
    } catch (err) {
      toastService.showError(err.message);
    }
    if (this.conditionId) {
      this.selectedCondition = this.activeConditions.find(
        (condition: ConditionDisplay) => condition.id === this.conditionId,
      );
    }
  }

  private selectDate = (day: number) => {
    console.log(day, this.currentViewYear, this.currentViewMonth);
    this.selectedDate = new Date(this.currentViewYear, this.currentViewMonth, day);

    this.loadActiveConditions(this.selectedDate, this.user.id);
    if (this.selectedDoctor) {
      this.loadAvailableTimes(this.selectedDoctor, this.selectedDate);
    }
  };

  private handleTimeChange = (event: Event) => {
    this.selectedTime = (event.target as HTMLSelectElement).value;
  };

  private handleAppointmentTypeChange = (event: Event) => {
    this.selectedAppointmentType = (event.target as HTMLSelectElement).value as AppointmentType;
  };

  private handleDoctorChange = (event: Event) => {
    const doctorId: string = (event.target as HTMLSelectElement).value;
    this.selectedDoctor = this.availableDoctors.find((doctor: Doctor) => doctor.id === doctorId);

    if (this.selectedDate) {
      this.loadAvailableTimes(this.selectedDoctor, this.selectedDate);
    }
  };

  private handleConditionChange = (event: Event) => {
    const conditionId: string = (event.target as HTMLSelectElement).value;
    this.selectedCondition = this.activeConditions.find(
      (condition: ConditionDisplay) => condition.id === conditionId,
    );
  };

  private handleAppointmentReasonChange = (event: Event) => {
    this.appointmentReason = (event.target as HTMLTextAreaElement).value;
  };

  private handleScheduleAppointment = async () => {
    const appointmentDateTime: Date = getSelectedDateTimeObject(
      this.selectedDate,
      this.selectedTime,
    );

    const newAppointment: NewAppointmentRequest = {
      patientId: this.user.id,
      appointmentDateTime: appointmentDateTime,
      type: this.selectedAppointmentType,
      conditionId: this.selectedCondition?.id,
      reason: this.appointmentReason,
      doctorId: this.selectedDoctor.id,
    };

    try {
      await this.api.appointments.requestAppointment({ newAppointmentRequest: newAppointment });
      this.navigate('./homepage');
    } catch (err) {
      toastService.showError(err.message);
    }
  };

  private showDetailsPanel = () => {
    return this.selectedDate !== null && this.selectedDoctor !== null;
  };

  render() {
    const showDetails: boolean = this.showDetailsPanel();

    return (
      <div class="flex h-screen w-full flex-1 flex-col overflow-auto">
        {/* Header */}
        <xcastven-xkilian-project-header
          navigate={this.navigate}
          type="scheduleAppointment"
          isDoctor={false}
        />

        {/* Content */}
        <div class="mx-auto flex w-full flex-1 flex-col md:flex-row">
          {/* Left panel - Date & Doctor */}
          <div
            class={`flex flex-col items-center justify-center bg-gray-300 p-6 transition-all duration-600 ease-in-out ${showDetails ? 'w-full md:w-1/2' : 'w-full'}`}
          >
            <xcastven-xkilian-project-date-picker
              selectedDate={this.selectedDate}
              selectDate={this.selectDate}
              currentViewMonth={this.currentViewMonth}
              setCurrentViewMonth={(newMonth: number) => (this.currentViewMonth = newMonth)}
              currentViewYear={this.currentViewYear}
              setCurrentViewYear={(newYear: number) => (this.currentViewYear = newYear)}
            />

            <div class="w-full max-w-md px-4">
              <md-filled-select
                label="Doctor / Specialist"
                class="w-full"
                value={this.selectedDoctor}
                onInput={(e: Event) => this.handleDoctorChange(e)}
              >
                {this.availableDoctors.map((doctor: Doctor) => (
                  <md-select-option value={doctor.id}>
                    <div slot="headline">
                      Dr. {doctor.firstName} {doctor.lastName} -{' '}
                      {formatSpecialization(doctor.specialization)}
                    </div>
                  </md-select-option>
                ))}
              </md-filled-select>
            </div>
          </div>

          {/* Right panel - Details */}
          {showDetails && (
            <div
              class={`m-auto flex w-full max-w-lg transform animate-[slideInFromBottom_0.5s_ease-out] flex-col justify-center p-6 opacity-100 transition-all duration-500 ease-in-out md:w-1/2 md:animate-[slideInFromRight_0.5s_ease-out]`}
            >
              <div class="mb-6 font-medium text-[#7357be]">{formatDate(this.selectedDate)}</div>

              <div class="mb-6">
                <md-filled-select
                  label="Appointment type"
                  class="w-full"
                  value={this.selectedAppointmentType}
                  onInput={(e: Event) => this.handleAppointmentTypeChange(e)}
                >
                  {Object.values(AppointmentType).map((appointmentType: AppointmentType) => (
                    <md-select-option value={appointmentType}>
                      <div slot="headline">{formatAppointmentType(appointmentType)}</div>
                    </md-select-option>
                  ))}
                </md-filled-select>
              </div>

              <div class="mb-6">
                <md-filled-select
                  label="Select a time"
                  class="w-full"
                  value={this.selectedTime}
                  onInput={(e: Event) => this.handleTimeChange(e)}
                >
                  {this.availableTimes.map((time: TimeSlot) => (
                    <md-select-option value={time.time} disabled={time.status !== 'available'}>
                      <div slot="headline">{time.time}</div>
                    </md-select-option>
                  ))}
                </md-filled-select>
              </div>

              <div class="mb-6">
                <md-filled-select
                  label="Assign this appointment to a condition"
                  class="w-full"
                  value={this.selectedCondition}
                  onInput={(e: Event) => this.handleConditionChange(e)}
                >
                  {this.activeConditions.map((condition: ConditionDisplay) => (
                    <md-select-option value={condition.id}>
                      <div slot="headline">
                        {condition.name} - since {formatDate(condition.start)}
                      </div>
                    </md-select-option>
                  ))}
                </md-filled-select>
              </div>

              <div class="mb-6">
                <md-filled-text-field
                  class="w-full"
                  label="Appointment reason"
                  placeholder="Briefly describe your reason for visit (optional)"
                  type="textarea"
                  rows="4"
                  value={this.appointmentReason}
                  onInput={(e: Event) => this.handleAppointmentReasonChange(e)}
                ></md-filled-text-field>
              </div>

              <md-filled-button
                class="w-full rounded-full bg-[#7357be] text-white"
                onClick={() => this.handleScheduleAppointment()}
              >
                Schedule
              </md-filled-button>
            </div>
          )}
        </div>
      </div>
    );
  }
}
