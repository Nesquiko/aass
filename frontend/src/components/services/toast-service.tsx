interface ToastElement extends HTMLElement {
  show(message: string, type: 'success' | 'error' | 'info', duration?: number): void;
}

export class ToastService {
  private static instance: ToastService;
  private toastElement: ToastElement | null = null;

  private constructor() {}

  public static getInstance(): ToastService {
    if (!ToastService.instance) {
      ToastService.instance = new ToastService();
    }
    return ToastService.instance;
  }

  public initialize() {
    if (!this.toastElement) {
      this.toastElement = document.querySelector('xcastven-xkilian-project-toast') as ToastElement;

      if (!this.toastElement) {
        this.toastElement = document.createElement(
          'xcastven-xkilian-project-toast',
        ) as ToastElement;
        document.body.appendChild(this.toastElement);
      }
    }
  }

  public showSuccess(message: string, duration?: number) {
    this.initialize();
    if (this.toastElement) {
      this.toastElement.show(message, 'success', duration);
    }
  }

  public showError(message: string, duration?: number) {
    this.initialize();
    if (this.toastElement) {
      this.toastElement.show(message, 'error', duration);
    }
  }

  public showInfo(message: string, duration?: number) {
    this.initialize();
    if (this.toastElement) {
      this.toastElement.show(message, 'info', duration);
    }
  }
}

export const toastService: ToastService = ToastService.getInstance();
