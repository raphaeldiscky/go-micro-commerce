import { QUERY_KEY } from '@/constants/query-key'
import {
  GET_ADDRESS_QUERY,
  GET_DEFAULT_ADDRESS_QUERY,
  LIST_ADDRESSES_QUERY,
} from '@/lib/graphql/address'
import type {
  GetAddressQuery,
  GetDefaultAddressQuery,
  ListAddressesQuery,
} from '@/lib/graphql/address.generated'
import { graphClient } from '@/lib/graphql/client'
import { useQuery } from '@tanstack/react-query'

/**
 * Hook to list user addresses with cursor pagination
 */
export function useAddresses(limit: number, cursor?: string) {
  return useQuery({
    queryKey: QUERY_KEY.address.list(limit, cursor),
    queryFn: async () => {
      const data = await graphClient.request<ListAddressesQuery>(
        LIST_ADDRESSES_QUERY,
        { limit, cursor },
      )
      return data.listAddresses
    },
  })
}

/**
 * Hook to get a single address by ID
 */
export function useAddress(id: string) {
  return useQuery({
    queryKey: QUERY_KEY.address.detail(id),
    queryFn: async () => {
      const data = await graphClient.request<GetAddressQuery>(
        GET_ADDRESS_QUERY,
        { id },
      )
      return data.getAddress
    },
    enabled: !!id,
  })
}

/**
 * Hook to get the default address
 */
export function useDefaultAddress() {
  return useQuery({
    queryKey: QUERY_KEY.address.default(),
    queryFn: async () => {
      const data = await graphClient.request<GetDefaultAddressQuery>(
        GET_DEFAULT_ADDRESS_QUERY,
      )
      return data.getDefaultAddress
    },
  })
}
