import {
  AppointmentDisplay,
  AppointmentStatus,
  ConditionDisplay,
  PrescriptionDisplay,
  UserRole,
} from '../../api/generated';
import {
  AppointmentStatusColor,
  ConditionOrderColors,
  DAYS_OF_WEEK,
  formatDate,
  formatTime,
  PrescriptionOrderColors,
} from '../../utils/utils';
import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-calendar',
  shadow: false,
})
export class Calendar {
  @Prop() user: { email: string; role: UserRole };
  @Prop() isDoctor: boolean;
  @Prop() appointments: Array<AppointmentDisplay>;
  @Prop() conditions: Array<ConditionDisplay>;
  @Prop() prescriptions: Array<PrescriptionDisplay>;
  @Prop() getConditionsForDate: (date: Date) => Array<ConditionDisplay>;
  @Prop() getPrescriptionsForDate: (date: Date) => Array<PrescriptionDisplay>;
  @Prop() handleSelectDate: (date: Date) => void;
  @Prop() handleSelectAppointment: (appointment: AppointmentDisplay) => void;
  @Prop() handleSelectCondition: (condition: ConditionDisplay) => void;
  @Prop() handleSelectPrescription: (prescription: PrescriptionDisplay) => void;
  @Prop() handleSelectAppointmentStatusGroup: (date: Date, status: AppointmentStatus) => void;
  @Prop() hoveredConditionId: string;
  @Prop() setHoveredConditionId: (value: string | null) => void;
  @Prop() hoveredPrescriptionId: string;
  @Prop() setHoveredPrescriptionId: (value: string | null) => void;
  @Prop() currentViewMonth: number;
  @Prop() currentViewYear: number;
  @Prop() handlePreviousMonth: () => void;
  @Prop() handleNextMonth: () => void;

  private getDaysInMonth = (year: number, month: number) => {
    return new Date(year, month + 1, 0).getDate();
  };

  private getFirstDayOfMonth = (year: number, month: number) => {
    const firstDay: number = new Date(year, month, 1).getDay();
    return firstDay === 0 ? 6 : firstDay - 1;
  };

  private getAppointmentTooltipContent = (appointment: AppointmentDisplay) => {
    const displayStatus: string = appointment.status[0].toUpperCase() + appointment.status.slice(1);

    return (
      <div
        class="pointer-events-none absolute z-100 rounded bg-black px-2 py-1 text-xs whitespace-nowrap text-white opacity-0 transition-opacity duration-200 group-hover:opacity-100"
        style={{
          top: '4px',
          right: 'calc(100% + 8px)',
        }}
      >
        <div class="flex w-full flex-col gap-y-1">
          <span class="w-full text-center">{appointment.type}</span>

          <div class="flex flex-col items-center justify-center gap-y-1">
            <div class="flex items-center space-x-1 text-gray-400">
              <md-icon style={{ fontSize: '16px' }}>event</md-icon>
              <div>{formatDate(appointment.appointmentDateTime)}</div>
            </div>

            <div class="flex items-center space-x-1 text-gray-400">
              <md-icon style={{ fontSize: '16px' }}>timer</md-icon>
              <div>{formatTime(appointment.appointmentDateTime)}</div>
            </div>

            <div class="flex items-center space-x-1 text-gray-400">
              <md-icon style={{ fontSize: '16px' }}>format_list_bulleted</md-icon>
              <div>{displayStatus}</div>
            </div>
          </div>
        </div>
      </div>
    );
  };

  private getConditionTooltipContent = (condition: ConditionDisplay) => {
    return (
      <div class="pointer-events-none absolute bottom-full left-1/2 z-100 mb-1 -translate-x-1/2 transform rounded bg-black px-2 py-1 text-xs whitespace-nowrap text-white opacity-0 transition-opacity duration-200 group-hover:opacity-100">
        <div class="flex w-full flex-col gap-y-1">
          <span class="w-full text-center">{condition.name}</span>

          <div class="flex flex-col items-center justify-center gap-y-1">
            <div class="flex items-center space-x-1 text-gray-400">
              <md-icon style={{ fontSize: '16px' }}>line_start_circle</md-icon>
              <div>{formatDate(condition.start)}</div>
            </div>

            {condition.end && (
              <div class="flex items-center space-x-1 text-gray-400">
                <md-icon style={{ fontSize: '16px' }}>line_end_circle</md-icon>
                <div>{formatDate(condition.end)}</div>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  };

  private getPrescriptionTooltipContent = (
    prescription: PrescriptionDisplay,
    offsetRight: string,
  ) => {
    return (
      <div
        class="pointer-events-none absolute z-100 rounded bg-black px-2 py-1 text-xs whitespace-nowrap text-white opacity-0 transition-opacity duration-200 group-hover:opacity-100"
        style={{
          top: '4px',
          right: `calc(${offsetRight} + 24px)`,
        }}
      >
        <div class="flex w-full flex-col gap-y-1">
          <span class="w-full text-center">{prescription.name}</span>

          <div class="flex flex-col items-center justify-center gap-y-1">
            <div class="flex items-center space-x-1 text-gray-400">
              <md-icon class="text-base">line_start_circle</md-icon>
              <div>{formatDate(prescription.start)}</div>
            </div>

            {prescription.end && (
              <div class="flex items-center space-x-1 text-gray-400">
                <md-icon class="text-base">line_end_circle</md-icon>
                <div>{formatDate(prescription.end)}</div>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  };

  private renderPatientsCalendar = () => {
    const year: number = this.currentViewYear;
    const month: number = this.currentViewMonth;
    const today: Date = new Date();
    today.setHours(0, 0, 0, 0);

    const daysInMonth: number = this.getDaysInMonth(year, month);
    const firstDayOfMonth: number = this.getFirstDayOfMonth(year, month);

    const prevMonthDays = [];
    const currentMonthDays = [];
    const nextMonthDays = [];

    const appointmentsByDate = new Map<string, Array<AppointmentDisplay>>();

    this.appointments.forEach((appointment: AppointmentDisplay) => {
      const date: Date = appointment.appointmentDateTime;
      const dateKey: string = date.getFullYear() + '-' + date.getMonth() + '-' + date.getDate();

      if (!appointmentsByDate.has(dateKey)) {
        appointmentsByDate.set(dateKey, []);
      }

      appointmentsByDate.get(dateKey).push(appointment);
    });

    const daysInPrevMonth: number = this.getDaysInMonth(year, month - 1);
    for (let i: number = firstDayOfMonth - 1; i >= 0; i--) {
      prevMonthDays.push(
        <div
          class="flex w-full items-center justify-center bg-gray-200 px-3 py-2 text-center text-sm text-gray-400"
          onClick={this.handlePreviousMonth}
        >
          {daysInPrevMonth - i}
        </div>,
      );
    }

    const prescriptionColorMap: Record<string, string> = {};
    this.prescriptions.forEach((prescription: PrescriptionDisplay, index: number) => {
      prescriptionColorMap[prescription.id] =
        PrescriptionOrderColors[index % PrescriptionOrderColors.length];
    });

    const conditionColorMap: Record<string, string> = {};
    this.conditions.forEach((condition: ConditionDisplay, index: number) => {
      conditionColorMap[condition.id] = ConditionOrderColors[index % ConditionOrderColors.length];
    });

    const calculateConditionHeights = () => {
      const sortedConditions: Array<ConditionDisplay> = [...this.conditions].sort(
        (a: ConditionDisplay, b: ConditionDisplay): number => a.start.getTime() - b.start.getTime(),
      );

      const timeline: Map<number, Set<string>> = new Map();

      sortedConditions.forEach((condition: ConditionDisplay) => {
        const startTime: number = condition.start.getTime();
        const endTime: number = condition.end ? condition.end.getTime() : startTime;

        let currentDate: Date = new Date(startTime);
        currentDate.setHours(0, 0, 0, 0);

        while (currentDate.getTime() <= endTime) {
          const timeKey: number = currentDate.getTime();

          if (!timeline.has(timeKey)) {
            timeline.set(timeKey, new Set());
          }

          timeline.get(timeKey).add(condition.id);

          currentDate.setDate(currentDate.getDate() + 1);
        }
      });

      const overlaps: Map<string, Set<string>> = new Map();

      sortedConditions.forEach((condition: ConditionDisplay) => {
        overlaps.set(condition.id, new Set());
      });

      Array.from(timeline.values()).forEach(activeConditions => {
        if (activeConditions.size > 1) {
          const activeArray: Array<string> = Array.from(activeConditions);
          for (let i = 0; i < activeArray.length; i++) {
            for (let j = i + 1; j < activeArray.length; j++) {
              overlaps.get(activeArray[i]).add(activeArray[j]);
              overlaps.get(activeArray[j]).add(activeArray[i]);
            }
          }
        }
      });

      const conditionHeightMap: Record<string, number> = {};

      sortedConditions.forEach(condition => {
        const usedHeights = new Set<number>();
        overlaps.get(condition.id).forEach(overlapId => {
          if (conditionHeightMap[overlapId] !== undefined) {
            usedHeights.add(conditionHeightMap[overlapId]);
          }
        });

        let heightLevel: number = 0;
        while (usedHeights.has(10 + heightLevel * 6)) {
          heightLevel++;
        }

        conditionHeightMap[condition.id] = 10 + heightLevel * 6;
      });

      return conditionHeightMap;
    };

    const conditionHeightMap: Record<string, number> = calculateConditionHeights();

    for (let i: number = 1; i <= daysInMonth; i++) {
      const currentDate: Date = new Date(year, month, i);
      currentDate.setHours(0, 0, 0, 0);

      const dateKey: string = year + '-' + month + '-' + i;
      const appointmentsForDay: Array<AppointmentDisplay> = appointmentsByDate.get(dateKey) ?? [];

      const isToday: boolean =
        i === today.getDate() &&
        this.currentViewMonth === today.getMonth() &&
        this.currentViewYear === today.getFullYear();
      const isPastDate: boolean = currentDate < today;

      const conditionsForDate: Array<ConditionDisplay> = this.getConditionsForDate(currentDate);

      const prescriptionsForDate: Array<PrescriptionDisplay> =
        this.getPrescriptionsForDate(currentDate);
      let offset: number = 0;
      let zIndex: number = 10;
      let ellipsisDisplayed: boolean = false;

      currentMonthDays.push(
        <div
          role="button"
          class={`relative flex w-full flex-col items-center justify-center text-sm ${isPastDate ? 'text-gray-400' : 'bg-[#f0eafa] hover:border-2 hover:border-[#9d83c6]'} `}
          onClick={(event: MouseEvent) => {
            event.stopPropagation();
            this.handleSelectDate(currentDate);
          }}
        >
          <div
            class={`relative mb-[1px] flex w-full flex-row items-center justify-between py-1 ${isToday && 'bg-[#7357be] text-white'}`}
          >
            <div class="h-full w-full text-center">{i}</div>
            {prescriptionsForDate.length > 0 &&
              prescriptionsForDate
                .reverse()
                .map((prescription: PrescriptionDisplay, index: number) => {
                  const prescriptionOffset: string = (4 + offset).toString() + 'px';
                  const prescriptionZIndex: string = zIndex.toString();
                  offset += 16;
                  zIndex -= 1;

                  const prescriptionColor: string = prescriptionColorMap[prescription.id];

                  if (index >= 2 && !ellipsisDisplayed) {
                    ellipsisDisplayed = true;
                    return (
                      <span style={{ position: 'absolute', top: '4px', right: '40px' }}>...</span>
                    );
                  }

                  return (
                    <div class="group inline-block">
                      <md-icon
                        class="cursor-pointer transition-all duration-200"
                        style={{
                          fontSize:
                            this.hoveredPrescriptionId === prescription.id ? '24px' : '20px',
                          position: 'absolute',
                          top: '4px',
                          right: prescriptionOffset,
                          zIndex: prescriptionZIndex,
                          color: prescriptionColor,
                        }}
                        onMouseEnter={() => this.setHoveredPrescriptionId(prescription.id)}
                        onMouseLeave={() => this.setHoveredPrescriptionId(null)}
                        onClick={(event: Event) => {
                          event.stopPropagation();
                          this.handleSelectPrescription(prescriptionsForDate[0]);
                        }}
                      >
                        medication
                      </md-icon>
                      {this.getPrescriptionTooltipContent(prescription, prescriptionOffset)}
                    </div>
                  );
                })}
          </div>

          <div class="relative flex h-full w-full flex-row flex-wrap items-start justify-center gap-1">
            {appointmentsForDay.length > 0 &&
              appointmentsForDay.map((appointment, index: number) => (
                <button
                  key={appointment.id + index}
                  class="circle-container group relative"
                  onClick={(event: MouseEvent) => {
                    event.stopPropagation();
                    this.handleSelectAppointment(appointment);
                  }}
                >
                  <div
                    class={`circle h-3 w-3 cursor-pointer rounded-full transition-all duration-200 hover:h-4 hover:w-4 sm:h-[0.875rem] sm:w-[0.875rem] sm:hover:h-5 sm:hover:w-5 md:h-4 md:w-4 md:hover:h-6 md:hover:w-6 lg:h-5 lg:w-5 lg:hover:h-7 lg:hover:w-7`}
                    style={{
                      backgroundColor: AppointmentStatusColor[appointment.status].background,
                    }}
                  ></div>

                  {this.getAppointmentTooltipContent(appointment)}
                </button>
              ))}
          </div>

          {conditionsForDate.map((condition: ConditionDisplay) => {
            const isStartDate: boolean = currentDate.getTime() === condition.start.getTime();
            const isEndDate: boolean =
              condition.end && currentDate.getTime() === condition.end.getTime();

            const conditionHeight: number = conditionHeightMap[condition.id];
            const conditionColor: string = conditionColorMap[condition.id];

            return (
              <div
                key={condition.id}
                class="group absolute right-0 left-0 cursor-pointer transition-all duration-200"
                style={{
                  bottom: '0px',
                  height:
                    this.hoveredConditionId === condition.id
                      ? `${conditionHeight + 6}px`
                      : `${conditionHeight}px`,
                  zIndex: (50 - conditionHeight).toString(),
                  backgroundColor: conditionColor,
                }}
                onMouseEnter={() => this.setHoveredConditionId(condition.id)}
                onMouseLeave={() => this.setHoveredConditionId(null)}
                onClick={(event: MouseEvent) => {
                  event.stopPropagation();
                  this.handleSelectCondition(condition);
                }}
              >
                <div class="relative">
                  {isStartDate && (
                    <div
                      class="absolute top-0 left-0 h-5 w-5 -translate-y-3 transform rounded-tr-full"
                      style={{
                        backgroundColor: conditionColor,
                      }}
                    ></div>
                  )}

                  {isEndDate && (
                    <div
                      class="absolute top-0 right-0 h-5 w-5 -translate-y-3 transform rounded-tl-full"
                      style={{
                        backgroundColor: conditionColor,
                      }}
                    ></div>
                  )}

                  {this.getConditionTooltipContent(condition)}
                </div>
              </div>
            );
          })}
        </div>,
      );
    }

    const totalCells = 42;
    const remainingCells = totalCells - (prevMonthDays.length + currentMonthDays.length);
    for (let i = 1; i <= remainingCells; i++) {
      nextMonthDays.push(
        <div
          class="flex w-full items-center justify-center bg-gray-200 px-3 py-2 text-center text-sm text-gray-400"
          onClick={this.handleNextMonth}
        >
          {i}
        </div>,
      );
    }

    return [...prevMonthDays, ...currentMonthDays, ...nextMonthDays];
  };

  private getAppointmentStatusGroupTooltipContent = (
    appointmentStatus: AppointmentStatus,
    appointmentCount: number,
  ) => {
    return (
      <div
        class="pointer-events-none absolute z-100 rounded bg-black px-2 py-1 text-xs whitespace-nowrap text-white opacity-0 transition-opacity duration-200 group-hover:opacity-100"
        style={{
          top: '4px',
          right: 'calc(100% + 8px)',
        }}
      >
        <div class="flex w-full flex-col gap-y-1">
          <span class="w-full text-center">
            {appointmentStatus[0].toUpperCase() + appointmentStatus.slice(1)}
          </span>

          <div class="flex items-center space-x-1 text-gray-400">
            <md-icon style={{ fontSize: '16px' }}>calendar_month</md-icon>
            <div>
              {appointmentCount} appointment{appointmentCount !== 1 && 's'}
            </div>
          </div>
        </div>
      </div>
    );
  };

  private renderDoctorsCalendar = () => {
    const year: number = this.currentViewYear;
    const month: number = this.currentViewMonth;
    const today: Date = new Date();
    today.setHours(0, 0, 0, 0);

    const daysInMonth: number = this.getDaysInMonth(year, month);
    const firstDayOfMonth: number = this.getFirstDayOfMonth(year, month);

    const prevMonthDays = [];
    const currentMonthDays = [];
    const nextMonthDays = [];

    const appointmentsByDateAndStatus = new Map<
      string,
      Map<AppointmentStatus, Array<AppointmentDisplay>>
    >();

    this.appointments.forEach((appointment: AppointmentDisplay) => {
      const date: Date = appointment.appointmentDateTime;
      const dateKey: string = date.getFullYear() + '-' + date.getMonth() + '-' + date.getDate();
      const status: AppointmentStatus = appointment.status;

      if (!appointmentsByDateAndStatus.has(dateKey)) {
        appointmentsByDateAndStatus.set(dateKey, new Map());
      }

      const appointmentsByStatus: Map<
        AppointmentStatus,
        Array<AppointmentDisplay>
      > = appointmentsByDateAndStatus.get(dateKey);

      for (const statusType of ['requested', 'scheduled', 'denied', 'completed', 'cancelled']) {
        if (!appointmentsByStatus.has(statusType as AppointmentStatus)) {
          appointmentsByStatus.set(statusType as AppointmentStatus, []);
        }
      }

      appointmentsByStatus.get(status).push(appointment);
    });

    const daysInPrevMonth: number = this.getDaysInMonth(year, month - 1);
    for (let i: number = firstDayOfMonth - 1; i >= 0; i--) {
      prevMonthDays.push(
        <div
          class="flex w-full items-center justify-center bg-gray-200 px-3 py-2 text-center text-sm text-gray-400"
          onClick={this.handlePreviousMonth}
        >
          {daysInPrevMonth - i}
        </div>,
      );
    }

    for (let i: number = 1; i <= daysInMonth; i++) {
      const currentDate: Date = new Date(year, month, i);
      currentDate.setHours(0, 0, 0, 0);

      const dateKey: string = year + '-' + month + '-' + i;

      let appointmentsForDay: Map<AppointmentStatus, Array<AppointmentDisplay>>;
      if (appointmentsByDateAndStatus.has(dateKey)) {
        appointmentsForDay = appointmentsByDateAndStatus.get(dateKey);
      }

      const isToday: boolean =
        i === today.getDate() &&
        this.currentViewMonth === today.getMonth() &&
        this.currentViewYear === today.getFullYear();
      const isPastDate: boolean = currentDate.getTime() >= today.getTime();

      currentMonthDays.push(
        <div
          role="button"
          class={`relative flex w-full cursor-pointer flex-col items-center justify-start text-sm hover:border-2 hover:border-[#9d83c6]`}
          onClick={(event: MouseEvent) => {
            event.stopPropagation();
            this.handleSelectDate(currentDate);
          }}
        >
          <div
            class={`relative mb-[1px] flex w-full flex-col items-center py-1 ${isToday && 'bg-[#7357be] text-white'}`}
          >
            <div class="h-full w-full text-center">{i}</div>
          </div>

          <div class="flex h-full w-full flex-row flex-wrap items-start justify-center gap-1">
            {appointmentsForDay &&
              isPastDate &&
              ['scheduled', 'requested'].map((appointmentStatus: AppointmentStatus) => {
                if (appointmentsForDay.get(appointmentStatus).length === 0) return;

                return (
                  <button
                    class="circle-container group relative"
                    onClick={(event: MouseEvent) => {
                      event.stopPropagation();
                      this.handleSelectAppointmentStatusGroup(currentDate, appointmentStatus);
                    }}
                  >
                    <div
                      class={`circle flex h-5 w-5 cursor-pointer items-center justify-center rounded-full text-white transition-all duration-200 hover:h-6 hover:w-6 sm:h-6 sm:w-6 sm:hover:h-7 sm:hover:w-7 md:h-7 md:w-7 md:hover:h-8 md:hover:w-8 lg:h-8 lg:w-8 lg:hover:h-9 lg:hover:w-9`}
                      style={{
                        backgroundColor: AppointmentStatusColor[appointmentStatus].background,
                      }}
                    >
                      {appointmentsForDay.get(appointmentStatus).length}
                    </div>

                    {this.getAppointmentStatusGroupTooltipContent(
                      appointmentStatus,
                      appointmentsForDay.get(appointmentStatus).length,
                    )}
                  </button>
                );
              })}

            {appointmentsForDay &&
              !isPastDate &&
              ['completed', 'cancelled'].map((appointmentStatus: AppointmentStatus) => {
                if (appointmentsForDay.get(appointmentStatus).length === 0) return;

                return (
                  <button
                    class="circle-container group relative"
                    onClick={(event: MouseEvent) => {
                      event.stopPropagation();
                      this.handleSelectAppointmentStatusGroup(currentDate, appointmentStatus);
                    }}
                  >
                    <div
                      class={`circle flex h-5 w-5 cursor-pointer items-center justify-center rounded-full text-white transition-all duration-200 hover:h-6 hover:w-6 sm:h-6 sm:w-6 sm:hover:h-7 sm:hover:w-7 md:h-7 md:w-7 md:hover:h-8 md:hover:w-8 lg:h-8 lg:w-8 lg:hover:h-9 lg:hover:w-9`}
                      style={{
                        backgroundColor: AppointmentStatusColor[appointmentStatus].background,
                      }}
                    >
                      {appointmentsForDay.get(appointmentStatus).length}
                    </div>

                    {this.getAppointmentStatusGroupTooltipContent(
                      appointmentStatus,
                      appointmentsForDay.get(appointmentStatus).length,
                    )}
                  </button>
                );
              })}
          </div>
        </div>,
      );
    }

    const totalCells = 42;
    const remainingCells: number = totalCells - (prevMonthDays.length + currentMonthDays.length);
    for (let i = 1; i <= remainingCells; i++) {
      nextMonthDays.push(
        <div
          class="flex w-full items-center justify-center bg-gray-200 px-3 py-2 text-center text-sm text-gray-400"
          onClick={this.handleNextMonth}
        >
          {i}
        </div>,
      );
    }

    return [...prevMonthDays, ...currentMonthDays, ...nextMonthDays];
  };

  render() {
    return (
      <div class="flex h-[calc(100vh-96px)] flex-1 flex-col overflow-x-hidden overflow-y-auto">
        <div class="grid grid-cols-7 bg-[#d8c7ed]">
          {DAYS_OF_WEEK.map((day: { short: string; long: string }) => (
            <div class="px-4 py-3 text-center text-sm font-medium text-[#7357be]">
              <span class="inline md:hidden">{day.short}</span>
              <span class="hidden md:inline">{day.long}</span>
            </div>
          ))}
        </div>

        <div class="relative z-20 grid flex-1 grid-cols-7 grid-rows-6 divide-x divide-y divide-[#d8c7ed] border-x border-t border-[#d8c7ed]">
          {!this.isDoctor ? this.renderPatientsCalendar() : this.renderDoctorsCalendar()}
        </div>
      </div>
    );
  }
}
