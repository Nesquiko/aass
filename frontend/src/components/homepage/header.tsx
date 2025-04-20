import { Navigate } from '../../utils/types';
import { MONTHS, TODAY } from '../../utils/utils';
import { Component, h, Prop, State } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-header',
  shadow: false,
})
export class Header {
  @Prop() navigate: Navigate;
  @Prop() isDoctor: boolean;
  @Prop() type: 'calendar' | 'account' | 'scheduleAppointment' | 'registerCondition';
  @Prop() currentViewMonth?: number;
  @Prop() currentViewYear?: number;
  @Prop() handlePreviousMonth?: () => void;
  @Prop() handleNextMonth?: () => void;
  @Prop() handleYearChange?: (event: Event) => void;

  @State() isMenuOpen: boolean = false;

  private handleToggleMenu = () => {
    this.isMenuOpen = !this.isMenuOpen;
  };

  private getMonthName = () => {
    return MONTHS[this.currentViewMonth];
  };

  private getTitle = () => {
    switch (this.type) {
      case 'account':
        return 'Account';
      case 'scheduleAppointment':
        return 'Schedule an appointment';
      case 'registerCondition':
        return 'Register a condition';
      default:
        return '';
    }
  };

  render() {
    const currentYear: number = TODAY.getFullYear();
    const yearOptions: Array<number> = [];
    for (let i: number = -5; i <= 5; i++) {
      yearOptions.push(currentYear + i);
    }

    return (
      <div class="z-10 flex h-[48px] items-center bg-gray-800 px-3 py-1 text-white">
        <md-icon-button id="menu-button" class="mr-2" onClick={this.handleToggleMenu}>
          <md-icon class="text-white">menu</md-icon>
        </md-icon-button>

        {this.isMenuOpen && (
          <div class="fixed inset-0 z-99 bg-black/50" onClick={() => (this.isMenuOpen = false)} />
        )}

        <xcastven-xkilian-project-menu
          navigate={this.navigate}
          isDoctor={this.isDoctor}
          isMenuOpen={this.isMenuOpen}
          handleResetMenu={() => (this.isMenuOpen = false)}
        />

        {this.type === 'calendar' ? (
          <div class="flex flex-1 items-center justify-center gap-x-10">
            <md-icon-button onClick={this.handlePreviousMonth} title="Previous month">
              <md-icon class="text-white">chevron_left</md-icon>
            </md-icon-button>
            <div class="flex w-48 items-center justify-center text-center">
              <span class="font-medium">{this.getMonthName()}</span>
              <span>,</span>
              <select
                class="border-none bg-transparent font-medium"
                onChange={(e: Event) => this.handleYearChange(e)}
              >
                {yearOptions.map(year => (
                  <option
                    value={year.toString()}
                    selected={year === this.currentViewYear}
                    class="text-black hover:text-white"
                  >
                    {year}
                  </option>
                ))}
              </select>
            </div>
            <md-icon-button onClick={this.handleNextMonth} title="Next month">
              <md-icon class="text-white">chevron_right</md-icon>
            </md-icon-button>
          </div>
        ) : (
          <h1 class="w-full text-center text-xl font-medium text-white">{this.getTitle()}</h1>
        )}

        <md-icon-button onClick={() => this.navigate('./account')}>
          <md-icon class="text-white">account_circle</md-icon>
        </md-icon-button>
      </div>
    );
  }
}
