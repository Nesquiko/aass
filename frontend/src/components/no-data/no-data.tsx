import { Component, h, Prop } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-no-data',
  shadow: false,
})
export class NoData {
  @Prop() displayTitle: string;
  @Prop() icon: string = 'error';
  @Prop() iconSize: number = 80;

  render() {
    const iconFontSize: string = this.iconSize.toString() + 'px';

    return (
      <div class="flex h-full w-full flex-1 items-center justify-center">
        <div class="flex max-w-sm flex-col items-center justify-center gap-y-3 font-medium text-gray-600">
          <md-icon class="h-full w-full" style={{ fontSize: iconFontSize }}>
            {this.icon}
          </md-icon>
          <h2 class="text-center text-xl">{this.displayTitle}</h2>
        </div>
      </div>
    );
  }
}
