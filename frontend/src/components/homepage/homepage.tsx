import { Api } from '../../api/api';
import {
  AppointmentDisplay,
  AppointmentStatus,
  Condition,
  ConditionDisplay,
  Doctor,
  DoctorAppointment,
  DoctorCalendar,
  Equipment,
  Facility,
  Medicine,
  NewPrescription,
  PatientAppointment,
  PatientsCalendar,
  Prescription,
  PrescriptionDisplay,
  UpdatePrescription,
  User,
  UserRole,
} from '../../api/generated';
import { Navigate } from '../../utils/types';
import { TODAY } from '../../utils/utils';
import { StyledHost } from '../StyledHost';
import { toastService } from '../services/toast-service';
import { Component, h, Prop, State } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-home-page',
  shadow: false,
})
export class Homepage {
  private user: User = JSON.parse(sessionStorage.getItem('user'));
  private isDoctor: boolean = this.user.role === 'doctor';

  @Prop() navigate: Navigate;
  @Prop() api: Api;

  @State() appointments: Array<AppointmentDisplay> = [];
  @State() conditions: Array<ConditionDisplay> = [];
  @State() prescriptions: Array<PrescriptionDisplay> = [];

  @State() selectedDate: Date = null;
  @State() selectedAppointment: AppointmentDisplay = null;
  @State() selectedCondition: ConditionDisplay = null;
  @State() selectedPrescription: PrescriptionDisplay = null;
  @State() selectedAppointmentStatusGroup: AppointmentStatus = null;

  @State() isDrawerOpen: boolean = false;
  @State() showLegend: boolean = false;

  @State() hoveredConditionId: string = null;
  @State() hoveredPrescriptionId: string = null;

  @State() currentViewMonth: number = TODAY.getMonth();
  @State() currentViewYear: number = TODAY.getFullYear();

  @State() activeTab: number = 0;

  async componentWillLoad() {
    await this.loadCalendar();
  }

  private async loadCalendar() {
    const fromDate = new Date(this.currentViewYear, this.currentViewMonth, 1);
    const toDate = new Date(this.currentViewYear, this.currentViewMonth + 1, 0);

    try {
      if (this.isDoctor) {
        const calendar: DoctorCalendar = await this.api.doctors.doctorsCalendar({
          doctorId: this.user.id,
          from: fromDate,
          to: toDate,
        });
        this.appointments = calendar.appointments ?? [];
      } else {
        const calendar: PatientsCalendar = await this.api.patients.patientsCalendar({
          patientId: this.user.id,
          from: fromDate,
          to: toDate,
        });
        this.appointments = calendar.appointments ?? [];
        this.conditions = calendar.conditions ?? [];
        this.prescriptions = calendar.prescriptions ?? [];
      }
    } catch (err) {
      toastService.showError(err.message);
    }
  }

  private handlePreviousMonth = () => {
    if (this.currentViewMonth === 0) {
      this.currentViewMonth = 11;
      this.currentViewYear--;
    } else {
      this.currentViewMonth--;
    }
    this.loadCalendar();
  };

  private handleNextMonth = () => {
    if (this.currentViewMonth === 11) {
      this.currentViewMonth = 0;
      this.currentViewYear++;
    } else {
      this.currentViewMonth++;
    }
    this.loadCalendar();
  };

  private handleYearChange = (event: Event) => {
    this.currentViewYear = parseInt((event.target as HTMLSelectElement).value);
    this.loadCalendar();
  };

  private handleToggleLegendMenu = () => {
    this.selectedDate = null;
    this.selectedAppointment = null;
    this.selectedCondition = null;
    this.showLegend = true;
    this.isDrawerOpen = true;
  };

  private handleRescheduleAppointment = async (
    appointment: PatientAppointment | DoctorAppointment,
    newAppointmentDateTime: Date,
    newAppointmentDoctor: Doctor,
    reason: string,
  ) => {
    console.log(
      'Re-schedule appointment:',
      appointment,
      'to new dateTime',
      newAppointmentDateTime,
      'with a doctor',
      newAppointmentDoctor,
      'with a reason',
      reason,
    );

    try {
      await this.api.appointments.rescheduleAppointment({
        appointmentId: appointment.id,
        appointmentReschedule: {
          newAppointmentDateTime: newAppointmentDateTime,
          reason: reason,
        },
      });
    } catch (err) {
      toastService.showError(err.message);
    }
  };

  private handleCancelAppointment = async (
    appointment: PatientAppointment | DoctorAppointment,
    cancellationReason: string,
    by: UserRole,
  ) => {
    try {
      await this.api.appointments.cancelAppointment({
        appointmentId: appointment.id,
        appointmentCancellation: { reason: cancellationReason, by },
      });
    } catch (err) {
      toastService.showError(err.message);
    }
  };

  private handleAcceptAppointment = async (
    appointment: PatientAppointment | DoctorAppointment,
    resources: Partial<{
      facility: Facility;
      equipment: Equipment;
      medicine: Medicine;
    }>,
  ): Promise<DoctorAppointment | undefined> => {
    try {
      return await this.api.appointments.decideAppointment({
        appointmentId: appointment.id,
        appointmentDecision: {
          action: 'accept',
          medicine: resources.medicine ? [resources.medicine] : undefined,
          facilities: resources.facility ? [resources.facility] : undefined,
          equipment: resources.equipment ? [resources.equipment] : undefined,
        },
      });
    } catch (err) {
      toastService.showError(err.message);
    }
  };

  private handleDenyAppointment = async (
    appointment: PatientAppointment | DoctorAppointment,
    denyReason: string,
  ) => {
    try {
      await this.api.appointments.decideAppointment({
        appointmentId: appointment.id,
        appointmentDecision: { action: 'reject', reason: denyReason },
      });
    } catch (err) {
      toastService.showError(err.message);
    }
  };

  private handleSaveResourcesOnAppointment = async (
    appointment: PatientAppointment | DoctorAppointment,
    resources: Partial<{
      facility: Facility;
      equipment: Equipment;
      medicine: Medicine;
    }>,
  ): Promise<DoctorAppointment | undefined> => {
    try {
      return this.api.resources.reserveAppointmentResources({
        appointmentId: appointment.id,
        reserveAppointmentResourcesRequest: {
          start: appointment.appointmentDateTime,
          facilityId: resources.facility?.id,
          equipment: resources.equipment?.id,
          medicine: resources.medicine?.id,
        },
      });
    } catch (err) {
      toastService.showError(err.message);
    }
  };

  private handleUpdatePrescriptionForAppointment = async (
    prescriptionId: string,
    updatedPrescription: UpdatePrescription,
  ): Promise<Prescription | undefined> => {
    try {
      return this.api.medicalHistory.updatePrescription({
        prescriptionId,
        updatePrescription: updatedPrescription,
      });
    } catch (err) {
      toastService.showError(err.message);
      return undefined;
    }
  };

  private handleAddPrescriptionForAppointment = async (
    appointment: DoctorAppointment,
    newPrescription: NewPrescription,
  ): Promise<Prescription | undefined> => {
    try {
      return this.api.medicalHistory.createPrescription({
        newPrescription: {
          name: newPrescription.name,
          start: newPrescription.start,
          end: newPrescription.end,
          appointmentId: appointment.id,
          patientId: appointment.patient.id,
          doctorsNote: newPrescription.doctorsNote,
        },
      });
    } catch (err) {
      toastService.showError(err.message);
      return undefined;
    }
  };

  private handleDeletePrescriptionFromAppointment = async (
    _appointment: DoctorAppointment,
    prescriptionToDelete: PrescriptionDisplay,
  ): Promise<void> => {
    try {
      await this.api.medicalHistory.deletePrescription({
        prescriptionId: prescriptionToDelete.id,
      });
    } catch (err) {
      toastService.showError(err.message);
    }
  };

  private handleResetSelection = () => {
    this.isDrawerOpen = false;
    this.selectedAppointment = null;
    this.selectedCondition = null;
    this.selectedDate = null;
    this.selectedPrescription = null;
    this.showLegend = false;
  };

  private handleSelectDate = (date: Date) => {
    this.handleResetSelection();
    this.selectedDate = date;
    this.selectedAppointmentStatusGroup = 'scheduled';
    this.isDrawerOpen = true;
  };

  private handleSelectAppointment = (appointment: AppointmentDisplay) => {
    this.handleResetSelection();
    this.selectedAppointment = appointment;
    this.isDrawerOpen = true;
  };

  private handleSelectCondition = (condition: ConditionDisplay) => {
    this.handleResetSelection();
    this.selectedCondition = condition;
    this.isDrawerOpen = true;
  };

  private handleSelectPrescription = (prescription: PrescriptionDisplay) => {
    this.handleResetSelection();
    this.selectedPrescription = prescription;
    this.isDrawerOpen = true;
  };

  private handleSelectAppointmentStatusGroup = (
    date: Date,
    appointmentStatus: AppointmentStatus,
  ) => {
    this.handleResetSelection();
    this.selectedDate = date;
    this.selectedAppointmentStatusGroup = appointmentStatus;
    this.isDrawerOpen = true;

    if (this.isDoctor) {
      const statusToTabIndex: Record<AppointmentStatus, number> = {
        scheduled: 0,
        requested: 1,
        denied: 2,
        completed: 3,
        cancelled: 4,
      };
      this.activeTab = statusToTabIndex[appointmentStatus];
    }
  };

  private handleToggleConditionStatus = async (
    condition: Condition,
  ): Promise<Condition | undefined> => {
    // if there is an end on condition, toggle it to null, else toggle it to today
    const end = condition.end ? null : new Date();
    try {
      return this.api.conditions.updateCondition({
        conditionId: condition.id,
        updateCondition: { end },
      });
    } catch (err) {
      toastService.showError(err.message);
      return undefined;
    }
  };

  private getAppointmentsForDate = (date: Date): Array<AppointmentDisplay> => {
    return this.appointments.filter((appointment: AppointmentDisplay): boolean => {
      const appointmentDate: Date = new Date(appointment.appointmentDateTime);
      return (
        appointmentDate.getFullYear() === date.getFullYear() &&
        appointmentDate.getMonth() === date.getMonth() &&
        appointmentDate.getDate() === date.getDate()
      );
    });
  };

  private getConditionsForDate = (date: Date): Array<ConditionDisplay> => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return this.conditions.filter((condition: ConditionDisplay): boolean => {
      const { start, end } = condition;
      const effectiveEndDate: Date = end ? end : today;
      return date >= start && date <= effectiveEndDate;
    });
  };

  private getPrescriptionsForDate = (date: Date): Array<PrescriptionDisplay> => {
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    return this.prescriptions
      .filter((prescription: PrescriptionDisplay): boolean => {
        const { start, end } = prescription;
        return date >= start && date <= end;
      })
      .sort(
        (a: PrescriptionDisplay, b: PrescriptionDisplay) => b.start.getTime() - a.start.getTime(),
      );
  };

  private getAppointmentsForDateByStatus = (date: Date, appointmentStatus: AppointmentStatus) => {
    return this.appointments.filter((appointment: AppointmentDisplay): boolean => {
      if (appointment.status === appointmentStatus) {
        const appointmentDate: Date = new Date(appointment.appointmentDateTime);
        return (
          appointmentDate.getFullYear() === date.getFullYear() &&
          appointmentDate.getMonth() === date.getMonth() &&
          appointmentDate.getDate() === date.getDate()
        );
      }
    });
  };

  render() {
    return (
      <StyledHost class="flex h-screen w-full flex-col overflow-hidden">
        <xcastven-xkilian-project-header
          navigate={this.navigate}
          type="calendar"
          isDoctor={this.isDoctor}
          currentViewMonth={this.currentViewMonth}
          currentViewYear={this.currentViewYear}
          handlePreviousMonth={this.handlePreviousMonth}
          handleNextMonth={this.handleNextMonth}
          handleYearChange={this.handleYearChange}
        />

        <xcastven-xkilian-project-calendar
          user={this.user}
          isDoctor={this.isDoctor}
          appointments={this.appointments}
          conditions={this.conditions}
          prescriptions={this.prescriptions}
          getConditionsForDate={this.getConditionsForDate}
          getPrescriptionsForDate={this.getPrescriptionsForDate}
          handleSelectDate={this.handleSelectDate}
          handleSelectAppointment={this.handleSelectAppointment}
          handleSelectCondition={this.handleSelectCondition}
          handleSelectPrescription={this.handleSelectPrescription}
          handleSelectAppointmentStatusGroup={this.handleSelectAppointmentStatusGroup}
          hoveredConditionId={this.hoveredConditionId}
          setHoveredConditionId={(value: string | null) => (this.hoveredConditionId = value)}
          hoveredPrescriptionId={this.hoveredPrescriptionId}
          setHoveredPrescriptionId={(value: string | null) => (this.hoveredPrescriptionId = value)}
          currentViewMonth={this.currentViewMonth}
          currentViewYear={this.currentViewYear}
          handlePreviousMonth={this.handlePreviousMonth}
          handleNextMonth={this.handleNextMonth}
        />

        <xcastven-xkilian-project-footer
          navigate={this.navigate}
          isDoctor={this.isDoctor}
          handleToggleLegendMenu={this.handleToggleLegendMenu}
        />

        {this.isDrawerOpen && (
          <div class="fixed inset-0 z-99 bg-black/50" onClick={() => this.handleResetSelection()} />
        )}

        <xcastven-xkilian-project-drawer
          navigate={this.navigate}
          api={this.api}
          user={this.user}
          isDoctor={this.isDoctor}
          isDrawerOpen={this.isDrawerOpen}
          selectedDate={this.selectedDate}
          setSelectedDate={(date: Date) => (this.selectedDate = date)}
          selectedAppointment={this.selectedAppointment}
          selectedCondition={this.selectedCondition}
          selectedPrescription={this.selectedPrescription}
          selectedAppointmentStatusGroup={this.selectedAppointmentStatusGroup}
          showLegend={this.showLegend}
          activeTab={this.activeTab}
          handleTabChange={(event: any) => (this.activeTab = event.target.activeTabIndex)}
          handleResetSelection={this.handleResetSelection}
          getAppointmentsForDate={this.getAppointmentsForDate}
          getConditionsForDate={this.getConditionsForDate}
          getPrescriptionsForDate={this.getPrescriptionsForDate}
          getAppointmentsForDateByStatus={this.getAppointmentsForDateByStatus}
          handleSelectAppointment={this.handleSelectAppointment}
          handleSelectCondition={this.handleSelectCondition}
          handleSelectPrescription={this.handleSelectPrescription}
          handleRescheduleAppointment={this.handleRescheduleAppointment}
          handleCancelAppointment={this.handleCancelAppointment}
          handleAcceptAppointment={this.handleAcceptAppointment}
          handleDenyAppointment={this.handleDenyAppointment}
          handleSaveResourcesOnAppointment={this.handleSaveResourcesOnAppointment}
          handleUpdatePrescriptionForAppointment={this.handleUpdatePrescriptionForAppointment}
          handleAddPrescriptionForAppointment={this.handleAddPrescriptionForAppointment}
          handleDeletePrescriptionFromAppointment={this.handleDeletePrescriptionFromAppointment}
          handleToggleConditionStatus={this.handleToggleConditionStatus}
        />
      </StyledHost>
    );
  }
}
