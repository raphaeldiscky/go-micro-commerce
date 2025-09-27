import { env } from '@/env'
import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import type {
  BatchGetProductsByIDsRequest,
  ConfirmProductsDeductionRequest,
  HealthRequest,
  ReleaseProductsRequest,
  ReserveProductsRequest,
  RestoreProductsRequest,
} from '../../proto/product/v1/product_pb'
import { ProductService } from '../../proto/product/v1/product_pb'

// Create transport
const transport = createConnectTransport({
  baseUrl: env.VITE_API_GATEWAY_URL,
})

// Create Connect client using generated service
const client = createClient(ProductService, transport)

// Product API functions
export const productApi = {
  batchGetProductsByIDs: (request: BatchGetProductsByIDsRequest) =>
    client.batchGetProductsByIDs(request),
  reserveProducts: (request: ReserveProductsRequest) =>
    client.reserveProducts(request),
  releaseProducts: (request: ReleaseProductsRequest) =>
    client.releaseProducts(request),
  confirmProductsDeduction: (request: ConfirmProductsDeductionRequest) =>
    client.confirmProductsDeduction(request),
  restoreProducts: (request: RestoreProductsRequest) =>
    client.restoreProducts(request),
  health: (request: HealthRequest) => client.health(request),
} as const

export type ProductApi = typeof productApi
