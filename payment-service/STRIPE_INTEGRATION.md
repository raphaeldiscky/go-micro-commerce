# Stripe Payment Integration Guide

This guide explains how to integrate the PCI DSS compliant payment flow with your React/Next.js frontend.

## Overview

The payment service supports **immediate payment flow** with **client-side tokenization using Stripe.js**:

- ✅ **Raw card data never touches your servers** - tokenization happens in the browser
- ✅ **SAQ A compliance** - simplest PCI DSS compliance level
- ✅ **Reduced liability** - Stripe handles sensitive payment data
- ✅ **Support for multiple payment methods** - cards, Apple Pay, Google Pay, and regional methods

## Architecture

```
┌─────────────┐          ┌─────────────┐          ┌─────────────┐
│   Browser   │          │   Backend   │          │   Stripe    │
│  (Next.js)  │          │  (Go API)   │          │     API     │
└──────┬──────┘          └──────┬──────┘          └──────┬──────┘
       │                        │                        │
       │ 1. Create Order        │                        │
       ├───────────────────────>│                        │
       │                        │                        │
       │ 2. Order Created       │                        │
       │<───────────────────────┤                        │
       │                        │                        │
       │ 3. Tokenize Card       │                        │
       ├────────────────────────┼───────────────────────>│
       │ (Stripe.js)            │                        │
       │                        │                        │
       │ 4. PaymentMethod ID    │                        │
       │<───────────────────────┼────────────────────────┤
       │ (pm_xxx)               │                        │
       │                        │                        │
       │ 5. Process Payment     │                        │
       │ (with pm_xxx)          │                        │
       ├───────────────────────>│                        │
       │                        │                        │
       │                        │ 6. Create PaymentIntent│
       │                        ├───────────────────────>│
       │                        │                        │
       │                        │ 7. client_secret       │
       │                        │<───────────────────────┤
       │                        │                        │
       │ 8. Return client_secret│                        │
       │<───────────────────────┤                        │
       │                        │                        │
       │ 9. Confirm Payment     │                        │
       │ (with client_secret)   │                        │
       ├────────────────────────┼───────────────────────>│
       │                        │                        │
       │ 10. Confirmation       │                        │
       │<───────────────────────┼────────────────────────┤
       │                        │                        │
       │                        │ 11. Webhook Event      │
       │                        │<───────────────────────┤
       │                        │ (payment succeeded)    │
       │                        │                        │
```


## Frontend Setup

### 1. Install Stripe.js

```bash
npm install @stripe/stripe-js @stripe/react-stripe-js
```

### 2. Initialize Stripe

Create a Stripe context provider:

```typescript
// lib/stripe.ts
import { loadStripe, Stripe } from "@stripe/stripe-js";

let stripePromise: Promise<Stripe | null>;

export const getStripe = () => {
  if (!stripePromise) {
    stripePromise = loadStripe(process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY!);
  }
  return stripePromise;
};
```

### 3. Wrap Your App with Elements Provider

```typescript
// pages/_app.tsx or app/layout.tsx
import { Elements } from '@stripe/react-stripe-js';
import { getStripe } from '@/lib/stripe';

function MyApp({ Component, pageProps }) {
  return (
    <Elements stripe={getStripe()}>
      <Component {...pageProps} />
    </Elements>
  );
}
```

## Payment Flow Implementation

### Option 1: Using Payment Element (Recommended)

The Payment Element is a prebuilt UI component that automatically supports multiple payment methods.

```typescript
// components/CheckoutForm.tsx
import { useState } from 'react';
import {
  useStripe,
  useElements,
  PaymentElement,
} from '@stripe/react-stripe-js';

export function CheckoutForm({ orderId, amount }) {
  const stripe = useStripe();
  const elements = useElements();
  const [isProcessing, setIsProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!stripe || !elements) {
      return;
    }

    setIsProcessing(true);
    setError(null);

    try {
      // Step 1: Submit payment element to create PaymentMethod
      const { error: submitError } = await elements.submit();
      if (submitError) {
        setError(submitError.message);
        setIsProcessing(false);
        return;
      }

      // Step 2: Create PaymentMethod and get payment_method_id
      const { error: pmError, paymentMethod } = await stripe.createPaymentMethod({
        elements,
      });

      if (pmError) {
        setError(pmError.message);
        setIsProcessing(false);
        return;
      }

      // Step 3: Send PaymentMethod ID to your backend
      const response = await fetch(`/api/orders/${orderId}/process-payment`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          payment_method_id: paymentMethod.id, // e.g., pm_xxx
          payment_method: 'card',
          idempotency_key: crypto.randomUUID(),
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Payment failed');
      }

      // Step 4: Confirm payment with client_secret (handles 3DS if needed)
      if (data.client_secret) {
        const { error: confirmError, paymentIntent } = await stripe.confirmPayment({
          clientSecret: data.client_secret,
          confirmParams: {
            return_url: `${window.location.origin}/orders/${orderId}/confirmation`,
          },
          redirect: 'if_required', // Only redirect if 3DS is required
        });

        if (confirmError) {
          setError(confirmError.message);
          setIsProcessing(false);
          return;
        }

        // Payment succeeded!
        if (paymentIntent?.status === 'succeeded') {
          window.location.href = `/orders/${orderId}/confirmation`;
        }
      }
    } catch (err) {
      setError(err.message);
      setIsProcessing(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <PaymentElement />

      {error && (
        <div className="error">{error}</div>
      )}

      <button type="submit" disabled={!stripe || isProcessing}>
        {isProcessing ? 'Processing...' : `Pay $${amount}`}
      </button>
    </form>
  );
}
```

### Option 2: Using Card Element (Card-only)

If you only need to accept cards:

```typescript
// components/CardPaymentForm.tsx
import { useState } from 'react';
import {
  useStripe,
  useElements,
  CardElement,
} from '@stripe/react-stripe-js';

const CARD_ELEMENT_OPTIONS = {
  style: {
    base: {
      fontSize: '16px',
      color: '#424770',
      '::placeholder': {
        color: '#aab7c4',
      },
    },
    invalid: {
      color: '#9e2146',
    },
  },
};

export function CardPaymentForm({ orderId, amount, customerEmail }) {
  const stripe = useStripe();
  const elements = useElements();
  const [isProcessing, setIsProcessing] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!stripe || !elements) {
      return;
    }

    setIsProcessing(true);
    setError(null);

    const cardElement = elements.getElement(CardElement);

    if (!cardElement) {
      setError('Card element not found');
      setIsProcessing(false);
      return;
    }

    try {
      // Step 1: Create PaymentMethod from card data (tokenization)
      const { error: pmError, paymentMethod } = await stripe.createPaymentMethod({
        type: 'card',
        card: cardElement,
        billing_details: {
          email: customerEmail,
        },
      });

      if (pmError) {
        setError(pmError.message);
        setIsProcessing(false);
        return;
      }

      // Step 2: Send PaymentMethod ID to backend
      const response = await fetch(`/api/orders/${orderId}/process-payment`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          payment_method_id: paymentMethod.id,
          payment_method: 'card',
          customer_email: customerEmail,
          idempotency_key: crypto.randomUUID(),
        }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.message || 'Payment failed');
      }

      // Step 3: Handle client-side confirmation (3DS if required)
      if (data.client_secret) {
        const { error: confirmError, paymentIntent } = await stripe.confirmCardPayment(
          data.client_secret,
          {
            payment_method: paymentMethod.id,
          }
        );

        if (confirmError) {
          setError(confirmError.message);
          setIsProcessing(false);
          return;
        }

        // Payment succeeded!
        if (paymentIntent?.status === 'succeeded') {
          window.location.href = `/orders/${orderId}/confirmation`;
        }
      }
    } catch (err) {
      setError(err.message);
      setIsProcessing(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <div className="card-element">
        <CardElement options={CARD_ELEMENT_OPTIONS} />
      </div>

      {error && (
        <div className="error">{error}</div>
      )}

      <button type="submit" disabled={!stripe || isProcessing}>
        {isProcessing ? 'Processing...' : `Pay $${amount}`}
      </button>
    </form>
  );
}
```

## API Integration

### Backend API Endpoint

The payment service now expects `payment_method_id` instead of raw card data:

```typescript
// Your Next.js API route or client-side API call
interface ProcessPaymentRequest {
  payment_method_id: string; // Required: PM ID from Stripe (pm_xxx)
  payment_method: "card"; // Required: Payment method type
  customer_email: string; // Optional: For receipts
  idempotency_key: string; // Required: UUID for idempotency
}

interface ProcessPaymentResponse {
  id: string; // Payment ID
  order_id: string; // Order ID
  status: string; // 'processing' or 'completed'
  amount: number; // Amount in currency units
  currency: string; // e.g., 'usd'
  client_secret?: string; // For client-side confirmation
  requires_action?: boolean; // True if 3DS required
  next_action_type?: string; // Type of action required
}

// Example usage
const response = await fetch(
  `${process.env.NEXT_PUBLIC_API_URL}/orders/${orderId}/process-payment`,
  {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify({
      payment_method_id: paymentMethodId,
      payment_method: "card",
      customer_email: user.email,
      idempotency_key: crypto.randomUUID(),
    }),
  }
);
```

## Handling Payment Statuses

### Payment Lifecycle

1. **pending** - Payment record created, awaiting for user payment (order.status.payment_pending)
2. **processing** - Payment submitted to gateway, awaiting confirmation (order.status.payment_pending)
3. **completed** - Payment confirmed by Stripe webhook (order.status.paid)
4. **failed** - Payment failed (card declined, etc.) (order.status.payment_failed)
5. **timeout** - Payment took too long to process (order.status.payment_expired)

### Polling for Payment Status

Since webhooks update the payment status asynchronously, you may want to poll:

```typescript
async function pollPaymentStatus(orderId: string) {
  const maxAttempts = 30;
  const interval = 1000; // 1 second

  for (let i = 0; i < maxAttempts; i++) {
    const response = await fetch(
      `${process.env.NEXT_PUBLIC_API_URL}/orders/${orderId}/payment`
    );
    const payment = await response.json();

    if (payment.status === "completed") {
      return { success: true, payment };
    }

    if (payment.status === "failed") {
      return { success: false, payment };
    }

    await new Promise((resolve) => setTimeout(resolve, interval));
  }

  return { success: false, timeout: true };
}
```

## 3D Secure (SCA) Handling

The implementation automatically handles 3D Secure authentication:

```typescript
// Using confirmCardPayment or confirmPayment will trigger 3DS modal if needed
const { error, paymentIntent } = await stripe.confirmCardPayment(clientSecret, {
  payment_method: paymentMethodId,
});

// Stripe.js automatically:
// 1. Detects if 3DS is required
// 2. Opens authentication modal
// 3. Handles challenge flow
// 4. Returns result when complete
```

## Digital Wallets

The Payment Element automatically supports Apple Pay, Google Pay, and Link:

```typescript
// No additional code needed - Payment Element handles this automatically!
// Just ensure your Stripe account has wallets enabled and domain verified.

<Elements stripe={stripePromise}>
  <PaymentElement />
  {/* Apple Pay, Google Pay buttons appear automatically if available */}
</Elements>
```

## Error Handling

### Common Errors

```typescript
interface StripeError {
  type: string;
  code?: string;
  message: string;
}

// Handle different error types
function handlePaymentError(error: StripeError) {
  switch (error.code) {
    case "card_declined":
      return "Your card was declined. Please try another payment method.";
    case "insufficient_funds":
      return "Insufficient funds. Please try another card.";
    case "expired_card":
      return "Your card has expired. Please use a different card.";
    case "incorrect_cvc":
      return "Incorrect CVC. Please check and try again.";
    case "processing_error":
      return "An error occurred while processing your card. Please try again.";
    default:
      return error.message || "An unexpected error occurred. Please try again.";
  }
}
```

## Testing

### Test Cards

Use these test cards in development:

```text
Success: 4242 4242 4242 4242
Decline: 4000 0000 0000 0002
3DS Required: 4000 0025 0000 3155
Insufficient Funds: 4000 0000 0000 9995

Expiry: Any future date (e.g., 12/34)
CVC: Any 3 digits (e.g., 123)
ZIP: Any 5 digits (e.g., 12345)
```

### Testing Webhooks Locally

Use Stripe CLI to forward webhooks to your local server:

```bash
# Install Stripe CLI
brew install stripe/stripe-cli/stripe

# Login
stripe login

# Forward webhooks
stripe listen --forward-to localhost:8080/webhooks/stripe

# The CLI will output your webhook signing secret:
# whsec_xxx - add this to your .env
```

## Security Best Practices

1. ✅ **Never log PaymentMethod IDs** - they're sensitive
2. ✅ **Always verify webhook signatures** - prevent tampering
3. ✅ **Use HTTPS in production** - required by Stripe
4. ✅ **Implement idempotency** - use `idempotency_key` for retries
5. ✅ **Rate limit payment endpoints** - prevent abuse
6. ✅ **Validate amounts server-side** - never trust client amounts

## Environment Variables

### Frontend (.env.local)

```bash
NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_test_xxx
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Backend (.env)

```bash
STRIPE_API_KEY=sk_test_xxx
STRIPE_WEBHOOK_SECRET=whsec_xxx
```

## Troubleshooting

### Issue: "client_secret not returned"

**Cause**: Payment gateway might have immediately succeeded (rare)

**Solution**: Check `status` field - if `completed`, payment succeeded without confirmation

### Issue: "Webhook signature verification failed"

**Cause**: Wrong webhook secret or payload tampering

**Solution**:

1. Check webhook secret in config
2. Ensure raw request body is passed to verification
3. Check Stripe dashboard for webhook logs

### Issue: "Payment stuck in processing"

**Cause**: Webhook not received or failed to process

**Solution**:

1. Check webhook endpoint is accessible
2. Review webhook logs in Stripe dashboard
3. Implement polling as fallback

### Issue: "3DS modal not appearing"

**Cause**: Incorrect confirmCardPayment implementation

**Solution**: Ensure you're passing `clientSecret` correctly and not using `redirect: 'never'`

## Additional Resources

- [Stripe.js Documentation](https://stripe.com/docs/js)
- [Payment Intents API](https://stripe.com/docs/payments/payment-intents)
- [3D Secure Guide](https://stripe.com/docs/payments/3d-secure)
- [Testing Stripe](https://stripe.com/docs/testing)
- [PCI Compliance](https://stripe.com/docs/security/guide)

## Support

For issues or questions:

1. Check Stripe dashboard for payment/webhook logs
2. Review backend logs for errors
3. Use Stripe CLI for webhook debugging
4. Contact your backend team for API issues
