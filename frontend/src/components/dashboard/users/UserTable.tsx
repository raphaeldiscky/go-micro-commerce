import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { MockUser } from '@/data/users'
import { formatDate } from '@/lib/utils/date'
import { Edit, Eye, Shield } from 'lucide-react'

interface UserTableProps {
  users: Array<MockUser>
}

function getStatusColor(status: MockUser['status']) {
  switch (status) {
    case 'active':
      return 'border-green-200 text-green-700 dark:border-green-800 dark:text-green-400'
    case 'inactive':
      return 'border-gray-200 text-gray-700 dark:border-gray-800 dark:text-gray-400'
    default:
      return 'border-gray-200 text-gray-700 dark:border-gray-800 dark:text-gray-400'
  }
}

function getRoleColor(role: MockUser['role']) {
  switch (role) {
    case 'admin':
      return 'border-purple-200 text-purple-700 dark:border-purple-800 dark:text-purple-400'
    case 'user':
      return 'border-blue-200 text-blue-700 dark:border-blue-800 dark:text-blue-400'
    default:
      return 'border-gray-200 text-gray-700 dark:border-gray-800 dark:text-gray-400'
  }
}

export function UserTable({ users }: UserTableProps) {
  if (users.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center rounded-lg border border-dashed">
        <p className="text-muted-foreground">No users found</p>
      </div>
    )
  }

  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>User</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Role</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Joined Date</TableHead>
            <TableHead>Last Active</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {users.map((user) => (
            <TableRow key={user.id}>
              <TableCell>
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary text-primary-foreground">
                    {user.firstName[0]}
                    {user.lastName[0]}
                  </div>
                  <div>
                    <div className="font-medium">
                      {user.firstName} {user.lastName}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {user.id}
                    </div>
                  </div>
                </div>
              </TableCell>
              <TableCell>{user.email}</TableCell>
              <TableCell>
                <Badge className={getRoleColor(user.role)} variant="outline">
                  {user.role === 'admin' && <Shield className="mr-1 h-3 w-3" />}
                  {user.role}
                </Badge>
              </TableCell>
              <TableCell>
                <Badge
                  className={getStatusColor(user.status)}
                  variant="outline"
                >
                  {user.status}
                </Badge>
              </TableCell>
              <TableCell>{formatDate(user.joinedDate)}</TableCell>
              <TableCell>{formatDate(user.lastActive)}</TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                  <button className="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-accent">
                    <Eye className="h-4 w-4" />
                  </button>
                  <button className="inline-flex h-8 w-8 items-center justify-center rounded-md hover:bg-accent">
                    <Edit className="h-4 w-4" />
                  </button>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
