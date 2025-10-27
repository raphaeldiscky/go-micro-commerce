<img src="https://r2cdn.perplexity.ai/pplx-full-logo-primary-dark%402x.png" style="height:64px;margin-right:32px"/>

## Architectural Analysis: Payment Timing in Saga-Based Order Processing

The timing dilemma you've identified is a well-known challenge in distributed payment systems. Based on industry patterns and best practices, here's a comprehensive analysis of each approach.

![Comparison of Four Payment Timing Approaches](https://ppl-ai-code-interpreter-files.s3.amazonaws.com/web/direct-files/341af7e5085659e7c901b8dbf7a94353/cd981ccf-ce30-4fb7-8818-98f150a6011e/a52bf0af.png)

Comparison of Four Payment Timing Approaches

### **Recommended Approach: Option 4 (Pre-Calculate Pricing)**

**Option 4 is the strongest choice** for most e-commerce scenarios because it resolves the fundamental timing conflict by moving price calculation earlier in the workflow. Here's why:[^1][^2]

**Key Advantages:**

- **Immediate PaymentIntent Creation**: Since final amounts are known before saga initiation, the PaymentIntent can be created immediately with accurate amounts, enabling instant frontend redirect.[^3][^4]
- **Simplified Idempotency**: Idempotency is critical in payment systems. By pre-calculating prices in the cart-service, you establish a single source of truth before payment processing begins. This aligns with industry best practices where idempotency keys are tied to stable, calculated amounts rather than saga-dependent values.[^5][^6]
- **Reduced Saga Complexity**: The saga becomes simpler because SetFinalOrderPrices is no longer needed. The saga focuses on what it does best: orchestrating compensable transactions (reservations, fulfillment coordination).[^7][^8]
- **Deterministic Behavior**: Payment records can be created with certainty, reducing the risk of amount mismatches between what the customer authorizes and what is charged.[^9][^10]

**Implementation Pattern:**

```
Cart Service: Calculate and finalize prices
  ├─ Apply promotions
  ├─ Calculate shipping (synchronous call or cached rate)
  ├─ Apply taxes
  └─ Create PaymentIntent with final amount

Frontend: Immediate redirect with client secret

Saga Orchestration (now simplified):
  1. ReserveProducts
  2. CalculateShippingCost (validation only, not calculation)
  3. CreatePayment (idempotent verification)
  4. ProcessPayment
  5. CreateFulfillment
```

---

### **When to Use Each Option:**

**Option 1 (Preliminary PaymentIntent):** Use only if pricing accuracy is high. The risk is a customer seeing \$99.99 during checkout but being charged \$104.50 after saga completion—this creates customer service issues and potential disputes. Stripe technically supports updating PaymentIntent amounts, but this creates coordination problems: if the update call fails, you have a stale PaymentIntent with an incorrect amount already visible to the customer.[^11][^12]

**Option 2 (Two-Phase Payment):** This is a reasonable fallback when pre-calculation isn't possible. The major drawback is user experience: customers see a loading screen on the payment page while calculations complete. For e-commerce, every second of latency increases cart abandonment. Use this only when calculation times are consistently under 3-5 seconds.[^13]

**Option 3 (Frontend Waits):** The SSE/WebSocket approach adds infrastructure complexity without solving the core problem. Network conditions, browser limitations, and tab-switching can interrupt the connection. Additionally, if the saga takes longer than expected (>30-60 seconds), most frontend implementations have timeout issues, leaving customers in an ambiguous state.[^14][^13]

---

### **Critical Patterns to Implement:**

Regardless of which option you choose, implement these essential patterns:

**1. Idempotency Keys:** Generate a deterministic idempotency key in the cart-service based on cart contents and customer. Use this for PaymentIntent creation and payment processing. This ensures that network retries don't create duplicate charges.[^5][^6][^15]

**2. Compensating Transactions:** Even with Option 4, your saga must have compensation logic. If payment processing fails, ReserveProducts must be rolled back via inventory compensation.[^7][^8][^16]

**3. Eventual Consistency:** Accept that your system is eventually consistent. Payment status updates may lag; display provisional states ("Processing...") to customers. Once PaymentIntent.status transitions to `succeeded`, you can confidently fulfill the order.[^3][^17][^16][^13]

---

### **Implementation Recommendation:**

**Adopt Option 4 with these enhancements:**

1. **Pre-calculate all pricing in cart-service** before saga starts. This includes:
   - Base product prices (already known)
   - Shipping costs (either synchronous call to shipping-service with caching, or deterministic calculation)
   - Taxes (synchronous call to tax-service, but can be cached for common regions)
   - Promotions/discounts
2. **Create PaymentIntent immediately** after price calculation is complete, with idempotency key attached
3. **Saga becomes validation + fulfillment**:
   - ReserveProducts verifies inventory (can fail if stock changed)
   - CalculateShippingCost validates the pre-calculated rate is still valid
   - CreatePayment verifies PaymentIntent still exists and hasn't been used
   - ProcessPayment handles confirmation
   - CreateFulfillment triggers order fulfillment
4. **Handle price changes gracefully**:
   - If ReserveProducts fails due to insufficient inventory, reject the order (show customer the problem early)
   - If shipping rates changed between cart and checkout, reject and ask customer to recalculate

This approach is used by major payment processors. It treats the cart-service as the "authority" on what prices should be charged, not the saga. The saga's role is to execute the commitment, not negotiate it.[^15]

---

### **Why This Matters:**

The core insight is that **payment intents require certainty**. The payment processor (Stripe) needs an immutable amount before charging. Trying to keep amounts flexible until late in the saga creates cascading complexity: version tracking, update failures, reconciliation problems, and customer confusion. Pre-calculation moves the uncertainty to the cart phase (where it belongs) and makes the payment phase deterministic and auditable.[^3][^4]
<span style="display:none">[^18][^19][^20][^21][^22][^23][^24][^25][^26][^27][^28][^29][^30][^31][^32][^33][^34][^35][^36][^37][^38][^39][^40][^41][^42][^43][^44][^45][^46][^47][^48][^49][^50][^51][^52][^53][^54][^55]</span>

<div align="center">⁂</div>

[^1]: https://microservices.io/post/sagas/2019/12/12/developing-sagas-part-4.html

[^2]: https://strapi.io/blog/ecommerce-microservices-architecture-benefits-guide

[^3]: https://docs.stripe.com/api/payment_intents

[^4]: https://docs.stripe.com/api/payment_intents/create

[^5]: https://www.linkedin.com/pulse/designing-idempotent-microservices-avoiding-duplicate-amit-jindal-wwgcf

[^6]: https://www.youtube.com/watch?v=wYprypKHZiA\&vl=fr

[^7]: https://www.baeldung.com/cs/saga-pattern-microservices

[^8]: https://learn.microsoft.com/en-us/azure/architecture/patterns/saga

[^9]: https://www.geeksforgeeks.org/system-design/avoiding-double-payments-in-a-distributed-payments-system/

[^10]: https://microservices.io/patterns/communication-style/idempotent-consumer.html

[^11]: https://stackoverflow.com/questions/66964280/can-i-change-the-payment-intent-amount

[^12]: https://stackoverflow.com/questions/66630412/stripe-payment-intent-update-and-confirm-for-change-in-payment-amount

[^13]: https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/saga-choreography.html

[^14]: https://temporal.io/blog/mastering-saga-patterns-for-distributed-transactions-in-microservices

[^15]: https://github.com/williln/til/blob/main/stripe/payment_intents.md

[^16]: https://learn.microsoft.com/en-us/azure/architecture/patterns/compensating-transaction

[^17]: https://docs.stripe.com/payments/paymentintents/lifecycle

[^18]: https://www.designgurus.io/answers/detail/how-do-distributed-transactions-work-and-what-is-two-phase-commit-2pc

[^19]: https://stripe.dev/stripe-java/com/stripe/model/PaymentIntent.html

[^20]: https://www.youtube.com/watch?v=d2z78guUR4g

[^21]: https://martinfowler.com/articles/patterns-of-distributed-systems/two-phase-commit.html

[^22]: https://en.wikipedia.org/wiki/Two-phase_commit_protocol

[^23]: https://docs.aws.amazon.com/prescriptive-guidance/latest/modernization-data-persistence/saga-pattern.html

[^24]: https://www.youtube.com/watch?v=-_rdWB9hN1c

[^25]: https://docs.stripe.com/payments/payment-intents

[^26]: https://dev.to/jackynote/implementing-the-saga-pattern-with-spring-boot-and-activemq-in-microservice-14me

[^27]: https://relevant.software/blog/building-payment-systems-based-on-distributed-architecture/

[^28]: https://www.javadoc.io/doc/com.stripe/stripe-java/22.30.0/com/stripe/param/PaymentIntentCreateParams.html

[^29]: https://microservices.io/patterns/data/saga.html

[^30]: https://systemdr.substack.com/p/distributed-transactions-two-phase

[^31]: https://www.youtube.com/watch?v=e29KSI87vjM

[^32]: https://java-design-patterns.com/patterns/microservices-idempotent-consumer/

[^33]: https://www.reddit.com/r/stripe/comments/181i0mm/developer_best_practices_adjusting_paymentintent/

[^34]: https://www.reddit.com/r/softwaredevelopment/comments/ydp8wi/idempotency/

[^35]: https://blog.bytebytego.com/p/mastering-idempotency-building-reliable

[^36]: https://www.reddit.com/r/stripe/comments/f26wkk/paymentintent_best_practices/

[^37]: https://blogs.oracle.com/database/post/sagas-are-great-whats-the-problem

[^38]: https://blog.bitsrc.io/designing-an-idempotent-api-in-2024-d4a3cf8d8bf2

[^39]: https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/saga.html

[^40]: https://www.youtube.com/watch?v=jvAWI4MlZcQ

[^41]: https://apps.shopify.com/cart-shipping-calculator-pro

[^42]: https://docs.aws.amazon.com/prescriptive-guidance/latest/cloud-design-patterns/saga-orchestration.html

[^43]: https://www.reddit.com/r/askmath/comments/1e288al/calculating_individual_sales_prices_for_products/

[^44]: https://developer.salesforce.com/docs/commerce/salesforce-commerce/guide/cart-calculate-api.html

[^45]: https://www.reddit.com/r/learnmath/comments/lnwwyt/simple_formula_to_calculate_prices_after_discount/

[^46]: https://developer.broadleafcommerce.com/services/cart-operation-services/providers/shipping-provider

[^47]: https://github.com/cdddg/py-saga-orchestration

[^48]: https://www.calculatorsoup.com/calculators/financial/sale-price-calculator.php

[^49]: https://www.primermagazine.com/2023/learn/calculate-final-sale-prices

[^50]: https://dev.to/yoav_abrahami_736759c4edd/microservices-reliability-playbook-part-2-introduction-to-microservices-reliability-22k6

[^51]: https://docs.aws.amazon.com/prescriptive-guidance/latest/agentic-ai-patterns/saga-orchestration-patterns.html

[^52]: https://www.calculatorsoup.com/calculators/financial/sales-tax-calculator-reverse.php

[^53]: https://svitla.com/blog/microservices-for-ecommerce/

[^54]: https://www.infoq.com/articles/saga-orchestration-outbox/

[^55]: https://www.reddit.com/r/learnmath/comments/un8tj2/is_there_a_way_to_find_the_original_price_of_a/
