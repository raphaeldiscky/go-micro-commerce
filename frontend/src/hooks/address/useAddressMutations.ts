import { QUERY_KEY } from '@/constants/query-key'
import {
  CREATE_ADDRESS_MUTATION,
  DELETE_ADDRESS_MUTATION,
  SET_DEFAULT_ADDRESS_MUTATION,
  UPDATE_ADDRESS_MUTATION,
} from '@/lib/graphql/address'
import type {
  CreateAddressMutation,
  DeleteAddressMutation,
  SetDefaultAddressMutation,
  UpdateAddressMutation,
} from '@/lib/graphql/address.generated'
import { graphClient } from '@/lib/graphql/client'
import type {
  CreateAddressInput,
  UpdateAddressInput,
} from '@/types/__generated__/graphql'
import { useMutation, useQueryClient } from '@tanstack/react-query'

/**
 * Hook to create a new address
 */
export function useCreateAddress() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (input: CreateAddressInput) => {
      const data = await graphClient.request<CreateAddressMutation>(
        CREATE_ADDRESS_MUTATION,
        { input },
      )
      return data.createAddress
    },
    onSuccess: () => {
      // Invalidate all address queries to refetch
      queryClient.invalidateQueries({ queryKey: QUERY_KEY.address.all })
    },
  })
}

/**
 * Hook to update an existing address
 */
export function useUpdateAddress() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      id,
      input,
    }: {
      id: string
      input: UpdateAddressInput
    }) => {
      const data = await graphClient.request<UpdateAddressMutation>(
        UPDATE_ADDRESS_MUTATION,
        { id, input },
      )
      return data.updateAddress
    },
    onSuccess: (_data, variables) => {
      // Invalidate specific address and lists
      queryClient.invalidateQueries({
        queryKey: QUERY_KEY.address.detail(variables.id),
      })
      queryClient.invalidateQueries({ queryKey: QUERY_KEY.address.lists() })
      queryClient.invalidateQueries({ queryKey: QUERY_KEY.address.default() })
    },
  })
}

/**
 * Hook to delete an address
 */
export function useDeleteAddress() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      const data = await graphClient.request<DeleteAddressMutation>(
        DELETE_ADDRESS_MUTATION,
        { id },
      )
      return data.deleteAddress
    },
    onSuccess: (_data, id) => {
      // Remove from cache and invalidate lists
      queryClient.removeQueries({ queryKey: QUERY_KEY.address.detail(id) })
      queryClient.invalidateQueries({ queryKey: QUERY_KEY.address.lists() })
    },
  })
}

/**
 * Hook to set an address as default
 */
export function useSetDefaultAddress() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (id: string) => {
      const data = await graphClient.request<SetDefaultAddressMutation>(
        SET_DEFAULT_ADDRESS_MUTATION,
        { id },
      )
      return data.setDefaultAddress
    },
    onSuccess: () => {
      // Invalidate all address queries since default status changed
      queryClient.invalidateQueries({ queryKey: QUERY_KEY.address.all })
    },
  })
}
