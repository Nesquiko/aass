import { ConditionOrderColors, PrescriptionOrderColors } from '../../utils/utils';
import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-legend',
  shadow: false,
})
export class Legend {
  @Prop() handleResetSelection: () => void;

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
            Calendar Legend
          </h2>
        </div>

        <div class="flex flex-col gap-y-1 rounded-md bg-gray-200 p-2">
          <div
            slot="headline"
            class="z-98 flex w-full flex-row items-center justify-between gap-x-2 rounded-md border-2 border-transparent px-2 py-1 text-sm hover:border-[#9d83c6]"
          >
            <span class="font-medium text-gray-600">Schedule-able day</span>
            <div class="h-6 w-6 rounded-full border-2 border-white bg-[#d8c7ed]" />
          </div>

          <div
            slot="headline"
            class="z-98 flex w-full flex-row items-center justify-between gap-x-2 rounded-md border-2 border-transparent px-2 py-1 text-sm hover:border-[#9d83c6]"
          >
            <span class="font-medium text-gray-600">Non schedule-able day</span>
            <div class="h-6 w-6 rounded-full border-2 border-white bg-white" />
          </div>

          <div
            slot="headline"
            class="z-98 flex w-full flex-row items-center justify-between gap-x-2 rounded-md border-2 border-transparent px-2 py-1 text-sm hover:border-[#9d83c6]"
          >
            <span class="font-medium text-gray-600">Other month's day</span>
            <div class="h-6 w-6 rounded-full border-2 border-white bg-gray-200" />
          </div>

          <div
            slot="headline"
            class="z-98 flex w-full flex-row items-center justify-between gap-x-2 rounded-md border-2 border-transparent px-2 py-1 text-sm hover:border-[#9d83c6]"
          >
            <span class="font-medium text-gray-600">Scheduled appointment</span>
            <div class="h-6 w-6 rounded-full border-2 border-white bg-[#7357be]" />
          </div>

          <div
            slot="headline"
            class="z-98 flex w-full flex-row items-center justify-between gap-x-2 rounded-md border-2 border-transparent px-2 py-1 text-sm hover:border-[#9d83c6]"
          >
            <span class="font-medium text-gray-600">Requested appointment</span>
            <div class="h-6 w-6 rounded-full border-2 border-white bg-[#9d83c6]" />
          </div>

          <div
            slot="headline"
            class="z-98 flex w-full flex-row items-center justify-between gap-x-2 rounded-md border-2 border-transparent px-2 py-1 text-sm hover:border-[#9d83c6]"
          >
            <span class="font-medium text-gray-600">Completed appointment</span>
            <div class="h-6 w-6 rounded-full border-2 border-white bg-[#2E8B57]" />
          </div>

          <div
            slot="headline"
            class="z-98 flex w-full flex-row items-center justify-between gap-x-2 rounded-md border-2 border-transparent px-2 py-1 text-sm hover:border-[#9d83c6]"
          >
            <span class="font-medium text-gray-600">Cancelled appointment</span>
            <div class="h-6 w-6 rounded-full border-2 border-white bg-[#F08080]" />
          </div>

          <div
            slot="headline"
            class="z-98 flex w-full flex-row items-center justify-between gap-x-2 rounded-md border-2 border-transparent px-2 py-1 text-sm hover:border-[#9d83c6]"
          >
            <span class="font-medium text-gray-600">Denied appointment</span>
            <div class="h-6 w-6 rounded-full border-2 border-white bg-[#4f4f4f]" />
          </div>

          <div
            slot="headline"
            class="z-98 flex w-full flex-col items-center justify-between gap-y-2 rounded-md border-2 border-transparent px-2 py-1 text-sm hover:border-[#9d83c6]"
          >
            <span class="w-full text-left font-medium text-gray-600">Conditions</span>
            <div class="flex h-8 w-full overflow-hidden rounded-full border-2 border-white">
              {ConditionOrderColors.map((color: string) => (
                <div
                  class="flex-grow"
                  style={{
                    backgroundColor: color,
                  }}
                ></div>
              ))}
            </div>
          </div>

          <div
            slot="headline"
            class="z-98 flex w-full flex-col items-center justify-between gap-y-2 rounded-md border-2 border-transparent px-2 py-1 text-sm hover:border-[#9d83c6]"
          >
            <span class="w-full text-left font-medium text-gray-600">Prescriptions</span>
            <div class="flex h-8 w-full overflow-hidden rounded-full border-2 border-white">
              {PrescriptionOrderColors.map((color: string) => (
                <div
                  class="flex-grow"
                  style={{
                    backgroundColor: color,
                  }}
                ></div>
              ))}
            </div>
          </div>
        </div>
      </div>
    );
  }
}
