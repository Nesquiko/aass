import { PrescriptionDisplay } from '../../api/generated';
import { formatDate, TODAY } from '../../utils/utils';
import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-prescriptions-list',
  shadow: false,
})
export class PrescriptionsList {
  @Prop() prescriptions: Array<PrescriptionDisplay>;
  @Prop() handleSelectPrescription: (prescription: PrescriptionDisplay) => void;
  @Prop() selectedDate: Date;
  @Prop() setSelectedDate: (date: Date) => void;

  render() {
    return (
      <div class="w-full overflow-hidden rounded-lg bg-white shadow-sm">
        {this.prescriptions.length ? (
          this.prescriptions.map((prescription: PrescriptionDisplay, index: number) => {
            return (
              <div key={prescription.id}>
                <div
                  class={`flex w-full cursor-pointer flex-col justify-center border-2 border-transparent px-4 py-2 hover:border-[#9d83c6] ${index % 2 === 0 ? 'bg-gray-200' : 'bg-white'} ${index === 0 && 'rounded-t-lg'} `}
                  onClick={() => this.handleSelectPrescription(prescription)}
                >
                  <div class="flex flex-row items-center justify-between">
                    <div class="font-medium text-[#7357be]">{prescription.name}</div>
                    <div class="flex items-center">
                      {TODAY.getTime() > prescription.end.getTime() ? (
                        <md-icon class="text-gray-500" style={{ fontSize: '16px' }}>
                          check_circle
                        </md-icon>
                      ) : (
                        <md-icon class="text-gray-500" style={{ fontSize: '16px' }}>
                          pending
                        </md-icon>
                      )}
                    </div>
                  </div>
                  <div class="flex flex-row items-center justify-between">
                    <span class="text-sm text-gray-400">
                      From:
                      <span class="ml-2 text-sm font-medium text-gray-600">
                        {formatDate(prescription.start)}
                      </span>
                    </span>
                    {prescription.end && (
                      <span class="text-sm text-gray-400">
                        To:
                        <span class="ml-2 text-sm font-medium text-gray-600">
                          {formatDate(prescription.end)}
                        </span>
                      </span>
                    )}
                  </div>
                </div>
              </div>
            );
          })
        ) : (
          <div
            class={`flex h-16 w-full flex-col justify-center border-2 border-transparent bg-gray-200 px-4 py-2 text-center text-sm font-medium text-gray-600`}
          >
            No prescriptions for this date
          </div>
        )}

        <div class="flex h-12 w-full flex-row items-center justify-center" />
      </div>
    );
  }
}
