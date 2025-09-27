import { create } from '@bufbuild/protobuf'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { productApi } from '../lib/api/product'
import type { Product } from '../proto/product/v1/product_pb'
import {
  BatchGetProductsByIDsRequestSchema,
  ConfirmProductsDeductionRequestSchema,
  ReleaseProductsRequestSchema,
  ReserveProductsRequestSchema,
  RestoreProductsRequestSchema,
} from '../proto/product/v1/product_pb'

// Query keys
export const productKeys = {
  all: ['products'] as const,
  lists: () => [...productKeys.all, 'list'] as const,
  list: (ids: Array<string>) => [...productKeys.lists(), { ids }] as const,
  details: () => [...productKeys.all, 'detail'] as const,
  detail: (id: string) => [...productKeys.details(), id] as const,
}

// Get products hook
export function useProducts(ids: Array<string>) {
  return useQuery({
    queryKey: productKeys.list(ids),
    queryFn: async () => {
      const request = create(BatchGetProductsByIDsRequestSchema, { ids })
      const response = await productApi.batchGetProductsByIDs(request)
      return response.products
    },
    enabled: ids.length > 0,
    staleTime: 1000 * 60 * 5, // 5 minutes
  })
}

// Reserve products mutation
export function useReserveProducts() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (variables: {
      idempotencyKey: string
      items: Array<{ productId: string; quantity: bigint; version: bigint }>
    }) => {
      const request = create(ReserveProductsRequestSchema, {
        idempotencyKey: variables.idempotencyKey,
        items: variables.items.map((item) => ({
          productId: item.productId,
          quantity: item.quantity,
          version: item.version,
        })),
      })
      return await productApi.reserveProducts(request)
    },
    onSuccess: (data) => {
      // Invalidate and refetch products
      queryClient.invalidateQueries({ queryKey: productKeys.all })

      // Update individual product cache if we have the data
      data.reservedProducts.forEach((product) => {
        queryClient.setQueryData(productKeys.detail(product.id), product)
      })
    },
  })
}

// Release products mutation
export function useReleaseProducts() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (variables: {
      items: Array<{ productId: string; quantity: bigint }>
    }) => {
      const request = create(ReleaseProductsRequestSchema, {
        items: variables.items.map((item) => ({
          productId: item.productId,
          quantity: item.quantity,
        })),
      })
      return await productApi.releaseProducts(request)
    },
    onSuccess: () => {
      // Invalidate products cache
      queryClient.invalidateQueries({ queryKey: productKeys.all })
    },
  })
}

// Confirm products deduction mutation
export function useConfirmProductsDeduction() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (variables: {
      items: Array<{ productId: string; quantity: bigint }>
    }) => {
      const request = create(ConfirmProductsDeductionRequestSchema, {
        items: variables.items.map((item) => ({
          productId: item.productId,
          quantity: item.quantity,
        })),
      })
      return await productApi.confirmProductsDeduction(request)
    },
    onSuccess: (data) => {
      // Invalidate and refetch products
      queryClient.invalidateQueries({ queryKey: productKeys.all })

      // Update individual product cache if we have the data
      data.updatedProducts.forEach((product) => {
        queryClient.setQueryData(productKeys.detail(product.id), product)
      })
    },
  })
}

// Restore products mutation
export function useRestoreProducts() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (variables: {
      items: Array<{ productId: string; quantity: bigint }>
      reason: string
    }) => {
      const request = create(RestoreProductsRequestSchema, {
        items: variables.items.map((item) => ({
          productId: item.productId,
          quantity: item.quantity,
        })),
        reason: variables.reason,
      })
      return await productApi.restoreProducts(request)
    },
    onSuccess: (data) => {
      // Invalidate and refetch products
      queryClient.invalidateQueries({ queryKey: productKeys.all })

      // Update individual product cache if we have the data
      data.restoredProducts.forEach((product) => {
        queryClient.setQueryData(productKeys.detail(product.id), product)
      })
    },
  })
}

// Product service health check
export function useProductHealth() {
  return useQuery({
    queryKey: ['product', 'health'],
    queryFn: async () => {
      const { HealthRequestSchema } = await import(
        '../proto/product/v1/product_pb'
      )
      const request = create(HealthRequestSchema, {})
      return await productApi.health(request)
    },
    refetchInterval: 30000, // Check every 30 seconds
  })
}

// Utility type for easier usage
export type ProductItem = Product
export type ProductReservationItem = {
  productId: string
  quantity: bigint
  version: bigint
}
export type ProductQuantityItem = {
  productId: string
  quantity: bigint
}
