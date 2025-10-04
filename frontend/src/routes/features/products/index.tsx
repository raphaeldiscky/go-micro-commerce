import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'
import { PATH_FEATURES } from '@/constants'
import { useProducts } from '@/hooks/products/useProducts'
import type { FileRoutesByPath } from '@tanstack/react-router'
import { createFileRoute } from '@tanstack/react-router'
import { AlertCircle, Package } from 'lucide-react'
import { useMemo } from 'react'
import { ProductList } from '../../../components/features/products/ProductList'
import { ProductPagination } from '../../../components/features/products/ProductPagination'

export const Route = createFileRoute(
  PATH_FEATURES.products.root as keyof FileRoutesByPath,
)({
  component: ProductsPage,
})

function ProductsPage() {
  const {
    data,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
  } = useProducts('10')

  // Flatten all pages into a single array
  const allProducts = useMemo(() => {
    return data?.pages.flatMap((page) => page.products) ?? []
  }, [data?.pages])

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-white dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4 py-12">
        {/* Header */}
        <div className="text-center mb-12">
          <h1 className="text-4xl font-bold text-gray-900 dark:text-white mb-4 flex items-center justify-center gap-2">
            <Package className="h-10 w-10" />
            Product Catalog
          </h1>
          <p className="text-lg text-gray-600 dark:text-gray-300 max-w-2xl mx-auto">
            Browse our complete product catalog with real-time inventory
            tracking and cursor-based pagination.
          </p>
        </div>

        {/* Error State */}
        {error && (
          <Alert variant="destructive" className="mb-8 max-w-2xl mx-auto">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Error</AlertTitle>
            <AlertDescription>
              Failed to load products. Please try again later.
              <br />
              <span className="text-xs font-mono mt-2 block">
                {error.message}
              </span>
            </AlertDescription>
          </Alert>
        )}

        {/* Loading Skeleton */}
        {isLoading && allProducts.length === 0 && (
          <div className="grid gap-6 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="space-y-4">
                <Skeleton className="h-[300px] w-full rounded-lg" />
              </div>
            ))}
          </div>
        )}

        {/* Products List */}
        {!isLoading && !error && <ProductList products={allProducts} />}

        {/* Pagination */}
        {!error && allProducts.length > 0 && (
          <ProductPagination
            currentCount={allProducts.length}
            hasNext={hasNextPage}
            isLoading={isFetchingNextPage}
            onLoadMore={() => fetchNextPage()}
          />
        )}
      </div>
    </div>
  )
}
