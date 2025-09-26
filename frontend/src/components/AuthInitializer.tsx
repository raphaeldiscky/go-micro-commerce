import { useAuthStore } from '@/store/authStore'
import { useEffect, useRef } from 'react'

export default function AuthInitializer({
  children,
}: {
  children: React.ReactNode
}) {
  const checkAuthStatus = useAuthStore((state) => state.checkAuthStatus)
  const hasInitialized = useRef(false)

  useEffect(() => {
    if (!hasInitialized.current) {
      hasInitialized.current = true
      checkAuthStatus()
    }
  }, [checkAuthStatus])

  return <>{children}</>
}
