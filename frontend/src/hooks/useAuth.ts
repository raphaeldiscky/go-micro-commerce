import type { AuthResponse, LoginRequest, RegisterRequest } from '@/lib/api'
import { login, logout, register } from '@/lib/api'
import { useAuthStore } from '@/store/authStore'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useRouter } from '@tanstack/react-router'

export const AUTH_QUERY_KEYS = {
  all: ['auth'] as const,
  currentUser: ['auth', 'currentUser'] as const,
} as const

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
 * Hook for user login
 */
export function useLogin() {
  const queryClient = useQueryClient()
  const router = useRouter()
  const loginUser = useAuthStore((state) => state.login)

  return useMutation({
    mutationFn: (credentials: LoginRequest) => login(credentials),
    onError: (error) => {
      console.error('Login failed:', error)
    },
    onSuccess: (data: AuthResponse) => {
      // Update auth store
      loginUser(data.user)

      // Update React Query cache
      queryClient.setQueryData(AUTH_QUERY_KEYS.currentUser, data.user)

      // Invalidate all auth-related queries
      queryClient.invalidateQueries({ queryKey: AUTH_QUERY_KEYS.all })

      // Navigate to home or intended destination
      router.navigate({ to: '/' })
    },
  })
}

/**
 * Hook for user logout
 */
export function useLogout() {
  const queryClient = useQueryClient()
  const router = useRouter()
  const logoutUser = useAuthStore((state) => state.logout)

  return useMutation({
    mutationFn: logout,
    onSettled: () => {
      // Always clear auth state even if API call fails
      logoutUser()
      queryClient.removeQueries({ queryKey: AUTH_QUERY_KEYS.all })
    },
    onSuccess: () => {
      // Clear auth store
      logoutUser()

      // Clear all React Query cache
      queryClient.clear()

      // Navigate to home
      router.navigate({ to: '/' })
    },
  })
}

/**
 * Hook for user registration
 */
export function useRegister() {
  const queryClient = useQueryClient()
  const router = useRouter()
  const loginUser = useAuthStore((state) => state.login)

  return useMutation({
    mutationFn: (userData: RegisterRequest) => register(userData),
    onError: (error) => {
      console.error('Registration failed:', error)
    },
    onSuccess: (data: AuthResponse) => {
      // Update auth store
      loginUser(data.user)

      // Update React Query cache
      queryClient.setQueryData(AUTH_QUERY_KEYS.currentUser, data.user)

      // Invalidate all auth-related queries
      queryClient.invalidateQueries({ queryKey: AUTH_QUERY_KEYS.all })

      // Navigate to home
      router.navigate({ to: '/' })
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
