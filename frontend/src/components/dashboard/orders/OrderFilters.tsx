import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { OrderStatus } from '@/mocks/orders'
import { Search, X } from 'lucide-react'

interface OrderFiltersProps {
  searchQuery: string
  statusFilter: OrderStatus | 'all'
  onSearchChange: (value: string) => void
  onStatusChange: (value: OrderStatus | 'all') => void
  onReset: () => void
}

export function OrderFilters({
  onReset,
  onSearchChange,
  onStatusChange,
  searchQuery,
  statusFilter,
}: OrderFiltersProps) {
  return (
    <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div className="relative flex-1 sm:max-w-sm">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search by order ID or customer..."
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-9"
        />
      </div>

      <div className="flex items-center gap-2">
        <Select value={statusFilter} onValueChange={onStatusChange}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Filter by status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="pending">Pending</SelectItem>
            <SelectItem value="processing">Processing</SelectItem>
            <SelectItem value="completed">Completed</SelectItem>
            <SelectItem value="cancelled">Cancelled</SelectItem>
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
