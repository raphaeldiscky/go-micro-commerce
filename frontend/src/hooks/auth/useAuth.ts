import { QUERY_KEY } from '@/constants/query-key'
import { PATH_ROOT } from '@/constants/routes'
import { setAccessToken } from '@/lib/api/client'
import type {
  LoginMutation,
  LoginMutationVariables,
  LogoutMutation,
  RegisterMutation,
  RegisterMutationVariables,
} from '@/lib/graphql'
import {
  LOGIN_MUTATION,
  LOGOUT_MUTATION,
  REGISTER_MUTATION,
  graphqlClient,
  handleGraphQLRequest,
  mapGraphQLUserToApiUser,
} from '@/lib/graphql'
import { useAuthStore } from '@/store/authStore'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useRouter } from '@tanstack/react-router'

/**
 * Hook for getting current user profile
 * Note: This hook is disabled to prevent duplicate requests during initialization
 * The auth store handles user fetching during checkAuthStatus
 */
export function useCurrentUser() {
  const user = useAuthStore((state) => state.user)
  const isLoading = useAuthStore((state) => state.isLoading)
  const hasInitialized = useAuthStore((state) => state.hasInitialized)

  // Return mock query result to match TanStack Query interface
  return {
    data: user,
    error: null,
    isError: false,
    isLoading: !hasInitialized && isLoading,
  }
}

/**
 * Hook for user login using GraphQL
 */
export function useLogin() {
  const queryClient = useQueryClient()
  const router = useRouter()
  const loginUser = useAuthStore((state) => state.login)

  return useMutation({
    mutationFn: async (input: LoginMutationVariables['input']) => {
      return handleGraphQLRequest(async () => {
        const data = await graphqlClient.request<
          LoginMutation,
          LoginMutationVariables
        >(LOGIN_MUTATION, { input })
        return data.login
      }, 'Login failed')
    },
    onError: (error) => {
      console.error('Login failed:', error)
    },
    onSuccess: (data) => {
      // Store access token in memory (refresh token is in HTTP-only cookie)
      setAccessToken(data.token)

      // Map GraphQL user to API user format
      const user = mapGraphQLUserToApiUser(data.user)

      // Update auth store with user data
      loginUser(user)

      // Update React Query cache
      queryClient.setQueryData(QUERY_KEY.auth.currentUser(), user)

      // Invalidate all auth-related queries
      queryClient.invalidateQueries({ queryKey: QUERY_KEY.auth.all })

      // Navigate to home or intended destination
      router.navigate({ to: PATH_ROOT.home })
    },
  })
}

/**
 * Hook for user logout using GraphQL
 */
export function useLogout() {
  const queryClient = useQueryClient()
  const router = useRouter()
  const logoutUser = useAuthStore((state) => state.logout)

  return useMutation({
    mutationFn: async () => {
      return handleGraphQLRequest(async () => {
        const data =
          await graphqlClient.request<LogoutMutation>(LOGOUT_MUTATION)
        console.log('Logout API response:', data)
        return data
      }, 'Logout request failed')
    },
    onSuccess: () => {
      console.log('Logout successful, cleaning up auth state...')

      // Clear access token from memory
      setAccessToken(null)

      // Update auth store (sets user to null, isAuthenticated to false)
      logoutUser()

      // Clear all React Query cache
      queryClient.clear()

      console.log('Auth state cleared, navigating to home...')

      // Navigate to home page
      try {
        router.navigate({ to: PATH_ROOT.home })
        console.log('Navigation complete')
      } catch (navError) {
        console.error('Navigation failed:', navError)
        // Force reload as fallback
        window.location.href = PATH_ROOT.home
      }
    },
    onError: (error) => {
      console.error('Logout API failed:', error)

      // Even if server logout fails, clear client-side state
      setAccessToken(null)
      logoutUser()
      queryClient.clear()

      // Navigate to home even on error
      try {
        router.navigate({ to: PATH_ROOT.home })
      } catch (navError) {
        console.error('Navigation failed:', navError)
        window.location.href = PATH_ROOT.home
      }
    },
  })
}

/**
 * Hook for user registration using GraphQL
 */
export function useRegister() {
  const queryClient = useQueryClient()
  const router = useRouter()
  const loginUser = useAuthStore((state) => state.login)

  return useMutation({
    mutationFn: async (input: RegisterMutationVariables['input']) => {
      return handleGraphQLRequest(async () => {
        const data = await graphqlClient.request<
          RegisterMutation,
          RegisterMutationVariables
        >(REGISTER_MUTATION, { input })
        return data.register
      }, 'Registration failed')
    },
    onError: (error) => {
      console.error('Registration failed:', error)
    },
    onSuccess: (data) => {
      // Store access token in memory (refresh token is in HTTP-only cookie)
      setAccessToken(data.token)

      // Map GraphQL user to API user format
      const user = mapGraphQLUserToApiUser(data.user)

      // Update auth store with user data
      loginUser(user)

      // Update React Query cache
      queryClient.setQueryData(QUERY_KEY.auth.currentUser(), user)

      // Invalidate all auth-related queries
      queryClient.invalidateQueries({ queryKey: QUERY_KEY.auth.all })

      // Navigate to home
      router.navigate({ to: PATH_ROOT.home })
    },
  })
}

// Re-export store selectors for convenience
export {
  useAuthError,
  useAuthInitialized,
  useAuthLoading,
  useIsAuthenticated,
  useUser,
} from '@/store/authStore'
