import type { ClassValue } from "clsx"
import { clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/** Converts an ISO date string to the format expected by datetime-local inputs (YYYY-MM-DDTHH:MM). */
export function toDatetimeLocalValue(iso: string): string {
  return new Date(iso).toISOString().slice(0, 16)
}
