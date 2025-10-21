import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { UserRole } from '@/lib/mock-data/users'
import { Search, X } from 'lucide-react'

interface UserFiltersProps {
  searchQuery: string
  roleFilter: UserRole | 'all'
  onSearchChange: (value: string) => void
  onRoleChange: (value: UserRole | 'all') => void
  onReset: () => void
}

export function UserFilters({
  onReset,
  onRoleChange,
  onSearchChange,
  roleFilter,
  searchQuery,
}: UserFiltersProps) {
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div className="relative flex-1 sm:max-w-sm">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search by name or email..."
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-9"
        />
      </div>

      <div className="flex items-center gap-2">
        <Select value={roleFilter} onValueChange={onRoleChange}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Filter by role" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Roles</SelectItem>
            <SelectItem value="admin">Admin</SelectItem>
            <SelectItem value="user">User</SelectItem>
          </SelectContent>
        </Select>

        <Button
          variant="outline"
          size="icon"
          onClick={onReset}
          title="Reset filters"
        >
          <X className="h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
