Here’s your **clean version** of the document — all grid lines and borders have been removed while keeping the structure and readability intact:

---

## **24-Hour Payment Window with Stripe Integration — Implementation Plan**

### **Executive Summary**

Based on your requirements and current architecture, this plan implements a deferred payment flow where:

1. **Checkout → Place Order:** Creates order with `payment_pending` status
2. **Redirect to Payment Page:** User has 24 hours to complete payment
3. **Stripe Payment Element:** Collects payment using modern, PCI-compliant approach
4. **Payment Options:** Supports cards, wallets, bank transfers, and saves payment methods
5. **Timeout & Reminders:** Automated order cancellation and reminder notifications

---

### **Features to Implement**

#### **1. Payment Page Route (Frontend)**

- New route: `/payment/:orderId`
- Loads Stripe Payment Element with `client_secret`
- Displays order summary and countdown timer
- Handles payment confirmation and redirects

#### **2. Payment Intent Creation (Backend)**

- Modify `placeOrder` flow to create Stripe PaymentIntent
- Return `client_secret` to frontend for payment collection
- Store PaymentIntent ID in payment record

#### **3. Payment Timeout Job (Backend)**

- Scheduled job checks for payments older than 24h with `status=pending`
- Cancels expired payments and orders
- Publishes timeout events

#### **4. Payment Reminder Notifications (Backend)**

- Send reminders at 12h and 23h marks
- Email/SMS notifications with payment link
- Integration with `notification-service`

#### **5. Payment Method Retry (Frontend)**

- Allow users to retry failed payments
- Support changing payment methods within 24h window
- Real-time payment status updates

#### **6. Saved Payment Methods (Backend + Frontend)**

- Create/attach Stripe Customer on first payment
- Save PaymentMethod for future use
- Display saved payment methods in checkout

#### **7. Order Status Page (Frontend)**

- New route: `/orders/:orderId`
- Real-time payment status tracking
- Link to payment page if still pending

---

### **Complete Data Flow**

#### **Phase 1: Checkout → Place Order**

**Frontend Request**

```graphql
mutation PlaceOrder($sessionId: UUID!, $input: PlaceOrderInput!) {
  placeOrder(sessionId: $sessionId, input: $input) {
    id
    status
    payment {
      id
      status
      clientSecret
      expiresAt
    }
    order {
      id
      status
      totalPrice
    }
  }
}
```

**Backend Flow (cart-service)**

1. Validate checkout session
2. Publish `CheckoutSessionOrderPlacedEvent` to Kafka
3. Return response immediately

**Backend Flow (order-service Saga Consumer)**

1. Reserve products (`product-service` gRPC)
2. Get shipping cost (`fulfillment-service` gRPC)
3. Set final prices
4. Create payment record (`payment-service` HTTP)
5. Create Stripe PaymentIntent
6. Update order status to `payment_pending`
7. Publish `OrderCreatedEvent`

**Backend Response (payment-service)**

```json
{
  "payment": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "orderId": "660e8400-e29b-41d4-a716-446655440111",
    "amount": 150.0,
    "currency": "USD",
    "status": "pending",
    "clientSecret": "pi_3AB...xyz_secret_...",
    "expiresAt": "2025-10-25T12:00:00Z"
  },
  "order": {
    "id": "660e8400-e29b-41d4-a716-446655440111",
    "status": "payment_pending",
    "totalPrice": 150.0
  }
}
```

---

#### **Phase 2: Payment Page Load**

**Frontend Query**

```graphql
query GetOrderPayment($orderId: UUID!) {
  order(id: $orderId) {
    id
    totalPrice
    currency
    status
    items {
      productName
      quantity
      unitPrice
    }
  }
  payment(orderId: $orderId) {
    id
    status
    clientSecret
    expiresAt
    stripeCustomerID
    savedPaymentMethods {
      id
      type
      last4
      brand
    }
  }
}
```

**Stripe.js Integration**

```tsx
import { loadStripe } from '@stripe/stripe-js'
import { Elements, PaymentElement } from '@stripe/react-stripe-js'

const stripe = await loadStripe(STRIPE_PUBLISHABLE_KEY)

<Elements stripe={stripe} options={{ clientSecret }}>
  <PaymentElement />
  <button onClick={handleSubmit}>Pay Now</button>
</Elements>
```

---

#### **Phase 3: Payment Confirmation**

**Frontend**

```ts
const { error } = await stripe.confirmPayment({
  elements,
  confirmParams: {
    return_url: `${window.location.origin}/orders/${orderId}`,
  },
});
```

**Stripe Webhook**

```json
{
  "type": "payment_intent.succeeded",
  "data": {
    "object": {
      "id": "pi_3AB...xyz",
      "status": "succeeded",
      "metadata": {
        "transaction_id": "550e8400-e29b-41d4-a716-446655440000",
        "order_id": "660e8400-e29b-41d4-a716-446655440111"
      }
    }
  }
}
```

**Backend Processing**

1. Verify Stripe signature
2. Update payment status to `completed`
3. Publish `PaymentCompletedEvent`
4. Order service updates order to `paid`
5. Trigger fulfillment workflow

---

#### **Phase 4: Payment Timeout (24h Job)**

**Backend Job**

```go
// Scheduled every 5 minutes
func (j *paymentTimeoutJob) Execute(ctx context.Context) error {
  expiredPayments := paymentRepo.FindExpiredPayments(ctx, 24*time.Hour)
  for _, payment := range expiredPayments {
    stripe.PaymentIntent.Cancel(payment.GatewayReferenceID)
    paymentService.TimeoutPayment(ctx, payment.OrderID)
    publishPaymentTimeoutEvent(payment)
  }
}
```

**Order Service Consumer**

1. Update order status to `payment_expired`
2. Release reserved inventory
3. Publish `OrderExpiredEvent`
4. Notify customer

---

#### **Phase 5: Payment Reminders**

**Reminder Job**

```go
// Scheduled every 15 minutes
func checkPaymentReminders(ctx context.Context) {
  twelveHourPayments := findPaymentsPendingFor(12 * time.Hour)
  for _, p := range twelveHourPayments {
    sendReminderEmail(p.CustomerID, p.OrderID, "12h")
  }

  finalReminderPayments := findPaymentsPendingFor(23 * time.Hour)
  for _, p := range finalReminderPayments {
    sendReminderEmail(p.CustomerID, p.OrderID, "1h")
  }
}
```

---

### **Implementation Tasks**

#### **Backend**

- Payment service: add APIs, Stripe integration, timeout job, retries, saved payment methods
- Order service: handle `PaymentTimeoutEvent`, update saga, public order endpoint
- Cart service: extend `placeOrder` mutation
- Notification service: implement reminder jobs and templates
- GraphQL gateway: expose order/payment queries

#### **Frontend**

- `/payment/:orderId` route with Stripe integration, retry support, countdown
- `/orders/:orderId` page with live status tracking
- Checkout flow redirect to payment page

#### **Database**

```sql
ALTER TABLE payments ADD COLUMN expires_at TIMESTAMPTZ;
CREATE INDEX idx_payments_expires_at ON payments (expires_at, status);
```

---

### **Example Requests**

**Place Order**

```js
const result = await placeOrder({
  sessionId: checkoutSessionId,
  input: {
    addressId: "addr-123",
    carrierId: "fedex",
    paymentMethod: "CARD",
    paymentGateway: "STRIPE",
    idempotencyKey: "unique-key-123",
  },
});
```

**Response**

```json
{
  "data": {
    "placeOrder": {
      "id": "session-123",
      "status": "order_placed",
      "payment": {
        "id": "pay-123",
        "status": "pending",
        "clientSecret": "pi_3AB...xyz_secret_...",
        "expiresAt": "2025-10-25T12:00:00Z"
      },
      "order": {
        "id": "order-456",
        "status": "payment_pending",
        "totalPrice": 150.0
      }
    }
  }
}
```

---

### **Stripe Best Practices Applied**

✅ PaymentIntent API (not legacy Charges API)
✅ Payment Element (modern PCI-compliant UI)
✅ Client-side confirmation
✅ Dynamic payment methods
✅ Idempotency keys for safe retries
✅ Verified webhooks
✅ Off-session charging for saved methods
✅ Metadata linking internal IDs

---

### **Timeline Estimate**

- Backend implementation: 3–4 days
- Frontend implementation: 2–3 days
- Testing & integration: 2 days
  **→ Total: ~7–9 days**

---

Would you like me to reformat this into a **Markdown file** or **Notion-style document** (for clearer sharing with your team)?
