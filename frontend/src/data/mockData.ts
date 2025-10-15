import type {
  MockProduct,
  PaymentMethod,
  ShippingOption,
} from '@/types/cart'

// Mock products that match the existing Product structure
export const mockProducts: Array<MockProduct> = [
  {
    id: 'prod-1',
    name: 'Wireless Bluetooth Headphones',
    description:
      'Premium noise-cancelling wireless headphones with 30-hour battery life',
    price: 299.99,
    quantity: 50,
    reservedQuantity: 5,
    image: '/api/placeholder/200/200',
    category: 'Electronics',
    sku: 'WBH-001',
    version: 1,
    createdAt: new Date('2024-01-15'),
    updatedAt: new Date('2024-01-20'),
  },
  {
    id: 'prod-2',
    name: 'Smart Watch Pro',
    description: 'Fitness tracking smartwatch with heart rate monitor and GPS',
    price: 199.99,
    quantity: 75,
    reservedQuantity: 10,
    image: '/api/placeholder/200/200',
    category: 'Electronics',
    sku: 'SWP-002',
    version: 2,
    createdAt: new Date('2024-01-10'),
    updatedAt: new Date('2024-01-18'),
  },
  {
    id: 'prod-3',
    name: 'Organic Cotton T-Shirt',
    description: 'Comfortable and sustainable organic cotton t-shirt',
    price: 29.99,
    quantity: 200,
    reservedQuantity: 15,
    image: '/api/placeholder/200/200',
    category: 'Clothing',
    sku: 'OCT-003',
    version: 1,
    createdAt: new Date('2024-01-05'),
    updatedAt: new Date('2024-01-12'),
  },
  {
    id: 'prod-4',
    name: 'Stainless Steel Water Bottle',
    description: 'Insulated water bottle keeps drinks cold for 24 hours',
    price: 24.99,
    quantity: 150,
    reservedQuantity: 8,
    image: '/api/placeholder/200/200',
    category: 'Sports',
    sku: 'SSW-004',
    version: 1,
    createdAt: new Date('2024-01-08'),
    updatedAt: new Date('2024-01-16'),
  },
  {
    id: 'prod-5',
    name: 'Yoga Mat Premium',
    description: 'Non-slip exercise yoga mat with carrying strap',
    price: 39.99,
    quantity: 80,
    reservedQuantity: 12,
    image: '/api/placeholder/200/200',
    category: 'Sports',
    sku: 'YMP-005',
    version: 1,
    createdAt: new Date('2024-01-12'),
    updatedAt: new Date('2024-01-19'),
  },
  {
    id: 'prod-6',
    name: 'Laptop Backpack',
    description: 'Water-resistant backpack with laptop compartment up to 15.6"',
    price: 59.99,
    quantity: 60,
    reservedQuantity: 6,
    image: '/api/placeholder/200/200',
    category: 'Accessories',
    sku: 'LBB-006',
    version: 1,
    createdAt: new Date('2024-01-14'),
    updatedAt: new Date('2024-01-21'),
  },
  {
    id: 'prod-7',
    name: 'Wireless Mouse',
    description: 'Ergonomic wireless mouse with precision tracking',
    price: 34.99,
    quantity: 100,
    reservedQuantity: 20,
    image: '/api/placeholder/200/200',
    category: 'Electronics',
    sku: 'WMS-007',
    version: 1,
    createdAt: new Date('2024-01-11'),
    updatedAt: new Date('2024-01-17'),
  },
  {
    id: 'prod-8',
    name: 'Ceramic Coffee Mug Set',
    description: 'Set of 4 handcrafted ceramic coffee mugs',
    price: 44.99,
    quantity: 40,
    reservedQuantity: 5,
    image: '/api/placeholder/200/200',
    category: 'Home',
    sku: 'CCM-008',
    version: 1,
    createdAt: new Date('2024-01-09'),
    updatedAt: new Date('2024-01-15'),
  },
  {
    id: 'prod-9',
    name: 'Running Shoes',
    description: 'Lightweight running shoes with advanced cushioning',
    price: 89.99,
    quantity: 35,
    reservedQuantity: 8,
    image: '/api/placeholder/200/200',
    category: 'Sports',
    sku: 'RS-009',
    version: 2,
    createdAt: new Date('2024-01-13'),
    updatedAt: new Date('2024-01-20'),
  },
  {
    id: 'prod-10',
    name: 'Portable Phone Charger',
    description: '10000mAh portable power bank with fast charging',
    price: 49.99,
    quantity: 90,
    reservedQuantity: 15,
    image: '/api/placeholder/200/200',
    category: 'Electronics',
    sku: 'PPC-010',
    version: 1,
    createdAt: new Date('2024-01-16'),
    updatedAt: new Date('2024-01-22'),
  },
]

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

// Mock payment methods
export const mockPaymentMethods: Array<PaymentMethod> = [
  {
    id: 'pay-1',
    name: 'Credit Card',
    type: 'credit_card',
    icon: '/api/placeholder/32/32',
    isActive: true,
    description: 'Visa, Mastercard, American Express',
  },
  {
    id: 'pay-2',
    name: 'Debit Card',
    type: 'debit_card',
    icon: '/api/placeholder/32/32',
    isActive: true,
    description: 'Visa Debit, Maestro',
  },
  {
    id: 'pay-3',
    name: 'Bank Transfer',
    type: 'bank_transfer',
    icon: '/api/placeholder/32/32',
    isActive: true,
    description: 'Direct bank transfer',
  },
  {
    id: 'pay-4',
    name: 'PayPal',
    type: 'ewallet',
    icon: '/api/placeholder/32/32',
    isActive: true,
    description: 'PayPal digital wallet',
  },
  {
    id: 'pay-5',
    name: 'Cash on Delivery',
    type: 'cod',
    icon: '/api/placeholder/32/32',
    isActive: true,
    description: 'Pay when you receive',
  },
]

// Helper functions to find mock data
export const findProductById = (id: string): MockProduct | undefined => {
  return mockProducts.find((product) => product.id === id)
}


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

// Generate placeholder image URL
export const getPlaceholderImage = (
  width: number = 200,
  height: number = 200,
): string => {
  return `https://picsum.photos/seed/${Math.random().toString(36).substring(7)}/${width}/${height}.jpg`
}

// Format currency
export const formatCurrency = (amount: number): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
  }).format(amount)
}

