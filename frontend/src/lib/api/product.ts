import { env } from '@/env'
import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import type {
  ConfirmProductsDeductionRequest,
  HealthRequest,
  ReleaseProductsRequest,
  ReserveProductsRequest,
  RestoreProductsRequest,
} from '../../proto/product/v1/product_pb'
import { ProductService } from '../../proto/product/v1/product_pb'
import { getAccessToken } from './client'

// Create transport with auth interceptor
const transport = createConnectTransport({
  baseUrl: env.VITE_API_GATEWAY_URL,
  interceptors: [
    (next) => async (req) => {
      const token = getAccessToken()
      if (token) {
        req.header.set('Authorization', `Bearer ${token}`)
      }
      return await next(req)
    },
  ],
})

// Create Connect client using generated service
const client = createClient(ProductService, transport)

// Product API functions
export const productApi = {
  listProducts: (limit?: string, nextCursor?: string) =>
    client.listProducts({ limit, nextCursor }),
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
