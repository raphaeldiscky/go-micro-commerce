import {
  useAuthInitialized,
  useAuthLoading,
  useAuthStore,
} from '@/store/authStore'
import { useCartStore } from '@/store/cartStore'
import { useEffect, useRef } from 'react'

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const checkAuthStatus = useAuthStore((state) => state.checkAuthStatus)
  const user = useAuthStore((state) => state.user)
  const hasInitialized = useAuthInitialized()
  const isLoading = useAuthLoading()
  const initializeCart = useCartStore((state) => state.initializeCart)
  const hasChecked = useRef(false)
  const cartInitialized = useRef(false)

  useEffect(() => {
    if (!hasChecked.current) {
      hasChecked.current = true
      checkAuthStatus()
    }
  }, [checkAuthStatus])

  // Initialize cart when user is authenticated and cart hasn't been initialized yet
  useEffect(() => {
    if (user && !cartInitialized.current) {
      cartInitialized.current = true
      initializeCart(user.id)
      console.log('Cart initialized for user:', user.id)
    }
  }, [user, initializeCart])

  // Show loading state while auth is being initialized
  // This prevents queries from executing before the access token is available
  if (!hasInitialized || isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-50 to-white dark:from-gray-900 dark:to-gray-800">
        <div className="text-center">
          <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-current border-r-transparent motion-reduce:animate-[spin_1.5s_linear_infinite]" />
          <p className="mt-4 text-sm text-gray-600 dark:text-gray-400">
            Loading...
          </p>
        </div>
      </div>
    )
  }

  return <>{children}</>
}
