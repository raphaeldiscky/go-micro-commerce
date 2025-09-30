import type { Timestamp } from '@bufbuild/protobuf/wkt'
import {
  differenceInDays,
  differenceInHours,
  differenceInMinutes,
  format,
  formatISO,
  getYear,
  isAfter,
  isValid,
  parseISO,
} from 'date-fns'

/**
 * Format a timestamp to a time string (HH:mm)
 */
export function formatTime(timestamp: string): string {
  try {
    const date = parseISO(timestamp)
    if (!isValid(date)) return ''
    return format(date, 'HH:mm')
  } catch {
    return ''
  }
}

/**
 * Format relative time (e.g., "2 minutes ago", "3 hours ago")
 */
export function formatRelativeTime(timestamp: string): string {
  try {
    const date = parseISO(timestamp)
    if (!isValid(date)) return ''

    const now = new Date()
    const minutes = differenceInMinutes(now, date)
    const hours = differenceInHours(now, date)
    const days = differenceInDays(now, date)

    if (minutes < 1) return 'now'
    if (minutes < 60) return `${minutes}m ago`
    if (hours < 24) return `${hours}h ago`
    if (days < 7) return `${days}d ago`

    return format(date, 'MMM d, yyyy')
  } catch {
    return ''
  }
}

/**
 * Check if a timestamp is after another timestamp or current time
 */
export function isAfterDate(timestamp: string, compareWith?: string): boolean {
  try {
    const date = parseISO(timestamp)
    const compareDate = compareWith ? parseISO(compareWith) : new Date()
    return isValid(date) && isValid(compareDate) && isAfter(date, compareDate)
  } catch {
    return false
  }
}

/**
 * Check if a timestamp is expired (before current time)
 */
export function isExpired(timestamp: string): boolean {
  try {
    const date = parseISO(timestamp)
    const now = new Date()
    return isValid(date) && !isAfter(date, now)
  } catch {
    return true
  }
}

/**
 * Check if two messages are consecutive (same sender within 5 minutes)
 */
export function areMessagesConsecutive(
  currentTimestamp: string,
  previousTimestamp?: string,
): boolean {
  if (!previousTimestamp) return false

  try {
    const currentDate = parseISO(currentTimestamp)
    const previousDate = parseISO(previousTimestamp)

    if (!isValid(currentDate) || !isValid(previousDate)) return false

    const diffMinutes = differenceInMinutes(currentDate, previousDate)
    return diffMinutes < 5
  } catch {
    return false
  }
}

/**
 * Get current year
 */
export function getCurrentYear(): number {
  return getYear(new Date())
}

/**
 * Generate ISO timestamp for current time
 */
export function generateTimestamp(): string {
  return formatISO(new Date())
}

/**
 * Generate unique ID with timestamp
 */
export function generateUniqueId(prefix = 'temp'): string {
  return `${prefix}-${Date.now()}`
}

/**
 * Convert timestamp to date object
 */
export function timestampToDate(ts: Timestamp | undefined): Date | null {
  if (!ts) return null
  const millis = Number(ts.seconds) * 1000 + Math.floor(ts.nanos / 1_000_000)
  return new Date(millis)
}
