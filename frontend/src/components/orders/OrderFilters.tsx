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
import type { OrderStatus } from '@/types/order'
import { RotateCcwIcon, SearchIcon } from 'lucide-react'
import { useState } from 'react'

const ORDER_STATUSES: Array<{ value: OrderStatus; label: string }> = [
  { value: 'pending', label: 'Pending' },
  { value: 'processing', label: 'Processing' },
  { value: 'payment_pending', label: 'Payment Pending' },
  { value: 'payment_expired', label: 'Payment Expired' },
  { value: 'paid', label: 'Paid' },
  { value: 'shipped', label: 'Shipped' },
  { value: 'delivered', label: 'Delivered' },
  { value: 'completed', label: 'Completed' },
  { value: 'failed', label: 'Failed' },
  { value: 'canceled', label: 'Canceled' },
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
      dateFrom: undefined,
      dateTo: undefined,
      minAmount: undefined,
      maxAmount: undefined,
    })
  }

  const activeFiltersCount = [
    filters.status,
    filters.search,
    filters.dateFrom,
    filters.dateTo,
    filters.minAmount,
    filters.maxAmount,
  ].filter(Boolean).length

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

      {/* Future filters can be added here */}
      {/*
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="flex flex-col gap-2">
          <Label htmlFor="date-from">Date From</Label>
          <Input
            id="date-from"
            type="date"
            value={filters.dateFrom || ''}
            onChange={(e) => setFilters({ ...filters, dateFrom: e.target.value || undefined })}
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label htmlFor="date-to">Date To</Label>
          <Input
            id="date-to"
            type="date"
            value={filters.dateTo || ''}
            onChange={(e) => setFilters({ ...filters, dateTo: e.target.value || undefined })}
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label htmlFor="min-amount">Min Amount</Label>
          <Input
            id="min-amount"
            type="number"
            placeholder="0.00"
            value={filters.minAmount || ''}
            onChange={(e) => setFilters({ ...filters, minAmount: e.target.value ? Number(e.target.value) : undefined })}
          />
        </div>
        <div className="flex flex-col gap-2">
          <Label htmlFor="max-amount">Max Amount</Label>
          <Input
            id="max-amount"
            type="number"
            placeholder="0.00"
            value={filters.maxAmount || ''}
            onChange={(e) => setFilters({ ...filters, maxAmount: e.target.value ? Number(e.target.value) : undefined })}
          />
        </div>
      </div>
      */}
    </div>
  )
}
