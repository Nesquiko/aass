import {
  Appointment,
  AppointmentStatus,
  Equipment,
  Facility,
  Medicine,
} from '../../../api/generated';
import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-appointment-actions',
  shadow: false,
})
export class AppointmentActions {
  @Prop() isDoctor: boolean;
  @Prop() appointment: Appointment;
  @Prop() handleCancelDoctorAppointment: (appointment: Appointment) => void;
  @Prop() handleAcceptAppointment: (appointment: Appointment) => void;
  @Prop() handleDenyAppointment: (appointment: Appointment) => void;
  @Prop() handleRescheduleAppointment: () => void;
  @Prop() handleCancelPatientAppointment: (appointment: Appointment) => void;

  private patientButton = (
    displayTitle: string,
    widthClass: string,
    onClick: (appointment: Appointment) => void,
  ) => {
    return (
      <md-filled-button class={`${widthClass} rounded-full bg-[#7357be]`} onClick={onClick}>
        {displayTitle}
      </md-filled-button>
    );
  };

  private getPatientAppointmentActions = (
    appointmentStatus: AppointmentStatus,
    handleRescheduleAppointment: () => void,
    handleCancelAppointment: (appointment: Appointment) => void,
  ) => {
    switch (appointmentStatus) {
      case 'scheduled':
        return (
          <div class="flex w-full max-w-md flex-row items-center justify-between gap-x-3">
            {this.patientButton('Re-schedule', 'w-1/2', handleRescheduleAppointment)}
            {this.patientButton('Cancel', 'w-1/2', handleCancelAppointment)}
          </div>
        );
      case 'requested':
        return (
          <div class="flex w-full max-w-md flex-row items-center justify-between gap-x-3">
            {this.patientButton('Cancel request', 'w-full', handleCancelAppointment)}
          </div>
        );
      default:
        return null;
    }
  };

  private doctorButton = (
    displayTitle: string,
    widthClass: string,
    onClick: (
      appointment: Appointment,
      resources?: {
        facilities: Array<Facility>;
        equipment: Array<Equipment>;
        medicine: Array<Medicine>;
      },
    ) => void,
  ) => {
    return (
      <md-filled-button class={`${widthClass} rounded-full bg-[#7357be]`} onClick={onClick}>
        {displayTitle}
      </md-filled-button>
    );
  };

  private getDoctorAppointmentActions = (
    appointmentStatus: AppointmentStatus,
    handleCancelAppointment: (appointment: Appointment) => void,
    handleAcceptAppointment: (appointment: Appointment) => void,
    handleDenyAppointment: (appointment: Appointment) => void,
  ) => {
    switch (appointmentStatus) {
      case 'scheduled':
        return (
          <div class="flex w-full max-w-md flex-row items-center justify-between gap-x-3">
            {this.doctorButton('Cancel', 'w-full', handleCancelAppointment)}
          </div>
        );
      case 'requested':
        return (
          <div class="flex w-full max-w-md flex-row items-center justify-between gap-x-3">
            {this.doctorButton('Accept', 'w-1/2', handleAcceptAppointment)}
            {this.doctorButton('Deny', 'w-1/2', handleDenyAppointment)}
          </div>
        );
      default:
        return null;
    }
  };

  render() {
    if (this.isDoctor) {
      return this.getDoctorAppointmentActions(
        this.appointment.status,
        this.handleCancelDoctorAppointment,
        this.handleAcceptAppointment,
        this.handleDenyAppointment,
      );
    } else {
      return this.getPatientAppointmentActions(
        this.appointment.status,
        this.handleRescheduleAppointment,
        this.handleCancelPatientAppointment,
      );
    }
  }
}
