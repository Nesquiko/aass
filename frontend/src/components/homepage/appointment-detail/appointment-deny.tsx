import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-appointment-deny',
  shadow: false,
})
export class AppointmentDeny {
  @Prop() denying: boolean;
  @Prop() setDenying: (denying: boolean) => void;
  @Prop() denyingAppointmentReason: string;
  @Prop() handleDenyingAppointmentReasonChange: (event: Event) => void;
  @Prop() handleDeny: () => void;
  @Prop() disabled: boolean;

  render() {
    return (
      <div class="my-6 w-full max-w-md rounded-md bg-gray-200 px-4 py-3">
        <h4 class="mb-2 text-sm font-medium text-[#7357be]">Deny appointment request</h4>
        <div class="mb-3 flex w-full flex-col gap-y-3">
          <md-outlined-text-field
            required={true}
            label="Reason for denying"
            class="w-full"
            value={this.denyingAppointmentReason}
            onInput={this.handleDenyingAppointmentReasonChange}
          />
        </div>

        <div class="flex flex-row items-center justify-between gap-x-2">
          <md-filled-button
            class={`w-1/2 rounded-full bg-[#7357be]`}
            onClick={this.handleDeny}
            disabled={this.disabled}
          >
            Confirm deny
          </md-filled-button>
          <md-outlined-button class="w-1/2 rounded-full" onClick={() => this.setDenying(false)}>
            Back
          </md-outlined-button>
        </div>
      </div>
    );
  }
}
