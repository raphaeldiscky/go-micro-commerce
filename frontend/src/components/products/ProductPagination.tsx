import { Button } from '@/components/ui/button'
import { ChevronRight, Loader2 } from 'lucide-react'

interface ProductPaginationProps {
  hasNext: boolean
  isLoading: boolean
  currentCount: number
  onLoadMore: () => void
}

export function ProductPagination({
  currentCount,
  hasNext,
  isLoading,
  onLoadMore,
}: ProductPaginationProps) {
  return (
    <div className="flex flex-col items-center justify-center space-y-4 py-8">
      <div className="text-sm text-gray-600 dark:text-gray-400">
        Showing {currentCount} product{currentCount !== 1 ? 's' : ''}
      </div>

      {hasNext && (
        <Button
          disabled={isLoading}
          onClick={onLoadMore}
          size="lg"
          variant="outline"
        >
          {isLoading ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Loading...
            </>
          ) : (
            <>
              Load More
              <ChevronRight className="ml-2 h-4 w-4" />
            </>
          )}
        </Button>
      )}

      {!hasNext && currentCount > 0 && (
        <p className="text-sm text-gray-500 dark:text-gray-400">
          You&apos;ve reached the end of the list
        </p>
      )}
    </div>
  )
}
