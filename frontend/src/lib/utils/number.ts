import Decimal from 'decimal.js'

/**
 * Format Decimal as currency string
 *
 * @param value - Decimal.js instance, string, or number
 * @param currency - Currency code (default: USD)
 * @param locale - Locale for formatting (default: en-US)
 * @returns Formatted currency string
 */
export function fCurrency(
  value: Decimal | string | number,
  currency: string = 'USD',
  locale: string = 'en-US',
): string {
  const num: number =
    typeof value === 'string'
      ? Number(value)
      : value instanceof Decimal
        ? value.toNumber()
        : value
  return new Intl.NumberFormat(locale, {
    style: 'currency',
    currency,
  }).format(num)
}
