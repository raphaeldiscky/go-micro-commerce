import { getAccessToken, setAccessToken } from '@/lib/api'
import type { MeQuery, RefreshTokenMutation } from '@/lib/graphql'
import {
  ME_QUERY,
  REFRESH_TOKEN_MUTATION,
  graphClient,
  handleGraphQLRequest,
  mapGraphQLUserToApiUser,
} from '@/lib/graphql'
import type { User } from '@/types/__generated__/graphql'
import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthActions {
  checkAuthStatus: () => Promise<void>
  clearError: () => void
  login: (user: User) => void
  logout: () => void
  setError: (error: null | string) => void
  setLoading: (loading: boolean) => void
  setUser: (user: null | User) => void
}

interface AuthState {
  error: null | string
  hasInitialized: boolean
  isAuthenticated: boolean
  isLoading: boolean
  user: null | User
}

type AuthStore = AuthActions & AuthState

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      checkAuthStatus: async () => {
        const state = useAuthStore.getState()

        // Prevent duplicate calls
        if (state.hasInitialized) {
          return
        }

        // Check if we have an access token in memory
        const currentToken = getAccessToken()

        // If persisted state shows we're logged out AND no token in memory,
        // skip refresh attempt (user is definitely logged out)
        if (!state.isAuthenticated && !currentToken) {
          set({
            error: null,
            hasInitialized: true,
            isAuthenticated: false,
            isLoading: false,
            user: null,
          })
          return
        }

        set({ isLoading: true })

        // If no access token in memory but persisted state shows we were authenticated,
        // try refresh token (e.g., after page refresh with valid refresh token cookie)
        if (!currentToken) {
          try {
            const refreshData = await handleGraphQLRequest(async () => {
              // Note: No refreshToken parameter needed - cookie is sent automatically
              return await graphClient.request<RefreshTokenMutation>(
                REFRESH_TOKEN_MUTATION,
                {},
              )
            }, 'Token refresh failed')

            // Store new access token in memory
            setAccessToken(refreshData.refreshToken.token)

            // Fetch user with new token
            const userData = await graphClient.request<MeQuery>(ME_QUERY)
            const user = userData.me
              ? mapGraphQLUserToApiUser(userData.me)
              : null

            set({
              error: null,
              hasInitialized: true,
              isAuthenticated: !!user,
              isLoading: false,
              user,
            })
          } catch (refreshError) {
            // Refresh failed (no cookie or cookie expired), logout
            setAccessToken(null)
            set({
              error: null,
              hasInitialized: true,
              isAuthenticated: false,
              isLoading: false,
              user: null,
            })
          }
          return
        }

        // We have a token, try to fetch user
        try {
          const data = await graphClient.request<MeQuery>(ME_QUERY)
          const user = data.me ? mapGraphQLUserToApiUser(data.me) : null
          set({
            error: null,
            hasInitialized: true,
            isAuthenticated: !!user,
            isLoading: false,
            user,
          })
        } catch (error) {
          // Access token might be expired, try to refresh
          try {
            const refreshData = await handleGraphQLRequest(async () => {
              return await graphClient.request<RefreshTokenMutation>(
                REFRESH_TOKEN_MUTATION,
                {},
              )
            }, 'Token refresh failed')

            // Store new access token in memory
            setAccessToken(refreshData.refreshToken.token)

            // Retry fetching user with new token
            const userData = await graphClient.request<MeQuery>(ME_QUERY)
            const user = userData.me
              ? mapGraphQLUserToApiUser(userData.me)
              : null

            set({
              error: null,
              hasInitialized: true,
              isAuthenticated: !!user,
              isLoading: false,
              user,
            })
          } catch (refreshError) {
            // Refresh failed, logout
            setAccessToken(null)
            set({
              error: null,
              hasInitialized: true,
              isAuthenticated: false,
              isLoading: false,
              user: null,
            })
          }
        }
      },
      clearError: () => {
        set({ error: null })
      },
      error: null,
      hasInitialized: false,
      isAuthenticated: false,

      isLoading: true,

      // Actions
      login: (user: User) => {
        set({
          error: null,
          hasInitialized: true,
          isAuthenticated: true,
          isLoading: false,
          user,
        })
      },

      logout: () => {
        // Clear access token from memory
        // Note: Server clears HTTP-only refresh token cookie
        setAccessToken(null)
        set({
          error: null,
          hasInitialized: true,
          isAuthenticated: false,
          isLoading: false,
          user: null,
        })
      },

      setError: (error: null | string) => {
        set({ error })
      },

      setLoading: (loading: boolean) => {
        set({ isLoading: loading })
      },

      setUser: (user: null | User) => {
        set({
          isAuthenticated: !!user,
          user,
        })
      },

      // Initial state
      user: null,
    }),
    {
      name: 'auth-store',
      partialize: (state) => ({
        isAuthenticated: state.isAuthenticated,
        // Only persist the user data, not loading/error/initialization states
        user: state.user,
      }),
    },
  ),
)

// Selectors for easier access
export const useUser = () => useAuthStore((state) => state.user)
export const useIsAuthenticated = () =>
  useAuthStore((state) => state.isAuthenticated)
export const useAuthLoading = () => useAuthStore((state) => state.isLoading)
export const useAuthError = () => useAuthStore((state) => state.error)
export const useAuthInitialized = () =>
  useAuthStore((state) => state.hasInitialized)
