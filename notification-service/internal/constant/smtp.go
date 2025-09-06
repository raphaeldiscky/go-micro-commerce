package constant

const (
	// SendVerificationSubject is the subject for email verification.
	SendVerificationSubject = "Email Verification"
	// UserVerifiedSubject is the subject for User verification.
	UserVerifiedSubject = "User Verified"
)

// Email template file names.
const (
	TemplateFileOrderConfirmation   = "order_confirmation_template.html"
	TemplateFileOrderShipped        = "order_shipped_template.html"
	TemplateFileOrderCanceled       = "order_canceled_template.html"
	TemplateFilePaymentConfirmation = "payment_confirmation_template.html"
	TemplateFileGeneric             = "generic_template.html"
	TemplateFileEmailVerification   = "verification_template.html"
	TemplateFileUserVerified        = "user_verified_template.html"
)
