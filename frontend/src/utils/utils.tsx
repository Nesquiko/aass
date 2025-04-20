import { AppointmentType, SpecializationEnum } from '../api/generated';
import { h } from '@stencil/core';

export const DAYS_OF_WEEK: Array<{ short: string; long: string }> = [
  { short: 'Mo', long: 'Monday' },
  { short: 'Tu', long: 'Tuesday' },
  { short: 'We', long: 'Wednesday' },
  { short: 'Th', long: 'Thursday' },
  { short: 'Fr', long: 'Friday' },
  { short: 'Sa', long: 'Saturday' },
  { short: 'Su', long: 'Sunday' },
];

export const MONTHS: Array<string> = [
  'January',
  'February',
  'March',
  'April',
  'May',
  'June',
  'July',
  'August',
  'September',
  'October',
  'November',
  'December',
];

export const TODAY: Date = new Date();

export const formatDate = (date: Date) => {
  if (!date) return '';
  return `${MONTHS[date.getMonth()]} ${date.getDate()}, ${date.getFullYear()}`;
};

export const formatTime = (date: Date): string => {
  if (!date) return '';

  const hours: string = date.getHours().toString();
  const minutes: string = date.getMinutes().toString().padStart(2, '0');

  return `${hours}:${minutes}`;
};

export const getDateAndTimeTitle = (dateTime: Date, className?: string) => {
  return (
    <h2 class={'text-center text-2xl' + className && ' ' + className}>
      {dateTime && <span class={`font-medium text-[#7357be]`}>{formatDate(dateTime)}</span>}
      {dateTime.getHours() !== null && (
        <span class="text-gray-600">
          {' '}
          at <span class={`font-medium text-[#7357be]`}>{formatTime(dateTime)}</span>
        </span>
      )}
    </h2>
  );
};

export const getSelectedDateTimeObject = (date: Date, time: string): Date => {
  const [hours, minutes] = time.split(':').map(Number);
  date.setHours(hours);
  date.setMinutes(minutes);
  return date;
};

export const formatDateDelta = (startDate: Date, endDate: Date = new Date()) => {
  if (isNaN(startDate.getTime()) || isNaN(endDate.getTime())) {
    return 'Invalid date';
  }

  if (endDate < startDate) {
    return formatDateDelta(endDate, startDate);
  }

  let years: number = endDate.getFullYear() - startDate.getFullYear();
  let months: number = endDate.getMonth() - startDate.getMonth();
  let days: number = endDate.getDate() - startDate.getDate();

  if (days < 0) {
    const previousMonth = new Date(endDate.getFullYear(), endDate.getMonth(), 0);
    days += previousMonth.getDate();
    months--;
  }

  if (months < 0) {
    months += 12;
    years--;
  }

  if (years === 0 && months === 0 && days === 0) {
    return 'Less than a day';
  }

  const parts: string[] = [];

  if (years > 0) {
    parts.push(`${years} ${years === 1 ? 'year' : 'years'}`);
  }

  if (months > 0) {
    parts.push(`${months} ${months === 1 ? 'month' : 'months'}`);
  }

  if (days > 0 || (years === 0 && months === 0)) {
    parts.push(`${days + 1} ${days === 1 ? 'day' : 'days'}`);
  }

  if (parts.length === 1) {
    return parts[0];
  } else if (parts.length === 2) {
    return `${parts[0]} and ${parts[1]}`;
  } else {
    return `${parts[0]}, ${parts[1]} and ${parts[2]}`;
  }
};

export const AppointmentStatusColor = {
  requested: {
    background: '#9d83c6',
    foreground: '#000000',
  },
  scheduled: {
    background: '#7357be',
    foreground: '#ffffff',
  },
  completed: {
    background: '#2E8B57',
    foreground: '#ffffff',
  },
  cancelled: {
    background: '#F08080',
    foreground: '#ffffff',
  },
  denied: {
    background: '#4f4f4f',
    foreground: '#000000',
  },
};

export const ConditionOrderColors: Array<string> = [
  '#1976D2',
  '#7B1FA2',
  '#388E3C',
  '#D32F2F',
  '#F57C00',
  '#512DA8',
  '#00796B',
  '#795548',
  '#0097A7',
  '#FBC02D',
];

export const PrescriptionOrderColors: Array<string> = ['#FF8080', '#FF0000', '#800000', '#1A0000'];

export const formatSpecialization = (specialization: SpecializationEnum) => {
  const withSpaces: string = specialization.replace('_', ' ');
  return withSpaces[0].toUpperCase() + withSpaces.slice(1);
};

export const formatAppointmentType = (appointmentType: AppointmentType) => {
  const withSpaces: string = appointmentType.replace('_', ' ');
  return withSpaces[0].toUpperCase() + withSpaces.slice(1);
};

export const days: Array<number> = Array.from({ length: 31 }, (_, i) => i + 1);
export const months: Array<{ value: number; name: string }> = [
  { value: 0, name: 'January' },
  { value: 1, name: 'February' },
  { value: 2, name: 'March' },
  { value: 3, name: 'April' },
  { value: 4, name: 'May' },
  { value: 5, name: 'June' },
  { value: 6, name: 'July' },
  { value: 7, name: 'August' },
  { value: 8, name: 'September' },
  { value: 9, name: 'October' },
  { value: 10, name: 'November' },
  { value: 11, name: 'December' },
];
export const currentYear: number = new Date().getFullYear();
export const years: Array<number> = Array.from({ length: 11 }, (_, i) => currentYear - 5 + i); // Range: current year +/- 5 years

export const updateDatePart = (
  currentDate: Date | null,
  part: 'day' | 'month' | 'year',
  value: number,
): Date => {
  const date: Date = currentDate ? new Date(currentDate) : new Date();
  if (isNaN(value)) return date;

  let day: number = date.getDate();
  let month: number = date.getMonth();
  let year: number = date.getFullYear();

  switch (part) {
    case 'day':
      day = value;
      break;
    case 'month':
      month = value;
      break;
    case 'year':
      year = value;
      break;
  }

  return new Date(year, month, day);
};

export const getDatePart = (date: Date | null, part: 'day' | 'month' | 'year'): number | '' => {
  if (!date || isNaN(date.getTime())) return '';
  switch (part) {
    case 'day':
      return date.getDate();
    case 'month':
      return date.getMonth();
    case 'year':
      return date.getFullYear();
  }
};

export const renderDateSelects = (
  dateType: 'start' | 'end',
  label: string,
  dateValue: Date | null,
  changeHandler: (type: 'start' | 'end', part: 'day' | 'month' | 'year', event: Event) => void,
) => {
  return (
    <div class="flex w-full max-w-md flex-col gap-y-1">
      <label class="text-sm font-medium text-gray-600">{label}</label>
      <div class="flex w-full max-w-md flex-row justify-between gap-x-3">
        <md-outlined-select
          required={true}
          label="Day"
          class="min-w-0 flex-1"
          value={getDatePart(dateValue, 'day')}
          onInput={(e: Event) => changeHandler(dateType, 'day', e)}
        >
          {days.map((day: number) => (
            <md-select-option value={day.toString()} key={`${dateType}-day-${day}`}>
              <div slot="headline">{day}</div>
            </md-select-option>
          ))}
        </md-outlined-select>

        <md-outlined-select
          required={true}
          label="Month"
          class="min-w-0 flex-1"
          value={getDatePart(dateValue, 'month')}
          onInput={(e: Event) => changeHandler(dateType, 'month', e)}
        >
          {months.map((month: { value: number; name: string }) => (
            <md-select-option
              value={month.value.toString()}
              key={`${dateType}-month-${month.value}`}
            >
              <div slot="headline">{month.name}</div>
            </md-select-option>
          ))}
        </md-outlined-select>

        <md-outlined-select
          required={true}
          label="Year"
          class="min-w-0 flex-1"
          value={getDatePart(dateValue, 'year')}
          onInput={(e: Event) => changeHandler(dateType, 'year', e)}
        >
          {years.map((year: number) => (
            <md-select-option value={year.toString()} key={`${dateType}-year-${year}`}>
              <div slot="headline">{year}</div>
            </md-select-option>
          ))}
        </md-outlined-select>
      </div>
    </div>
  );
};
