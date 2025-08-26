package constant

// PaymentMethod represents the different payment methods available.
type PaymentMethod string

const (
	// PaymentMethodCreditCard represents the credit card payment method.
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	// PaymentMethodDebitCard represents the debit card payment method.
	PaymentMethodDebitCard PaymentMethod = "debit_card"
	// PaymentMethodPayPal represents the PayPal payment method.
	PaymentMethodPayPal PaymentMethod = "paypal"
	// PaymentMethodBankTransfer represents the bank transfer payment method.
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)
