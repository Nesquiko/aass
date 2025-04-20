import { Component, h, Prop, State, Method, Element } from '@stencil/core';

@Component({
  tag: 'xcastven-xkilian-project-toast',
  shadow: false,
})
export class Toast {
  @Element() el: HTMLElement;
  @Prop() duration: number = 3_000;
  @State() isVisible: boolean = false;
  @State() message: string = '';
  @State() type: 'success' | 'error' | 'info' = 'info';

  private timeoutId: number;

  @Method()
  async show(message: string, type: 'success' | 'error' | 'info' = 'info', duration?: number) {
    this.message = message;
    this.type = type;
    this.isVisible = true;

    clearTimeout(this.timeoutId);
    this.timeoutId = window.setTimeout(() => {
      this.isVisible = false;
    }, duration || this.duration);
  }

  render() {
    const iconNames = {
      success: 'check_circle',
      error: 'error',
      info: 'info',
    };

    return (
      <div
        class={`fixed bottom-4 left-4 z-50 flex transform flex-row items-center justify-center gap-x-2 rounded-md border-2 border-[#d8c7ed] bg-white px-4 py-3 transition-all duration-300 ${
          this.isVisible
            ? 'translate-y-0 opacity-100'
            : 'pointer-events-none translate-y-8 opacity-0'
        }`}
      >
        <md-icon>{iconNames[this.type]}</md-icon>
        <span>{this.message}</span>
      </div>
    );
  }
}
