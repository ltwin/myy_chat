/**
 * Toast notification hook
 *
 * This is a placeholder hook that will be replaced by shadcn/ui's
 * toast implementation when you run: npx shadcn-ui@latest add toast
 *
 * For now, it provides a basic type structure for toast notifications.
 */

/**
 * Toast notification options
 */
export interface ToastOptions {
  title?: string;
  description?: string;
  variant?: "default" | "destructive";
  duration?: number;
}

/**
 * Toast hook return type
 */
export interface UseToastReturn {
  toast: (options: ToastOptions) => void;
  dismiss: (toastId?: string) => void;
}

/**
 * Placeholder toast hook
 *
 * Install shadcn/ui toast component to use this hook:
 * ```bash
 * npx shadcn-ui@latest add toast
 * ```
 *
 * @returns Toast utilities
 */
export function useToast(): UseToastReturn {
  return {
    toast: (options: ToastOptions) => {
      // Placeholder implementation
      // Will be replaced by shadcn/ui toast
      console.log("Toast:", options);
    },
    dismiss: (toastId?: string) => {
      console.log("Dismiss toast:", toastId);
    },
  };
}
