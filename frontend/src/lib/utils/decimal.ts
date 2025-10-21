/**
 * Decimal utilities for handling monetary values with precision
 *
 * GraphQL Decimal scalar is transmitted as string to maintain precision.
 * We use decimal.js for calculations on the frontend.
 */

import Decimal from 'decimal.js'

/**
 * Parse GraphQL Decimal string to Decimal.js instance
 *
 * @param value - GraphQL Decimal as string (e.g., "19.99")
 * @returns Decimal.js instance
 */
export function parseDecimal(value: string): Decimal {
  return new Decimal(value)
}

/**
 * Convert Decimal.js instance to GraphQL Decimal string
 *
 * @param value - Decimal.js instance
 * @returns String representation for GraphQL
 */
export function toDecimalString(value: Decimal | number): string {
  if (value instanceof Decimal) {
    return value.toString()
  }
  return new Decimal(value).toString()
}

/**
 * Convert Decimal.js instance to number for display
 * Use only for display purposes, not for calculations
 *
 * @param value - Decimal.js instance
 * @returns Number for display
 */
export function toNumber(value: Decimal): number {
  return value.toNumber()
}

/**
 * Format Decimal as currency string
 *
 * @param value - Decimal.js instance
 * @param currency - Currency code (default: USD)
 * @param locale - Locale for formatting (default: en-US)
 * @returns Formatted currency string
 */
export function formatCurrency(
  value: Decimal,
  currency: string = 'USD',
  locale: string = 'en-US',
): string {
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
  }).format(value.toNumber())
}

/**
 * Create Decimal from number
 *
 * @param value - Number value
 * @returns Decimal.js instance
 */
export function fromNumber(value: number): Decimal {
  return new Decimal(value)
}
