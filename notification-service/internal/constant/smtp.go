package constant

const (
	// SMTPPort is the default SMTP port.
	SMTPPort = 1025
)

const (
	// SendVerificationSubject is the subject for email verification.
	SendVerificationSubject = "Email Verification"
	// UserVerifiedSubject is the subject for User verification.
	UserVerifiedSubject = "User Verified"
)

// Email template file names.
const (
	TemplateFileOrderConfirmed       = "order_confirmed_template.html"
	TemplateFileOrderShipped         = "order_shipped_template.html"
	TemplateFileOrderCanceled        = "order_canceled_template.html"
	TemplateFileOrderDelivered       = "order_delivered_template.html"
	TemplateFileOrderPaymentRequired = "order_payment_required_template.html"
	TemplateFileEmailVerification    = "verification_template.html"
	TemplateFileUserVerified         = "user_verified_template.html"
)
