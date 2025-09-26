import type { User } from '@/lib/api'
import { getAccessToken, getCurrentUser, setAccessToken } from '@/lib/api'
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

        const token = getAccessToken()

        if (!token) {
          set({
            hasInitialized: true,
            isLoading: false,
          })
          return
        }

        try {
          set({ isLoading: true })
          const user = await getCurrentUser()
          set({
            error: null,
            hasInitialized: true,
            isAuthenticated: true,
            isLoading: false,
            user,
          })
        } catch (error) {
          // Token might be expired or invalid
          setAccessToken(null)
          set({
            error: null,
            hasInitialized: true,
            isAuthenticated: false,
            isLoading: false,
            user: null,
          })
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
          isAuthenticated: true,
          isLoading: false,
          user,
        })
      },

      logout: () => {
        setAccessToken(null)
        set({
          error: null,
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
