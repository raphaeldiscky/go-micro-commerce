import type {
  PaymentGateway,
  PaymentMethod,
  ShippingOption,
} from '@/types/cart'
import type { ShippingCarrier } from '@/types/order'

// Mock shipping options
export const mockShippingOptions: Array<ShippingOption> = [
  {
    id: 'ship-1',
    name: 'Standard Delivery',
    description: 'Standard shipping within 5-7 business days',
    price: 9.99,
    estimatedDays: { min: 5, max: 7 },
    isActive: true,
  },
  {
    id: 'ship-2',
    name: 'Express Delivery',
    description: 'Fast shipping within 2-3 business days',
    price: 19.99,
    estimatedDays: { min: 2, max: 3 },
    isActive: true,
  },
  {
    id: 'ship-3',
    name: 'Next Day Delivery',
    description: 'Delivery by next business day',
    price: 29.99,
    estimatedDays: { min: 1, max: 1 },
    isActive: true,
  },
  {
    id: 'ship-4',
    name: 'Free Economy Shipping',
    description: 'Free shipping on orders over $100',
    price: 0,
    estimatedDays: { min: 7, max: 10 },
    isActive: true,
  },
]

export const mockPaymentGateways: Array<PaymentGateway> = [
  {
    id: 'gateway-1',
    name: 'Stripe',
    type: 'stripe',
  },
  {
    id: 'gateway-2',
    name: 'PayPal',
    type: 'paypal',
  },
]

// Mock payment methods
export const mockPaymentMethods: Array<PaymentMethod> = [
  {
    id: 'pay-1',
    name: 'Card',
    type: 'card',
    icon: '/api/placeholder/32/32',
    isActive: true,
    description: 'Visa, Mastercard, American Express',
    supportedGateways: mockPaymentGateways,
  },
]

// Mock shipping carriers
export const mockShippingCarriers: Array<ShippingCarrier> = [
  {
    id: 'carrier-fedex',
    name: 'FedEx',
    type: 'standard',
    estimatedDays: { min: 5, max: 7 },
  },
  {
    id: 'carrier-ups',
    name: 'UPS',
    type: 'express',
    estimatedDays: { min: 2, max: 3 },
  },
  {
    id: 'carrier-dhl',
    name: 'DHL Express',
    type: 'overnight',
    estimatedDays: { min: 1, max: 1 },
  },
  {
    id: 'carrier-usps',
    name: 'USPS',
    type: 'standard',
    estimatedDays: { min: 7, max: 10 },
  },
]

// Default warehouse address for shipping calculations
export const DEFAULT_WAREHOUSE_ADDRESS = {
  city: 'San Francisco',
  state: 'CA',
  postalCode: '94103',
  country: 'US',
}

// Default product dimensions (used when product doesn't have specific dimensions)
export const DEFAULT_PRODUCT_DIMENSIONS = {
  width: 10,
  height: 10,
  length: 10,
  unit: 'cm',
}

// Default product weight in kg
export const DEFAULT_PRODUCT_WEIGHT_KG = 0.5

export const findShippingOptionById = (
  id: string,
): ShippingOption | undefined => {
  return mockShippingOptions.find(
    (option) => option.id === id && option.isActive,
  )
}

export const findPaymentMethodById = (
  id: string,
): PaymentMethod | undefined => {
  return mockPaymentMethods.find(
    (method) => method.id === id && method.isActive,
  )
}

export const findShippingCarrierById = (
  id: string,
): ShippingCarrier | undefined => {
  return mockShippingCarriers.find((carrier) => carrier.id === id)
}

// Map shipping option to carrier ID
export const mapShippingOptionToCarrier = (
  shippingOptionId: string,
): string => {
  const carrierMap: Record<string, string> = {
    'ship-1': 'carrier-fedex',
    'ship-2': 'carrier-ups',
    'ship-3': 'carrier-dhl',
    'ship-4': 'carrier-usps',
  }
  return carrierMap[shippingOptionId] || 'carrier-fedex'
}

// Generate placeholder image URL
export const getPlaceholderImage = (
  width: number = 200,
  height: number = 200,
): string => {
  return `https://picsum.photos/seed/${Math.random().toString(36).substring(7)}/${width}/${height}.jpg`
}
