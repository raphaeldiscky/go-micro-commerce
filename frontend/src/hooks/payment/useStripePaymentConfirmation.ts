import { useElements, useStripe } from '@stripe/react-stripe-js'
import { useState } from 'react'
import { toast } from 'sonner'

/**
 * Hook to handle Stripe payment confirmation using Payment Element
 * Follows Stripe best practices with Payment Element (not Card Element)
 * @returns Payment confirmation handler and loading state
 */
export function useStripePaymentConfirmation() {
  const stripe = useStripe()
  const elements = useElements()
  const [isProcessing, setIsProcessing] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const confirmPayment = async (options: {
    returnUrl: string
    onSuccess?: () => void
    onError?: (error: string) => void
  }): Promise<{ success: boolean; error?: string }> => {
    if (!stripe || !elements) {
      const errorMsg = 'Stripe has not loaded yet. Please try again.'
      setError(errorMsg)
      toast.error(errorMsg)
      return { success: false, error: errorMsg }
    }

    setIsProcessing(true)
    setError(null)

    try {
      // Submit the form to validate all fields
      const { error: submitError } = await elements.submit()
      if (submitError) {
        setError(submitError.message || 'Payment validation failed')
        toast.error(submitError.message || 'Payment validation failed')
        options.onError?.(submitError.message || 'Payment validation failed')
        return {
          success: false,
          error: submitError.message || 'Payment validation failed',
        }
      }

      // Confirm the payment with Stripe
      // Stripe will handle 3DS/SCA authentication automatically
      const { error: confirmError } = await stripe.confirmPayment({
        elements,
        confirmParams: {
          return_url: options.returnUrl,
        },
        redirect: 'if_required',
      })

      if (confirmError) {
        // Payment failed or was declined
        const errorMessage =
          confirmError.message || 'Payment confirmation failed'
        setError(errorMessage)
        toast.error(errorMessage)
        options.onError?.(errorMessage)
        return { success: false, error: errorMessage }
      }

      // Payment succeeded
      toast.success('Payment confirmed successfully!')
      options.onSuccess?.()
      return { success: true }
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'An unexpected error occurred'
      setError(errorMessage)
      toast.error(errorMessage)
      options.onError?.(errorMessage)
      return { success: false, error: errorMessage }
    } finally {
      setIsProcessing(false)
    }
  }

  return {
    confirmPayment,
    isProcessing,
    error,
  }
}
