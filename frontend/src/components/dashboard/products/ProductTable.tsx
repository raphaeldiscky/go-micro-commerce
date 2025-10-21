import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { MockProduct } from '@/lib/mock-data/products'
import { formatDate } from '@/lib/utils/date'
import { Edit, Eye } from 'lucide-react'

interface ProductTableProps {
  products: Array<MockProduct>
}

function getStatusColor(status: MockProduct['status']) {
  switch (status) {
    case 'active':
      return 'border-green-200 text-green-700 dark:border-green-800 dark:text-green-400'
    case 'draft':
      return 'border-amber-200 text-amber-700 dark:border-amber-800 dark:text-amber-400'
    case 'archived':
      return 'border-gray-200 text-gray-700 dark:border-gray-800 dark:text-gray-400'
    default:
      return 'border-gray-200 text-gray-700 dark:border-gray-800 dark:text-gray-400'
  }
}

export function ProductTable({ products }: ProductTableProps) {
  if (products.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center rounded-lg border border-dashed">
        <p className="text-muted-foreground">No products found</p>
      </div>
    )
  }

  return (
    <div className="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Product</TableHead>
            <TableHead>SKU</TableHead>
            <TableHead>Category</TableHead>
            <TableHead>Price</TableHead>
            <TableHead>Stock</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {products.map((product) => (
            <TableRow key={product.id}>
              <TableCell>
                <div className="flex items-center gap-3">
                  <img
                    src={product.imageUrl}
                    alt={product.name}
                    className="h-10 w-10 rounded-md object-cover"
                  />
                  <div>
                    <div className="font-medium">{product.name}</div>
                    <div className="text-sm text-muted-foreground">
                      {product.id}
                    </div>
                  </div>
                </div>
              </TableCell>
              <TableCell className="font-mono text-sm">{product.sku}</TableCell>
              <TableCell>{product.category}</TableCell>
              <TableCell>${product.price.toFixed(2)}</TableCell>
              <TableCell>
                <span
                  className={
                    product.stock < 20
                      ? 'text-red-500 font-medium'
                      : 'text-foreground'
                  }
                >
                  {product.stock}
                </span>
              </TableCell>
              <TableCell>
                <Badge
                  className={getStatusColor(product.status)}
                  variant="outline"
                >
                  {product.status}
                </Badge>
              </TableCell>
              <TableCell>{formatDate(product.createdAt)}</TableCell>
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
