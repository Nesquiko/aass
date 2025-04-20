import { Navigate } from '../../utils/types';
import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-footer',
  shadow: false,
})
export class Footer {
  @Prop() navigate: Navigate;
  @Prop() isDoctor: boolean;
  @Prop() handleToggleLegendMenu: () => void;

  render() {
    return (
      <div class="flex h-[48px] w-full flex-row justify-between bg-[#d8c7ed] px-3 py-1">
        <div class="w-10"></div>

        {!this.isDoctor ? (
          <div class="flex w-full flex-row items-center justify-center gap-x-3">
            <md-text-button
              class="relative w-18 pr-3 text-sm text-[#7357be] sm:w-56"
              onClick={() => this.navigate('./scheduleAppointment')}
              title="Schedule an appointment"
            >
              <md-icon class="absolute top-[10px] left-4" style={{ fontSize: '20px' }}>
                event
              </md-icon>
              <md-icon
                class="absolute top-[10px] left-10 sm:text-transparent"
                style={{ fontSize: '20px' }}
              >
                add
              </md-icon>
              <span class="hidden pl-8 text-center sm:inline">Schedule an appointment</span>
            </md-text-button>

            <md-text-button
              class="relative w-18 pr-3 text-sm text-[#7357be] sm:w-48"
              onClick={() => this.navigate('./registerCondition')}
              title="Register a condition"
            >
              <md-icon class="absolute top-[10px] left-4" style={{ fontSize: '20px' }}>
                coronavirus
              </md-icon>
              <md-icon
                class="absolute top-[10px] left-10 sm:text-transparent"
                style={{ fontSize: '20px' }}
              >
                add
              </md-icon>
              <span class="hidden pl-8 text-center sm:inline">Register a condition</span>
            </md-text-button>
          </div>
        ) : (
          <div class="w-full" />
        )}

        <div class="relative z-98 w-10">
          <md-icon-button
            id="legend-button"
            class="mr-2 text-[#7357be]"
            onClick={this.handleToggleLegendMenu}
            title="Legend"
          >
            <md-icon class="text-[#7357be]">more_horiz</md-icon>
          </md-icon-button>
        </div>
      </div>
    );
  }
}
