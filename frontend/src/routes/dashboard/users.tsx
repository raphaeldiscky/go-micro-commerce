import { Card, CardContent } from '@/components/ui/card'
import { useUsers } from '@/hooks/dashboard'
import type { UserRole } from '@/mocks/users'
import { createFileRoute } from '@tanstack/react-router'
import { Loader2 } from 'lucide-react'
import { useState } from 'react'
import { CursorPagination } from '../../components/dashboard/shared/CursorPagination'
import { UserFilters } from '../../components/dashboard/users/UserFilters'
import { UserTable } from '../../components/dashboard/users/UserTable'
import { DashboardHeader } from '../../components/layout/DashboardHeader'
import { PATH_DASHBOARD } from '../../constants/routes'

export const Route = createFileRoute('/dashboard/users')({
  component: UsersPage,
})

function UsersPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [roleFilter, setRoleFilter] = useState<UserRole | 'all'>('all')
  const [currentCursor, setCurrentCursor] = useState<string | undefined>()
  const [cursorHistory, setCursorHistory] = useState<Array<string>>([])

  const { data: paginatedData, isLoading } = useUsers({
    searchQuery,
    role: roleFilter,
    cursor: currentCursor,
  })

  const handleReset = () => {
    setSearchQuery('')
    setRoleFilter('all')
    setCurrentCursor(undefined)
    setCursorHistory([])
  }

  const handleNextPage = () => {
    if (paginatedData?.endCursor) {
      // Add current cursor to history before moving forward (if not first page)
      if (currentCursor) {
        setCursorHistory((prev) => [...prev, currentCursor])
      }
      setCurrentCursor(paginatedData.endCursor)
    }
  }

  const handlePreviousPage = () => {
    if (cursorHistory.length > 0) {
      // Go back to previous cursor
      const previousCursors = [...cursorHistory]
      const previousCursor = previousCursors.pop()
      setCursorHistory(previousCursors)
      setCurrentCursor(previousCursor)
    } else {
      // Go to first page
      setCurrentCursor(undefined)
      setCursorHistory([])
    }
  }

  // Reset cursor when filters change
  const handleSearchChange = (value: string) => {
    setSearchQuery(value)
    setCurrentCursor(undefined)
    setCursorHistory([])
  }

  const handleRoleChange = (value: UserRole | 'all') => {
    setRoleFilter(value)
    setCurrentCursor(undefined)
    setCursorHistory([])
  }

  return (
    <div>
      <DashboardHeader
        title="Users Management"
        breadcrumbs={[
          { label: 'Dashboard', href: PATH_DASHBOARD.root },
          { label: 'Users' },
        ]}
      />

      <div className="space-y-4 p-6">
        <UserFilters
          searchQuery={searchQuery}
          roleFilter={roleFilter}
          onSearchChange={handleSearchChange}
          onRoleChange={handleRoleChange}
          onReset={handleReset}
        />

        <Card>
          <CardContent className="p-0">
            {isLoading ? (
              <div className="flex items-center justify-center py-12">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
              </div>
            ) : (
              <UserTable users={paginatedData?.data || []} />
            )}
          </CardContent>
        </Card>

        {paginatedData && (
          <CursorPagination
            hasNextPage={paginatedData.hasNextPage}
            hasPreviousPage={paginatedData.hasPreviousPage}
            onNextPage={handleNextPage}
            onPreviousPage={handlePreviousPage}
            disabled={isLoading}
          />
        )}
      </div>
    </div>
  )
}
