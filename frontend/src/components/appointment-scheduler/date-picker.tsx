import { DAYS_OF_WEEK, MONTHS, TODAY } from '../../utils/utils';
import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-date-picker',
  shadow: false,
})
export class DatePicker {
  @Prop() selectedDate: Date;
  @Prop() selectDate: (day: number) => void;
  @Prop() currentViewMonth: number;
  @Prop() setCurrentViewMonth: (currentViewMonth: number) => void;
  @Prop() currentViewYear: number;
  @Prop() setCurrentViewYear: (currentViewYear: number) => void;

  private getDaysInMonth = (year: number, month: number) => {
    return new Date(year, month + 1, 0).getDate();
  };

  private getFirstDayOfMonth = (year: number, month: number) => {
    const firstDay = new Date(year, month, 1).getDay();
    return firstDay === 0 ? 6 : firstDay - 1;
  };

  private getMonthName = () => {
    return MONTHS[this.currentViewMonth];
  };

  private prevMonth = () => {
    if (this.currentViewMonth === 0) {
      this.setCurrentViewMonth(11);
      this.setCurrentViewYear(this.currentViewYear - 1);
    } else {
      this.setCurrentViewMonth(this.currentViewMonth - 1);
    }
  };

  private nextMonth = () => {
    if (this.currentViewMonth === 11) {
      this.setCurrentViewMonth(0);
      this.setCurrentViewYear(this.currentViewYear + 1);
    } else {
      this.setCurrentViewMonth(this.currentViewMonth + 1);
    }
  };

  private handleYearChange = (event: Event) => {
    this.setCurrentViewYear(parseInt((event.target as HTMLSelectElement).value));
  };

  private renderCalendar = () => {
    const year: number = this.currentViewYear;
    const month: number = this.currentViewMonth;
    const today: Date = new Date();
    today.setHours(0, 0, 0, 0);

    const daysInMonth: number = this.getDaysInMonth(year, month);
    const firstDayOfMonth: number = this.getFirstDayOfMonth(year, month);

    const prevMonthDays = [];
    const currentMonthDays = [];
    const nextMonthDays = [];

    const daysInPrevMonth: number = this.getDaysInMonth(year, month - 1);
    for (let i: number = firstDayOfMonth - 1; i >= 0; i--) {
      prevMonthDays.unshift(
        <div class="px-3 py-2 text-center text-sm text-gray-400">{daysInPrevMonth - i}</div>,
      );
    }

    for (let i = 1; i <= daysInMonth; i++) {
      const currentDate = new Date(year, month, i);
      currentDate.setHours(0, 0, 0, 0);

      const isSelected: boolean =
        this.selectedDate &&
        this.selectedDate.getDate() === i &&
        this.selectedDate.getMonth() === month &&
        this.selectedDate.getFullYear() === year;

      const isPastDate: boolean = currentDate < today;

      currentMonthDays.push(
        <div
          class={`rounded-md px-3 py-2 text-center text-sm ${
            isSelected
              ? 'bg-[#7357be] text-white'
              : isPastDate
                ? 'cursor-not-allowed text-gray-400'
                : 'cursor-pointer hover:bg-gray-200'
          }`}
          onClick={() => !isPastDate && this.selectDate(i)}
        >
          {i}
        </div>,
      );
    }

    const totalCells = 42;
    const remainingCells: number = totalCells - (prevMonthDays.length + currentMonthDays.length);
    for (let i: number = 1; i <= remainingCells; i++) {
      nextMonthDays.push(<div class="px-3 py-2 text-center text-sm text-gray-400">{i}</div>);
    }

    return [...prevMonthDays, ...currentMonthDays, ...nextMonthDays];
  };

  render() {
    const currentYear: number = TODAY.getFullYear();
    const yearOptions: Array<number> = [];
    for (let i: number = 0; i <= 5; i++) {
      yearOptions.push(currentYear + i);
    }

    return (
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
              {yearOptions.map((year: number) => (
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

        <div class="grid grid-cols-7 gap-1">
          {DAYS_OF_WEEK.map(day => (
            <div class="px-3 py-2 text-center text-sm font-medium text-gray-600">{day.short}</div>
          ))}
          {this.renderCalendar()}
        </div>
      </div>
    );
  }
}
