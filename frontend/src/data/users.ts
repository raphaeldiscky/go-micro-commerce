export type UserRole = 'admin' | 'user'
export type UserStatus = 'active' | 'inactive'

export interface MockUser {
  id: string
  firstName: string
  lastName: string
  email: string
  role: UserRole
  status: UserStatus
  joinedDate: string
  lastActive: string
}

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
  'Isabella',
  'Charles',
  'Mia',
  'Daniel',
  'Charlotte',
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
  'Martin',
  'Lee',
  'Thompson',
  'White',
  'Harris',
]

const roles: Array<UserRole> = ['user', 'admin']
const statuses: Array<UserStatus> = ['active', 'inactive']

function generateRandomUser(index: number): MockUser {
  const firstName = firstNames[index % firstNames.length]
  const lastName =
    lastNames[Math.floor(index / firstNames.length) % lastNames.length]
  const role = index < 5 ? 'admin' : roles[index % roles.length]
  const status = statuses[index % 10 === 0 ? 1 : 0]
  const joinedDaysAgo = Math.floor(Math.random() * 365) + 30
  const lastActiveDaysAgo = Math.floor(Math.random() * 30)

  const joinedDate = new Date()
  joinedDate.setDate(joinedDate.getDate() - joinedDaysAgo)

  const lastActive = new Date()
  lastActive.setDate(lastActive.getDate() - lastActiveDaysAgo)

  return {
    email: `${firstName.toLowerCase()}.${lastName.toLowerCase()}@example.com`,
    firstName,
    id: `USR-${String(100 + index).padStart(4, '0')}`,
    joinedDate: joinedDate.toISOString(),
    lastActive: lastActive.toISOString(),
    lastName,
    role,
    status,
  }
}

export const mockUsers: Array<MockUser> = Array.from({ length: 60 }, (_, i) =>
  generateRandomUser(i),
)
