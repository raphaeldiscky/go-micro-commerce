export type OrderStatus = 'pending' | 'processing' | 'completed' | 'cancelled'

export interface MockOrder {
  id: string
  customerName: string
  customerEmail: string
  total: number
  status: OrderStatus
  items: number
  createdAt: string
}

const statuses: Array<OrderStatus> = [
  'pending',
  'processing',
  'completed',
  'cancelled',
]
const firstNames = [
  'John',
  'Jane',
  'Michael',
  'Emily',
  'David',
  'Sarah',
  'James',
  'Emma',
  'Robert',
  'Olivia',
  'William',
  'Ava',
  'Richard',
  'Sophia',
  'Thomas',
]
const lastNames = [
  'Smith',
  'Johnson',
  'Williams',
  'Brown',
  'Jones',
  'Garcia',
  'Miller',
  'Davis',
  'Rodriguez',
  'Martinez',
  'Wilson',
  'Anderson',
  'Taylor',
  'Moore',
  'Jackson',
]

function generateRandomOrder(index: number): MockOrder {
  const firstName = firstNames[index % firstNames.length]
  const lastName =
    lastNames[Math.floor(index / firstNames.length) % lastNames.length]
  const status = statuses[index % statuses.length]
  const total = Math.floor(Math.random() * 50000 + 2000) / 100
  const items = Math.floor(Math.random() * 5) + 1
  const daysAgo = Math.floor(Math.random() * 90)
  const date = new Date()
  date.setDate(date.getDate() - daysAgo)

  return {
    createdAt: date.toISOString(),
    customerEmail: `${firstName.toLowerCase()}.${lastName.toLowerCase()}@example.com`,
    customerName: `${firstName} ${lastName}`,
    id: `ORD-${String(1000 + index).padStart(4, '0')}`,
    items,
    status,
    total,
  }
}

export const mockOrders: Array<MockOrder> = Array.from({ length: 60 }, (_, i) =>
  generateRandomOrder(i),
)
