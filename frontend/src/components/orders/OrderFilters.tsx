import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { useOrderFilters, useOrderStore } from '@/store/orderStore'
import { OrderStatus } from '@/types/__generated__/graphql'
import { RotateCcwIcon, SearchIcon } from 'lucide-react'
import { useState } from 'react'

const ORDER_STATUSES: Array<{ value: OrderStatus; label: string }> = [
  { value: OrderStatus.Pending, label: 'Pending' },
  { value: OrderStatus.Processing, label: 'Processing' },
  { value: OrderStatus.PaymentPending, label: 'Payment Pending' },
  { value: OrderStatus.PaymentExpired, label: 'Payment Expired' },
  { value: OrderStatus.Paid, label: 'Paid' },
  { value: OrderStatus.Shipped, label: 'Shipped' },
  { value: OrderStatus.Delivered, label: 'Delivered' },
  { value: OrderStatus.Completed, label: 'Completed' },
  { value: OrderStatus.Failed, label: 'Failed' },
  { value: OrderStatus.Canceled, label: 'Canceled' },
]

interface OrderFiltersProps {
  className?: string
}

export function OrderFilters({ className }: OrderFiltersProps) {
  const filters = useOrderFilters()
  const { setFilters } = useOrderStore()
  const [searchInput, setSearchInput] = useState(filters.search || '')

  const handleStatusChange = (status: string) => {
    const newStatus = status === 'all' ? undefined : (status as OrderStatus)
    setFilters({ ...filters, status: newStatus })
  }

  const handleSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value
    setSearchInput(value)
  }

  const handleSearchSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    setFilters({ ...filters, search: searchInput.trim() || undefined })
  }

  const handleClearFilters = () => {
    setSearchInput('')
    setFilters({
      status: undefined,
      search: undefined,
    })
  }

  const activeFiltersCount = [filters.status, filters.search].filter(
    Boolean,
  ).length

  return (
    <div className={cn('space-y-4', className)}>
      <form onSubmit={handleSearchSubmit} className="flex gap-2">
        <div className="relative flex-1">
          <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search orders by ID or customer..."
            value={searchInput}
            onChange={handleSearchChange}
            className="pl-10"
          />
        </div>
        <Button type="submit" size="default">
          <SearchIcon className="mr-2 h-4 w-4" />
          Search
        </Button>
      </form>

      <div className="flex flex-wrap gap-4">
        <div className="flex flex-col gap-2 min-w-0 flex-1">
          <Label htmlFor="status-filter" className="text-sm font-medium">
            Status
          </Label>
          <Select
            value={filters.status || 'all'}
            onValueChange={handleStatusChange}
          >
            <SelectTrigger id="status-filter" className="w-full sm:w-48">
              <SelectValue placeholder="All statuses" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All statuses</SelectItem>
              {ORDER_STATUSES.map((status) => (
                <SelectItem key={status.value} value={status.value}>
                  {status.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        <div className="flex items-end gap-2">
          {activeFiltersCount > 0 && (
            <Button
              variant="outline"
              size="default"
              onClick={handleClearFilters}
              className="flex items-center gap-2"
            >
              <RotateCcwIcon className="h-4 w-4" />
              Clear ({activeFiltersCount})
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
