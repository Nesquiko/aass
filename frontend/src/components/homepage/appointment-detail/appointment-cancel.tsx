import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-appointment-cancel',
  shadow: false,
})
export class AppointmentCancel {
  @Prop() cancelling: boolean;
  @Prop() setCancelling: (cancelling: boolean) => void;
  @Prop() cancellingAppointmentReason: string;
  @Prop() handleCancellingAppointmentReasonChange: (event: Event) => void;
  @Prop() handleCancel: () => void;

  render() {
    return (
      <div class="my-6 w-full max-w-md rounded-md bg-gray-200 px-4 py-3">
        <h4 class="mb-2 text-sm font-medium text-[#7357be]">Cancel appointment</h4>
        <div class="mb-3 flex w-full flex-col gap-y-3">
          <md-outlined-text-field
            required={true}
            label="Cancellation reason"
            class="w-full"
            value={this.cancellingAppointmentReason}
            onInput={this.handleCancellingAppointmentReasonChange}
          />
        </div>

        <div class="flex flex-row items-center justify-between gap-x-2">
          <md-filled-button
            class={`w-1/2 rounded-full bg-[#7357be]`}
            onClick={this.handleCancel}
            disabled={this.cancellingAppointmentReason === ''}
          >
            Confirm cancel
          </md-filled-button>
          <md-outlined-button class="w-1/2 rounded-full" onClick={() => this.setCancelling(false)}>
            Back
          </md-outlined-button>
        </div>
      </div>
    );
  }
}
