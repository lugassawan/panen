export type ToastVariant = "success" | "error" | "warning" | "info";

export interface Toast {
  id: string;
  message: string;
  variant: ToastVariant;
}

const MAX_TOASTS = 3;
const DEFAULT_DURATION = 4000;

function createToastStore() {
  let toasts = $state<Toast[]>([]);

  return {
    get toasts() {
      return toasts;
    },
    add(message: string, variant: ToastVariant, duration = DEFAULT_DURATION) {
      const id = `toast-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;
      const toast: Toast = { id, message, variant };

      toasts = [...toasts, toast];
      if (toasts.length > MAX_TOASTS) {
        toasts = toasts.slice(toasts.length - MAX_TOASTS);
      }

      setTimeout(() => {
        toasts = toasts.filter((t) => t.id !== id);
      }, duration);
    },
    dismiss(id: string) {
      toasts = toasts.filter((t) => t.id !== id);
    },
    clear() {
      toasts = [];
    },
  };
}

export const toastStore = createToastStore();
