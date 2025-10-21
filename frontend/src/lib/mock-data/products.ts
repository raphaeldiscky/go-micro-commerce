export type ProductStatus = 'active' | 'draft' | 'archived'

export interface MockProduct {
  id: string
  name: string
  sku: string
  price: number
  stock: number
  status: ProductStatus
  category: string
  imageUrl: string
  createdAt: string
}

const productNames = [
  'Wireless Headphones',
  'Smart Watch',
  'Laptop Stand',
  'Mechanical Keyboard',
  'USB-C Hub',
  'Portable Charger',
  'Webcam HD',
  'Gaming Mouse',
  'Monitor 27"',
  'Desk Lamp',
  'Cable Organizer',
  'Phone Case',
  'Screen Protector',
  'Bluetooth Speaker',
  'Wireless Earbuds',
  'Power Bank',
  'HDMI Cable',
  'Mouse Pad',
  'Laptop Bag',
  'USB Drive',
]

const categories = [
  'Electronics',
  'Accessories',
  'Audio',
  'Peripherals',
  'Storage',
]
const statuses: Array<ProductStatus> = ['active', 'draft', 'archived']

function generateRandomProduct(index: number): MockProduct {
  const name = productNames[index % productNames.length]
  const category = categories[index % categories.length]
  const status = statuses[index % statuses.length]
  const price = Math.floor(Math.random() * 50000 + 1000) / 100
  const stock = Math.floor(Math.random() * 200)
  const daysAgo = Math.floor(Math.random() * 180)
  const date = new Date()
  date.setDate(date.getDate() - daysAgo)

  return {
    category,
    createdAt: date.toISOString(),
    id: `PRD-${String(100 + index).padStart(4, '0')}`,
    imageUrl: `https://placehold.co/400x400/e2e8f0/475569?text=${encodeURIComponent(name)}`,
    name: `${name} ${index > 19 ? `v${Math.floor(index / 20)}` : ''}`.trim(),
    price,
    sku: `SKU-${String(10000 + index * 11).padStart(5, '0')}`,
    status,
    stock,
  }
}

export const mockProducts: Array<MockProduct> = Array.from(
  { length: 60 },
  (_, i) => generateRandomProduct(i),
)
