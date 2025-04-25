import { Api } from '../../../api/api';
import {
  Appointment,
  AppointmentStatus,
  AvailableResources,
  Doctor,
  Equipment,
  Facility,
  Medicine,
  NewPrescription,
  Prescription,
  PrescriptionDisplay,
  TimeSlot,
  UpdatePrescription,
  User,
  UserRole,
} from '../../../api/generated';
import {
  formatAppointmentType,
  formatDate,
  formatTime,
  updateDatePart,
} from '../../../utils/utils';
import { toastService } from '../../services/toast-service';
import { Component, h, Prop, State } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-appointment-detail',
  shadow: false,
})
export class AppointmentDetail {
  @Prop() api: Api;
  @Prop() user: User;
  @Prop() isDoctor: boolean;
  @Prop() appointmentId: string;
  @Prop() handleResetSelection: () => void;

  @Prop() handleRescheduleAppointment: (
    appointment: Appointment,
    newAppointmentDateTime: Date,
    newAppointmentDoctor: Doctor,
    reason: string,
  ) => Promise<void>;
  @Prop() handleCancelAppointment: (
    appointment: Appointment,
    cancellationReason: string,
    by: UserRole,
  ) => Promise<void>;

  @Prop() handleAcceptAppointment: (
    appointment: Appointment,
    resources: Partial<{
      facility: Facility;
      equipment: Equipment;
      medicine: Medicine;
    }>,
  ) => Promise<Appointment | undefined>;
  @Prop() handleDenyAppointment: (appointment: Appointment, denyReason: string) => Promise<void>;
  @Prop() handleSaveResourcesOnAppointment: (
    appointment: Appointment,
    resources: Partial<{
      facility: Facility;
      equipment: Equipment;
      medicine: Medicine;
    }>,
  ) => Promise<void>;
  @Prop() handleSelectPrescription: (prescription: PrescriptionDisplay) => void;
  @Prop() handleUpdatePrescriptionForAppointment: (
    prescriptionId: string,
    updatedPrescription: UpdatePrescription,
  ) => Promise<Prescription | undefined>;
  @Prop() handleAddPrescriptionForAppointment: (
    appointment: Appointment,
    newPrescription: NewPrescription,
  ) => Promise<Prescription | undefined>;
  @Prop() handleDeletePrescriptionFromAppointment: (
    prescriptionToDelete: PrescriptionDisplay,
  ) => Promise<void>;

  @State() appointment: Appointment = undefined;
  @State() availableEquipment: Array<Equipment> = [];
  @State() availableFacilities: Array<Facility> = [];
  @State() availableMedicine: Array<Medicine> = [];

  @State() selectedEquipment: Equipment | undefined = this.appointment?.equipment
    ? this.appointment.equipment?.[0]
    : undefined;
  @State() selectedFacility: Facility | undefined = this.appointment?.facilities
    ? this.appointment.facilities?.[0]
    : undefined;
  @State() selectedMedicine: Medicine | undefined = this.appointment?.medicine
    ? this.appointment.medicine?.[0]
    : undefined;

  @State() prescriptionsExpanded: boolean = false;

  @State() editingResources: boolean = false;

  @State() addingPrescription: boolean = false;
  @State() addingPrescriptionName: string = '';
  @State() addingPrescriptionStart: Date = new Date();
  @State() addingPrescriptionEnd: Date = new Date(new Date().setDate(new Date().getDate() + 7));
  @State() addingPrescriptionDoctorsNote: string = '';

  @State() editingPrescription: PrescriptionDisplay = null;
  @State() editingPrescriptionNewName: string = '';
  @State() editingPrescriptionNewStart: Date = null;
  @State() editingPrescriptionNewEnd: Date = null;
  @State() editingPrescriptionNewDoctorsNote: string = '';

  @State() deletingPrescription: PrescriptionDisplay = null;

  @State() rescheduling: boolean = false;
  @State() reschedulingAppointmentDate: Date = null;
  @State() reschedulingAppointmentTime: string = '';
  @State() reschedulingAppointmentDoctor: Doctor = null;
  @State() reschedulingAppointmentReason: string = '';
  @State() reschedulingAvailableDoctors: Array<Doctor> = [];
  @State() reschedulingAvailableTimes: Array<TimeSlot> = [];

  @State() cancelling: boolean = false;
  @State() cancellingAppointmentReason: string = '';

  @State() denying: boolean = false;
  @State() denyingAppointmentReason: string = '';

  async componentWillLoad() {
    await this.loadAppointment();

    if (this.isDoctor) {
      await this.loadAvailableResources();
    }
  }

  private async loadAppointment() {
    try {
      const appointment = await this.api.appointments.appointmentById({
        appointmentId: this.appointmentId,
      });
      this.appointment = appointment;
      this.selectedEquipment = appointment.equipment?.[0] ?? undefined;
      this.selectedFacility = appointment.facilities?.[0] ?? undefined;
      this.selectedMedicine = appointment.medicine?.[0] ?? undefined;
    } catch (err) {
      toastService.showError(err.message);
    }
  }

  private async loadAvailableResources() {
    try {
      const resources: AvailableResources = await this.api.resources.getAvailableResources({
        dateTime: this.appointment.appointmentDateTime,
      });
      this.availableMedicine = resources.medicine;
      this.availableFacilities = resources.facilities;
      this.availableEquipment = resources.equipment;
    } catch (err) {
      toastService.showError(err.message);
    }
  }

  private async loadReschedulingDoctors() {
    try {
      const doctors = await this.api.doctors.getDoctors();
      this.reschedulingAvailableDoctors = doctors.doctors ?? [];
    } catch (err) {
      toastService.showError(err.message);
    }
  }

  private async loadReschedulingTimes(doc: Doctor, d: Date) {
    try {
      const slots = await this.api.doctors.doctorsTimeslots({
        doctorId: doc.id,
        date: d,
      });
      this.reschedulingAvailableTimes = slots.slots;
    } catch (err) {
      toastService.showError(err.message);
    }
  }

  private getPatientAppointmentStatusMessage = (appointmentStatus: AppointmentStatus) => {
    switch (appointmentStatus) {
      case 'requested':
        return "This appointment is waiting for a reaction from the Doctor's office.";
      case 'scheduled':
        return '';
      case 'completed':
        return 'This appointment has already been completed.';
      case 'denied':
        return "This appointment has already been denied by the Doctor's office.";
      case 'cancelled':
        return 'This appointment has already been cancelled.';
      default:
        return '';
    }
  };

  private getDoctorAppointmentStatusMessage = (appointmentStatus: AppointmentStatus) => {
    switch (appointmentStatus) {
      case 'requested':
        return '';
      case 'scheduled':
        return '';
      case 'completed':
        return 'This appointment has already been completed.';
      case 'denied':
        return 'This appointment has already been denied.';
      case 'cancelled':
        return 'This appointment has already been cancelled.';
      default:
        return '';
    }
  };

  private handleSelectFacility = (event: Event) => {
    const newFacilityId: string = (event.target as HTMLSelectElement).value;
    this.selectedFacility = this.availableFacilities.find(
      (facility: Facility): boolean => facility.id === newFacilityId,
    );
  };

  private handleSelectEquipment = (event: Event) => {
    const newEquipmentId: string = (event.target as HTMLSelectElement).value;
    this.selectedEquipment = this.availableEquipment.find(
      (equipment: Equipment): boolean => equipment.id === newEquipmentId,
    );
  };

  private handleSelectMedicine = (event: Event) => {
    const newMedicineId: string = (event.target as HTMLSelectElement).value;
    this.selectedMedicine = this.availableMedicine.find(
      (medicine: Medicine): boolean => medicine.id === newMedicineId,
    );
  };

  private handleDateSelectInput = (
    type: 'start' | 'end',
    part: 'day' | 'month' | 'year',
    event: Event,
  ) => {
    const value: number = parseInt((event.target as HTMLSelectElement).value, 10);
    this.reschedulingAppointmentDate = updateDatePart(
      this.reschedulingAppointmentDate,
      part,
      value,
    );
  };

  private handleDoctorSelectInput = (e: Event) => {
    const newDoctorId: string = (e.target as HTMLSelectElement).value;
    this.reschedulingAppointmentDoctor = this.reschedulingAvailableDoctors.find(
      (doctor: Doctor) => doctor.id === newDoctorId,
    );

    if (this.reschedulingAppointmentDoctor && this.reschedulingAppointmentDate) {
      this.loadReschedulingTimes(
        this.reschedulingAppointmentDoctor,
        this.reschedulingAppointmentDate,
      );
    }
  };

  private handleTimeSelectInput = (e: Event) => {
    this.reschedulingAppointmentTime = (e.target as HTMLSelectElement).value;

    if (this.reschedulingAppointmentDoctor && this.reschedulingAppointmentDate) {
      this.loadReschedulingTimes(
        this.reschedulingAppointmentDoctor,
        this.reschedulingAppointmentDate,
      );
    }
  };

  private handleReschedulingAppointmentReasonChange = (e: Event) => {
    this.reschedulingAppointmentReason = (e.target as HTMLInputElement).value;
  };

  private handleReschedule = () => {
    if (
      this.reschedulingAppointmentDate &&
      this.reschedulingAppointmentTime &&
      this.reschedulingAppointmentDoctor
    ) {
      const newDateTime = new Date(this.reschedulingAppointmentDate);
      const [hours, minutes] = this.reschedulingAppointmentTime.split(':').map(Number);
      newDateTime.setHours(hours);
      newDateTime.setMinutes(minutes);

      this.handleRescheduleAppointment(
        this.appointment,
        newDateTime,
        this.reschedulingAppointmentDoctor,
        this.reschedulingAppointmentReason,
      ).then(() => {
        this.appointment = {
          ...this.appointment,
          status: 'requested',
          facilities: [],
          equipment: [],
          medicine: [],
          appointmentDateTime: newDateTime,
          doctor: this.reschedulingAppointmentDoctor,
        };
        this.rescheduling = false;
      });
    }
  };

  private handleCancellingAppointmentReasonChange = (e: Event) => {
    this.cancellingAppointmentReason = (e.target as HTMLInputElement).value;
  };

  private handleCancel = () => {
    const by = this.isDoctor ? UserRole.Doctor : UserRole.Patient;
    this.handleCancelAppointment(this.appointment, this.cancellingAppointmentReason, by).then(
      () => {
        this.appointment = {
          ...this.appointment,
          status: 'cancelled',
        };
        this.cancelling = false;
      },
    );
  };

  private handleAccept = () => {
    this.handleAcceptAppointment(this.appointment, {
      facility: this.selectedFacility,
      equipment: this.selectedEquipment,
      medicine: this.selectedMedicine,
    }).then((updatedAppointment: Appointment | undefined) => {
      if (!updatedAppointment) return;
      this.appointment = updatedAppointment;
    });
  };

  private handleDeny = () => {
    this.handleDenyAppointment(this.appointment, this.denyingAppointmentReason).then(() => {
      this.appointment = {
        ...this.appointment,
        status: 'denied',
      };
      this.denying = false;
    });
  };

  private handleSaveResources = () => {
    const newResources: Partial<{
      facility: Facility;
      equipment: Equipment;
      medicine: Medicine;
    }> = {
      facility: this.selectedFacility,
      equipment: this.selectedEquipment,
      medicine: this.selectedMedicine,
    };

    this.handleSaveResourcesOnAppointment(this.appointment, newResources).then(() => {
      this.appointment = {
        ...this.appointment,
        facilities: this.selectedFacility ? [this.selectedFacility] : undefined,
        equipment: this.selectedEquipment ? [this.selectedEquipment] : undefined,
        medicine: this.selectedMedicine ? [this.selectedMedicine] : undefined,
      };
      this.editingResources = false;
    });
  };

  private handleAddPrescriptionNameChange = (event: Event) => {
    this.addingPrescriptionName = (event.target as HTMLSelectElement).value;
  };

  private handleAddPrescriptionDateChange = (
    type: 'start' | 'end',
    part: 'day' | 'month' | 'year',
    event: Event,
  ) => {
    const value: number = parseInt((event.target as HTMLSelectElement).value, 10);
    if (type === 'start') {
      this.addingPrescriptionStart = updateDatePart(this.addingPrescriptionStart, part, value);
    } else {
      this.addingPrescriptionEnd = updateDatePart(this.addingPrescriptionEnd, part, value);
    }
  };

  private handleAddPrescriptionDoctorsNoteChange = (event: Event) => {
    this.addingPrescriptionDoctorsNote = (event.target as HTMLSelectElement).value;
  };

  private handleUpdatePrescriptionNameChange = (event: Event) => {
    this.editingPrescriptionNewName = (event.target as HTMLSelectElement).value;
  };

  private handleUpdatePrescriptionDateChange = (
    type: 'start' | 'end',
    part: 'day' | 'month' | 'year',
    event: Event,
  ) => {
    const value: number = parseInt((event.target as HTMLSelectElement).value, 10);
    if (type === 'start') {
      this.editingPrescriptionNewStart = updateDatePart(
        this.editingPrescriptionNewStart,
        part,
        value,
      );
    } else {
      this.editingPrescriptionNewEnd = updateDatePart(this.editingPrescriptionNewEnd, part, value);
    }
  };

  private handleUpdatePrescriptionDoctorsNoteChange = (event: Event) => {
    this.editingPrescriptionNewDoctorsNote = (event.target as HTMLSelectElement).value;
  };

  private handleAddPrescription = () => {
    const newPrescription: NewPrescription = {
      name: this.addingPrescriptionName,
      start: this.addingPrescriptionStart,
      end: this.addingPrescriptionEnd,
      doctorsNote: this.addingPrescriptionDoctorsNote,
      patientId: this.appointment.patient.id,
    };

    this.handleAddPrescriptionForAppointment(this.appointment, newPrescription).then(
      prescription => {
        if (!prescription) return;

        if (!this.appointment.prescriptions) {
          this.appointment.prescriptions = [prescription];
        } else {
          this.appointment.prescriptions.push(prescription);
        }

        this.addingPrescription = false;
        this.addingPrescriptionName = '';
        this.addingPrescriptionStart = new Date();
        this.addingPrescriptionEnd = new Date(new Date().setDate(new Date().getDate() + 7));
        this.addingPrescriptionDoctorsNote = '';
      },
    );
  };

  private handleUpdatePrescription = () => {
    const updatedPrescription: PrescriptionDisplay & { doctorsNote?: string } = {
      id: this.editingPrescription.id,
      name: this.editingPrescriptionNewName,
      start: this.editingPrescriptionNewStart,
      end: this.editingPrescriptionNewEnd,
      appointmentId: this.editingPrescription.appointmentId,
      doctorsNote: this.editingPrescriptionNewDoctorsNote,
    };

    this.handleUpdatePrescriptionForAppointment(
      this.editingPrescription.id,
      updatedPrescription,
    ).then((prescription: Prescription) => {
      if (!prescription) return;
      const index: number = this.appointment.prescriptions.findIndex(
        (prescription: PrescriptionDisplay) => prescription.id === prescription.id,
      );
      if (index !== -1) {
        this.appointment.prescriptions[index] = prescription;
      }
      this.editingPrescription = null;

      this.editingPrescriptionNewName = '';
      this.editingPrescriptionNewStart = null;
      this.editingPrescriptionNewEnd = null;
      this.editingPrescriptionNewDoctorsNote = null;
    });
  };

  private handleDeletePrescription = () => {
    this.handleDeletePrescriptionFromAppointment(this.deletingPrescription).then(() => {
      this.appointment = {
        ...this.appointment,
        prescriptions: this.appointment.prescriptions.filter(
          (prescription: PrescriptionDisplay): boolean =>
            prescription.id !== this.deletingPrescription.id,
        ),
      };
    });
  };

  private cancelledByStatus = (): ' by doctor' | ' by patient' | undefined => {
    if (!this.appointment.canceledBy) return;
    return this.appointment.canceledBy === 'doctor' ? ' by doctor' : ' by patient';
  };

  render() {
    if (!this.appointment) return null;

    return (
      <div class="w-full max-w-md">
        <div class="relative w-full max-w-md">
          <div class="absolute top-1/2 left-0 -translate-x-0 -translate-y-1/2 transform">
            <md-icon-button onClick={this.handleResetSelection}>
              <md-icon class="text-gray-600">arrow_forward</md-icon>
            </md-icon-button>
          </div>

          <h2 class="mb-6 w-full text-center text-xl font-medium text-[#7357be]">
            Appointment details
          </h2>
        </div>

        <div class="mb-6 w-full max-w-md rounded-md bg-gray-200 px-4 py-3">
          <div class="mb-1 flex w-full flex-row items-center justify-between">
            <div class="flex flex-row items-center gap-x-2 text-gray-500">
              <md-icon style={{ fontSize: '16px' }}>calendar_month</md-icon>
              Date
            </div>
            <span class="font-medium text-gray-600">
              {formatDate(this.appointment.appointmentDateTime)}
            </span>
          </div>

          <div class="mb-1 flex w-full flex-row items-center justify-between">
            <div class="flex flex-row items-center gap-x-2 text-gray-500">
              <md-icon style={{ fontSize: '16px' }}>schedule</md-icon>
              Time
            </div>
            <span class="font-medium text-gray-600">
              {formatTime(this.appointment.appointmentDateTime)}
            </span>
          </div>

          <div class="mb-1 flex w-full flex-row items-center justify-between">
            <div class="flex flex-row items-center gap-x-2 text-gray-500">
              <md-icon style={{ fontSize: '16px' }}>format_list_bulleted</md-icon>
              Type
            </div>
            <span class="font-medium text-gray-600">
              {formatAppointmentType(this.appointment.type)}
            </span>
          </div>

          {this.isDoctor && (
            <div class="mb-1 flex w-full flex-row items-center justify-between">
              <div class="flex flex-row items-center gap-x-2 text-gray-500">
                <md-icon style={{ fontSize: '16px' }}>person</md-icon>
                Patient
              </div>
              <span class="font-medium text-gray-600">
                {this.appointment.patient.firstName} {this.appointment.patient.lastName}
              </span>
            </div>
          )}
          {!this.isDoctor && (
            <div class="mb-1 flex w-full flex-row items-center justify-between">
              <div class="flex flex-row items-center gap-x-2 text-gray-500">
                <md-icon style={{ fontSize: '16px' }}>person</md-icon>
                Doctor
              </div>
              <span class="font-medium text-gray-600">
                Dr. {this.appointment.doctor.firstName} {this.appointment.doctor.lastName}
              </span>
            </div>
          )}

          <div class="flex w-full flex-row items-center justify-between">
            <div class="flex flex-row items-center gap-x-2 text-gray-500">
              <md-icon style={{ fontSize: '16px' }}>info</md-icon>
              Status
            </div>
            <span class="font-medium text-gray-600">
              {this.appointment.status[0].toUpperCase() + this.appointment.status.slice(1)}
              {this.appointment.status === 'cancelled' && this.cancelledByStatus()}
            </span>
          </div>
        </div>

        <div class="mb-6 max-h-32 w-full max-w-md overflow-y-auto rounded-md bg-gray-200 px-4 py-3">
          <div class="flex flex-row items-center gap-x-2 text-gray-500">
            <md-icon style={{ fontSize: '16px' }}>description</md-icon>
            Reason
          </div>
          {this.appointment.reason && (
            <p class="mt-1 ml-1 text-sm font-medium text-wrap text-gray-600">
              {this.appointment.reason}
            </p>
          )}
        </div>

        {this.appointment.status === 'cancelled' && (
          <div class="mb-6 max-h-32 w-full max-w-md overflow-y-auto rounded-md bg-gray-200 px-4 py-3">
            <div class="flex flex-row items-center gap-x-2 text-gray-500">
              <md-icon style={{ fontSize: '16px' }}>description</md-icon>
              Cancellation reason
            </div>
            {this.appointment.cancellationReason && (
              <p class="mt-1 ml-1 text-sm font-medium text-wrap text-gray-600">
                {this.appointment.cancellationReason}
              </p>
            )}
          </div>
        )}

        {this.appointment.status === 'denied' && (
          <div class="mb-6 max-h-32 w-full max-w-md overflow-y-auto rounded-md bg-gray-200 px-4 py-3">
            <div class="flex flex-row items-center gap-x-2 text-gray-500">
              <md-icon style={{ fontSize: '16px' }}>description</md-icon>
              Denial reason
            </div>
            {this.appointment.denialReason && (
              <p class="mt-1 ml-1 text-sm font-medium text-wrap text-gray-600">
                {this.appointment.denialReason}
              </p>
            )}
          </div>
        )}

        {/* Appointment Resources */}
        {this.isDoctor && (
          <xcastven-xkilian-project-appointment-resources
            isDoctor={this.isDoctor}
            appointment={this.appointment}
            editingResources={this.editingResources}
            setEditingResources={(editingResources: boolean) =>
              (this.editingResources = editingResources)
            }
            availableFacilities={this.availableFacilities}
            selectedFacility={this.selectedFacility}
            handleSelectFacility={this.handleSelectFacility}
            availableEquipment={this.availableEquipment}
            selectedEquipment={this.selectedEquipment}
            handleSelectEquipment={this.handleSelectEquipment}
            availableMedicine={this.availableMedicine}
            selectedMedicine={this.selectedMedicine}
            handleSelectMedicine={this.handleSelectMedicine}
            handleSaveResources={this.handleSaveResources}
          />
        )}

        {/* Appointment Prescriptions */}
        {['scheduled', 'completed'].includes(this.appointment.status) && (
          <xcastven-xkilian-project-appointment-prescriptions
            isDoctor={this.isDoctor}
            appointment={this.appointment}
            prescriptionsExpanded={this.prescriptionsExpanded}
            handleSelectPrescription={this.handleSelectPrescription}
            addingPrescription={this.addingPrescription}
            setAddingPrescription={(addingPrescription: boolean) =>
              (this.addingPrescription = addingPrescription)
            }
            addingPrescriptionName={this.addingPrescriptionName}
            addingPrescriptionStart={this.addingPrescriptionStart}
            addingPrescriptionEnd={this.addingPrescriptionEnd}
            addingPrescriptionDoctorsNote={this.addingPrescriptionDoctorsNote}
            handleAddPrescriptionNameChange={this.handleAddPrescriptionNameChange}
            handleAddPrescriptionDateChange={this.handleAddPrescriptionDateChange}
            handleAddPrescriptionDoctorsNoteChange={this.handleAddPrescriptionDoctorsNoteChange}
            handleAddPrescription={this.handleAddPrescription}
            editingPrescription={this.editingPrescription}
            setEditingPrescription={(editingPrescription: PrescriptionDisplay) =>
              (this.editingPrescription = editingPrescription)
            }
            editingPrescriptionNewName={this.editingPrescriptionNewName}
            editingPrescriptionNewStart={this.editingPrescriptionNewStart}
            editingPrescriptionNewEnd={this.editingPrescriptionNewEnd}
            editingPrescriptionNewDoctorsNote={this.editingPrescriptionNewDoctorsNote}
            handleUpdatePrescriptionNameChange={this.handleUpdatePrescriptionNameChange}
            handleUpdatePrescriptionDateChange={this.handleUpdatePrescriptionDateChange}
            handleUpdatePrescriptionDoctorsNoteChange={
              this.handleUpdatePrescriptionDoctorsNoteChange
            }
            handleUpdatePrescription={this.handleUpdatePrescription}
            deletingPrescription={this.deletingPrescription}
            setDeletingPrescription={(deletingPrescription: PrescriptionDisplay) =>
              (this.deletingPrescription = deletingPrescription)
            }
            handleDeletePrescription={this.handleDeletePrescription}
          />
        )}

        {this.isDoctor && !(this.appointment.status in ['scheduled', 'requested']) && (
          <p class="mb-6 text-center text-sm text-wrap text-gray-600">
            {this.getDoctorAppointmentStatusMessage(this.appointment.status)}
          </p>
        )}
        {!this.isDoctor && !(this.appointment.status in ['scheduled']) && (
          <p class="mb-6 text-center text-sm text-wrap text-gray-600">
            {this.getPatientAppointmentStatusMessage(this.appointment.status)}
          </p>
        )}

        <xcastven-xkilian-project-appointment-actions
          isDoctor={this.isDoctor}
          appointment={this.appointment}
          handleCancelDoctorAppointment={() => {
            this.cancelling = !this.cancelling;
            this.rescheduling = false;
            this.denying = false;
          }}
          handleAcceptAppointment={this.handleAccept}
          handleDenyAppointment={() => {
            this.denying = !this.denying;
            this.cancelling = false;
            this.rescheduling = false;
          }}
          handleRescheduleAppointment={() => {
            this.rescheduling = !this.rescheduling;
            this.cancelling = false;
            this.denying = false;
            if (this.rescheduling) {
              this.loadReschedulingDoctors();
            }
          }}
          handleCancelPatientAppointment={() => {
            this.cancelling = !this.cancelling;
            this.rescheduling = false;
            this.denying = false;
          }}
        />

        {this.rescheduling && !this.isDoctor && (
          <xcastven-xkilian-project-appointment-reschedule
            rescheduling={this.rescheduling}
            reschedulingAppointmentDate={this.reschedulingAppointmentDate}
            reschedulingAppointmentTime={this.reschedulingAppointmentTime}
            reschedulingAppointmentDoctor={this.reschedulingAppointmentDoctor}
            reschedulingAppointmentReason={this.reschedulingAppointmentReason}
            reschedulingAvailableDoctors={this.reschedulingAvailableDoctors}
            reschedulingAvailableTimes={this.reschedulingAvailableTimes}
            handleDateSelectInput={this.handleDateSelectInput}
            handleDoctorSelectInput={this.handleDoctorSelectInput}
            handleTimeSelectInput={this.handleTimeSelectInput}
            handleReschedulingAppointmentReasonChange={
              this.handleReschedulingAppointmentReasonChange
            }
            handleReschedule={this.handleReschedule}
          />
        )}

        {this.cancelling && (
          <xcastven-xkilian-project-appointment-cancel
            cancelling={this.cancelling}
            setCancelling={(cancelling: boolean) => (this.cancelling = cancelling)}
            cancellingAppointmentReason={this.cancellingAppointmentReason}
            handleCancellingAppointmentReasonChange={this.handleCancellingAppointmentReasonChange}
            handleCancel={this.handleCancel}
          />
        )}

        {this.denying && this.isDoctor && (
          <xcastven-xkilian-project-appointment-deny
            denying={this.denying}
            setDenying={(denying: boolean) => (this.denying = denying)}
            denyingAppointmentReason={this.denyingAppointmentReason}
            handleDenyingAppointmentReasonChange={(e: Event) => {
              this.denyingAppointmentReason = (e.target as HTMLInputElement).value;
            }}
            handleDeny={this.handleDeny}
            disabled={this.denyingAppointmentReason === ''}
          />
        )}
      </div>
    );
  }
}
