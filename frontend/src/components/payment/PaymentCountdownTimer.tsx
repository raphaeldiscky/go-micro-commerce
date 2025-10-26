import { Alert, AlertDescription } from '@/components/ui/alert'
import { differenceInSeconds, formatDistance } from 'date-fns'
import { Clock } from 'lucide-react'
import { useEffect, useState } from 'react'

interface PaymentCountdownTimerProps {
  expiresAt: string
  onExpired?: () => void
}

/**
 * Countdown timer showing time remaining until payment expires
 * Shows warning when less than 1 hour remaining
 * Automatically calls onExpired callback when time runs out
 */
export function PaymentCountdownTimer({
  expiresAt,
  onExpired,
}: PaymentCountdownTimerProps) {
  const [timeRemaining, setTimeRemaining] = useState<number>(0)
  const [isExpired, setIsExpired] = useState(false)

  useEffect(() => {
    const calculateTimeRemaining = () => {
      const expiryDate = new Date(expiresAt)
      const now = new Date()
      const seconds = differenceInSeconds(expiryDate, now)

      if (seconds <= 0) {
        setIsExpired(true)
        setTimeRemaining(0)
        onExpired?.()
      } else {
        setTimeRemaining(seconds)
      }
    }

    // Initial calculation
    calculateTimeRemaining()

    // Update every second
    const interval = setInterval(calculateTimeRemaining, 1000)

    return () => clearInterval(interval)
  }, [expiresAt, onExpired])

  // Calculate hours, minutes, seconds
  const hours = Math.floor(timeRemaining / 3600)
  const minutes = Math.floor((timeRemaining % 3600) / 60)
  const seconds = timeRemaining % 60

  // Show warning when less than 1 hour remaining
  const isLowTime = hours === 0 && minutes < 60

  if (isExpired) {
    return (
      <Alert variant="destructive">
        <Clock className="h-4 w-4" />
        <AlertDescription>
          Payment window has expired. Please create a new order.
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <Alert variant={isLowTime ? 'default' : 'default'}>
      <Clock className="h-4 w-4" />
      <AlertDescription>
        {isLowTime ? (
          <span className="font-semibold text-orange-600">
            Payment expires in {minutes}m {seconds}s
          </span>
        ) : (
          <span>
            Payment expires{' '}
            {formatDistance(new Date(expiresAt), new Date(), {
              addSuffix: true,
            })}
          </span>
        )}
      </AlertDescription>
    </Alert>
  )
}
