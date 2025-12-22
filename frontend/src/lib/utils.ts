import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Utility function to merge Tailwind CSS classes efficiently
 *
 * This function combines clsx for conditional class names and
 * tailwind-merge to intelligently merge Tailwind utility classes,
 * removing conflicts and duplicates.
 *
 * @example
 * ```ts
 * cn("px-4 py-2", "px-6") // => "py-2 px-6"
 * cn("text-red-500", condition && "text-blue-500") // => "text-blue-500" if condition is true
 * ```
 *
 * @param inputs - Class names to merge (strings, objects, arrays, etc.)
 * @returns Merged class name string
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
