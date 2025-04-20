import { Api } from '../../api/api';
import { Prescription } from '../../api/generated';
import { formatDate, formatDateDelta } from '../../utils/utils';
import { toastService } from '../services/toast-service';
import { Component, Prop, State, h } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-prescription-detail',
  shadow: false,
})
export class PrescriptionDetail {
  @Prop() api: Api;
  @Prop() prescriptionId: string;
  @Prop() handleResetSelection: () => void;

  @State() prescription: Prescription;

  async componentWillLoad() {
    try {
      this.prescription = await this.api.medicalHistory.prescriptionDetail({
        prescriptionId: this.prescriptionId,
      });
    } catch (err) {
      toastService.showError(err.message);
    }
  }

  render() {
    return (
      <div class="w-full max-w-md">
        <div class="relative w-full max-w-md">
          <div class="absolute top-1/2 left-0 -translate-x-0 -translate-y-1/2 transform">
            <md-icon-button onClick={this.handleResetSelection}>
              <md-icon class="text-gray-600">arrow_forward</md-icon>
            </md-icon-button>
          </div>

          <h2 class="mb-6 w-full text-center text-xl font-medium text-[#7357be]">
            Prescription details
          </h2>
        </div>

        <div class="mb-6 w-full max-w-md rounded-md bg-gray-200 px-4 py-3">
          <div class="mb-1 flex w-full flex-row items-center justify-between">
            <div class="flex flex-row items-center gap-x-2 text-gray-500">
              <md-icon style={{ fontSize: '16px' }}>medication</md-icon>
              Name
            </div>
            <span class="font-medium text-gray-600">{this.prescription.name}</span>
          </div>

          <div class="mb-1 flex w-full flex-row items-center justify-between">
            <div class="flex flex-row items-center gap-x-2 text-gray-500">
              <md-icon style={{ fontSize: '16px' }}>line_start_circle</md-icon>
              From
            </div>
            <span class="font-medium text-gray-600">{formatDate(this.prescription.start)}</span>
          </div>

          {this.prescription.end && (
            <>
              <div class="mb-1 flex w-full flex-row items-center justify-between">
                <div class="flex flex-row items-center gap-x-2 text-gray-500">
                  <md-icon style={{ fontSize: '16px' }}>line_end_circle</md-icon>
                  To
                </div>
                <span class="font-medium text-gray-600">{formatDate(this.prescription.end)}</span>
              </div>

              <div class="mb-1 flex w-full flex-row items-center justify-between">
                <div class="flex flex-row items-center gap-x-2 text-gray-500">
                  <md-icon style={{ fontSize: '16px' }}>timer</md-icon>
                  Duration
                </div>
                <span class="font-medium text-gray-600">
                  {formatDateDelta(this.prescription.start, this.prescription.end)}
                </span>
              </div>
            </>
          )}
        </div>

        <div class="mb-6 max-h-32 w-full max-w-md overflow-y-auto rounded-md bg-gray-200 px-4 py-3">
          <div class="flex flex-row items-center gap-x-2 text-gray-500">
            <md-icon style={{ fontSize: '16px' }}>description</md-icon>
            Doctor's note
          </div>
          {this.prescription.doctorsNote && (
            <p class="mt-1 ml-1 text-sm font-medium text-wrap text-gray-600">
              {this.prescription.doctorsNote}
            </p>
          )}
        </div>
      </div>
    );
  }
}
