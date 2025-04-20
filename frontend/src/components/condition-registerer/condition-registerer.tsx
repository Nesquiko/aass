import { Api } from '../../api/api';
import { NewCondition } from '../../api/generated';
import { User } from '../../components';
import { Navigate } from '../../utils/types';
import { DAYS_OF_WEEK, formatDate, formatDateDelta, MONTHS, TODAY } from '../../utils/utils';
import { toastService } from '../services/toast-service';
import { Component, h, Prop, State } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-condition-registerer',
  shadow: false,
})
export class ConditionRegisterer {
  @Prop() navigate: Navigate;
  @Prop() api: Api;
  @Prop() user: User;
  @Prop() startDate: Date = null;

  @State() selectedStart: Date = null;
  @State() selectedEnd: Date = null;
  @State() conditionName: string = '';
  @State() isOngoing: boolean = false;
  @State() hoveredDate: Date = null;

  @State() startError: string = null;
  @State() endError: string = null;
  @State() nameError: string = null;

  @State() currentViewMonth: number = TODAY.getMonth();
  @State() currentViewYear: number = TODAY.getFullYear();

  componentWillLoad() {
    if (this.startDate) {
      this.selectedStart = this.startDate;
    }
  }

  private getDaysInMonth = (year: number, month: number) => {
    return new Date(year, month + 1, 0).getDate();
  };

  private getFirstDayOfMonth = (year: number, month: number) => {
    const firstDay = new Date(year, month, 1).getDay();
    return firstDay === 0 ? 6 : firstDay - 1;
  };

  private handleSelectDay = (day: number) => {
    const date: Date = new Date(this.currentViewYear, this.currentViewMonth, day);

    if (this.selectedStart && this.selectedEnd) {
      this.selectedStart = date;
      this.selectedEnd = null;
    } else if (this.selectedStart && date < this.selectedStart) {
      this.selectedStart = date;
    } else if (this.selectedStart && date >= this.selectedStart) {
      this.selectedEnd = date;
      this.isOngoing = false;
    } else {
      this.selectedStart = date;
    }
  };

  private handleMouseOver = (day: number) => {
    this.hoveredDate = new Date(this.currentViewYear, this.currentViewMonth, day);
  };

  private handleMouseOut = () => {
    this.hoveredDate = null;
  };

  private renderCalendar = () => {
    const year = this.currentViewYear;
    const month = this.currentViewMonth;
    const today = new Date();
    today.setHours(0, 0, 0, 0);

    const daysInMonth = this.getDaysInMonth(year, month);
    const firstDayOfMonth = this.getFirstDayOfMonth(year, month);

    const prevMonthDays = [];
    const currentMonthDays = [];
    const nextMonthDays = [];

    const daysInPrevMonth = this.getDaysInMonth(year, month - 1);
    for (let i = firstDayOfMonth - 1; i >= 0; i--) {
      prevMonthDays.unshift(
        <div class="px-3 py-2 text-center text-sm text-gray-400">{daysInPrevMonth - i}</div>,
      );
    }

    for (let i = 1; i <= daysInMonth; i++) {
      const currentDate = new Date(year, month, i);
      currentDate.setHours(0, 0, 0, 0);

      const isSelectedStart: boolean =
        this.selectedStart &&
        this.selectedStart.getDate() === i &&
        this.selectedStart.getMonth() === month &&
        this.selectedStart.getFullYear() === year;
      const isSelectedEnd: boolean =
        this.selectedEnd &&
        this.selectedEnd.getDate() === i &&
        this.selectedEnd.getMonth() === month &&
        this.selectedEnd.getFullYear() === year;
      const isInRange: boolean =
        this.selectedStart &&
        this.selectedEnd &&
        currentDate >= this.selectedStart &&
        currentDate <= this.selectedEnd;
      const isHoveredInRange: boolean =
        this.selectedStart &&
        this.hoveredDate &&
        currentDate >= this.selectedStart &&
        currentDate <= this.hoveredDate;

      const isPastDate: boolean = currentDate > today;

      currentMonthDays.push(
        <div
          class={`rounded-md px-3 py-2 text-center text-sm ${
            isSelectedStart || isSelectedEnd
              ? 'bg-[#7357be] text-white'
              : this.isOngoing && !isPastDate && this.selectedStart < currentDate
                ? 'bg-[#d8c7ed]'
                : isInRange
                  ? 'bg-[#d8c7ed] text-black'
                  : isHoveredInRange
                    ? 'cursor-pointer bg-gray-200'
                    : isPastDate
                      ? 'cursor-not-allowed text-gray-400'
                      : 'cursor-pointer hover:bg-gray-200'
          } `}
          onClick={() => !isPastDate && this.handleSelectDay(i)}
          onMouseOver={() => !isPastDate && this.handleMouseOver(i)}
          onMouseOut={this.handleMouseOut}
        >
          {i}
        </div>,
      );
    }

    const totalCells = 42;
    const remainingCells = totalCells - (prevMonthDays.length + currentMonthDays.length);
    for (let i = 1; i <= remainingCells; i++) {
      nextMonthDays.push(<div class="px-3 py-2 text-center text-sm text-gray-400">{i}</div>);
    }

    return [...prevMonthDays, ...currentMonthDays, ...nextMonthDays];
  };

  private getMonthName = () => {
    return MONTHS[this.currentViewMonth];
  };

  private prevMonth = () => {
    if (this.currentViewMonth === 0) {
      this.currentViewMonth = 11;
      this.currentViewYear--;
    } else {
      this.currentViewMonth--;
    }
  };

  private nextMonth = () => {
    if (this.currentViewMonth === 11) {
      this.currentViewMonth = 0;
      this.currentViewYear++;
    } else {
      this.currentViewMonth++;
    }
  };

  private handleYearChange = (event: Event) => {
    this.currentViewYear = parseInt((event.target as HTMLSelectElement).value);
  };

  private handleRegisterCondition = async () => {
    this.startError = null;
    this.endError = null;
    this.nameError = null;

    if (!this.selectedStart) {
      this.startError = 'Condition start is required';
    } else if (!this.selectedEnd && !this.isOngoing) {
      this.endError = 'Condition can either end or be ongoing';
    } else if (this.conditionName === '') {
      this.nameError = 'Condition name is required';
    }

    if (this.startError || this.endError || this.nameError) {
      return;
    }

    try {
      const request: NewCondition = {
        id: '',
        name: this.conditionName,
        patientId: this.user.id,
        start: this.selectedStart,
        end: this.selectedEnd || undefined,
      };
      await this.api.conditions.createPatientCondition({ newCondition: request });
      this.navigate('./homepage');
    } catch (err) {
      toastService.showError(err.message);
    }
  };

  private resetSelection = () => {
    this.selectedStart = null;
    this.selectedEnd = null;
    this.conditionName = '';
    this.isOngoing = false;
  };

  private showDetailsPanel = () => {
    return this.selectedStart !== null && (this.selectedEnd !== null || this.isOngoing);
  };

  private handleConditionNameChange = (event: Event) => {
    this.conditionName = (event.target as HTMLInputElement).value;
  };

  render() {
    const currentYear: number = TODAY.getFullYear();
    const yearOptions: Array<number> = [];
    for (let i = 0; i <= 5; i++) {
      yearOptions.push(currentYear + i);
    }

    const showDetails = this.showDetailsPanel();

    return (
      <div class="flex h-screen w-full flex-1 flex-col overflow-auto">
        <xcastven-xkilian-project-header
          navigate={this.navigate}
          type="registerCondition"
          isDoctor={false}
        />

        {/* Content */}
        <div class="mx-auto flex w-full flex-1 flex-col md:flex-row">
          {/* Left panel - Calendar */}
          <div
            class={`flex flex-col items-center justify-center bg-gray-300 p-6 transition-all duration-600 ease-in-out ${showDetails ? 'w-full md:w-1/2' : 'w-full'}`}
          >
            <div class="mb-6 w-full max-w-md rounded-lg bg-white p-4 shadow-md">
              <div class="mb-4 flex items-center justify-between">
                <md-icon-button onClick={() => this.prevMonth()}>
                  <md-icon>chevron_left</md-icon>
                </md-icon-button>
                <div class="flex items-center text-center">
                  <span class="font-medium">{this.getMonthName()}</span>
                  <span class="mx-1">,</span>
                  <select
                    class="border-none bg-transparent font-medium"
                    onChange={(e: Event) => this.handleYearChange(e)}
                  >
                    {yearOptions.map(year => (
                      <option value={year.toString()} selected={year === this.currentViewYear}>
                        {year}
                      </option>
                    ))}
                  </select>
                </div>
                <md-icon-button onClick={() => this.nextMonth()}>
                  <md-icon>chevron_right</md-icon>
                </md-icon-button>
              </div>

              <div class="mb-3 grid grid-cols-7 gap-1">
                {DAYS_OF_WEEK.map(day => (
                  <div class="px-3 py-2 text-center text-sm font-medium text-gray-600">
                    {day.short}
                  </div>
                ))}
                {this.renderCalendar()}
              </div>
            </div>
            {this.selectedStart && (
              <div class="flex max-w-md min-w-md animate-[slideInFromBottom_0.5s_ease-out] flex-row items-center justify-center gap-x-3">
                <label htmlFor="switch" class="font-medium text-gray-600">
                  Condition is still ongoing
                </label>
                <md-switch
                  id="switch"
                  checked={this.isOngoing}
                  onChange={() => {
                    if (this.selectedEnd) this.selectedEnd = null;
                    this.isOngoing = !this.isOngoing;
                  }}
                />
              </div>
            )}
          </div>

          {/* Right panel - Details */}
          {showDetails && (
            <div
              class={`relative m-auto flex h-full w-full max-w-lg flex-1 transform animate-[slideInFromBottom_0.5s_ease-out] flex-col justify-center p-6 opacity-100 transition-all duration-500 ease-in-out md:w-1/2 md:animate-[slideInFromRight_0.5s_ease-out]`}
            >
              <md-icon-button class="absolute top-5 left-5" onClick={this.resetSelection}>
                <md-icon>arrow_back</md-icon>
              </md-icon-button>

              <div class="mb-6 w-full max-w-md rounded-md bg-gray-200 px-4 py-3">
                <div class="mb-1 flex w-full flex-row items-center justify-between">
                  <div class="flex flex-row items-center gap-x-2 text-gray-500">
                    <md-icon style={{ fontSize: '16px' }}>line_start_circle</md-icon>
                    From
                  </div>
                  <span class="font-medium text-gray-600">{formatDate(this.selectedStart)}</span>
                </div>

                {this.selectedEnd && (
                  <>
                    <div class="mb-1 flex w-full flex-row items-center justify-between">
                      <div class="flex flex-row items-center gap-x-2 text-gray-500">
                        <md-icon style={{ fontSize: '16px' }}>line_end_circle</md-icon>
                        To
                      </div>
                      <span class="font-medium text-gray-600">{formatDate(this.selectedEnd)}</span>
                    </div>

                    <div class="mb-1 flex w-full flex-row items-center justify-between">
                      <div class="flex flex-row items-center gap-x-2 text-gray-500">
                        <md-icon style={{ fontSize: '16px' }}>timer</md-icon>
                        Duration
                      </div>
                      <span class="font-medium text-gray-600">
                        {formatDateDelta(this.selectedStart, this.selectedEnd)}
                      </span>
                    </div>
                  </>
                )}

                <div class="flex w-full flex-row items-center justify-between">
                  <div class="flex flex-row items-center gap-x-2 text-gray-500">
                    <md-icon style={{ fontSize: '16px' }}>
                      {this.selectedEnd ? 'check_circle' : 'pending'}
                    </md-icon>
                    Status
                  </div>
                  <span class="font-medium text-gray-600">
                    {this.selectedEnd ? 'Gone' : 'Ongoing'}
                  </span>
                </div>
              </div>

              <div class="mb-6">
                <md-filled-text-field
                  label="Condition name"
                  class="w-full"
                  value={this.conditionName}
                  onInput={(e: Event) => this.handleConditionNameChange(e)}
                />
              </div>

              {this.startError ? (
                <div class="mb-6 w-full text-center text-sm text-red-500">{this.startError}</div>
              ) : this.endError ? (
                <div class="mb-6 w-full text-center text-sm text-red-500">{this.endError}</div>
              ) : (
                this.nameError && (
                  <div class="mb-6 w-full text-center text-sm text-red-500">{this.startError}</div>
                )
              )}

              <md-filled-button
                class="w-full rounded-full bg-[#7357be] p-3 text-white hover:bg-[#5c469d]"
                onClick={() => this.handleRegisterCondition()}
              >
                Register this condition
              </md-filled-button>
            </div>
          )}
        </div>
      </div>
    );
  }
}
