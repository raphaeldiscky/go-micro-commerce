import type { Product } from '@/proto/product/v1/product_pb'
import { PackageOpen } from 'lucide-react'
import { ProductCard } from './ProductCard'

interface ProductListProps {
  products: Array<Product>
}

export function ProductList({ products }: ProductListProps) {
  if (products.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-16 px-4">
        <PackageOpen className="h-16 w-16 text-gray-400 dark:text-gray-600 mb-4" />
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
          No products found
        </h3>
        <p className="text-sm text-gray-600 dark:text-gray-400 text-center max-w-md">
          There are no products available at the moment. Please check back
          later.
        </p>
      </div>
    )
  }

  return (
    <div className="grid gap-6 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
      {products.map((product) => (
        <ProductCard key={product.id} product={product} />
      ))}
    </div>
  )
}
