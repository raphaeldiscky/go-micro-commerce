package constant

const (
	// SendVerificationSubject is the subject for email verification.
	SendVerificationSubject = "Email Verification"
	// UserVerifiedSubject is the subject for User verification.
	UserVerifiedSubject = "User Verified"
)

const (
	// SendVerificationTemplate is the email template for sending verification emails.
	SendVerificationTemplate = `
<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
	</head>
	<body>
		<h1>Email Verification</h1>
		<p>Dear User,</p>
		<p>Thank you for registering. Please click the following link to verify your email address:</p>
		<h2>%v</h2>
		<p>This link will expire in 10 minutes.</p>
		<p>If you did not request this, please ignore this email.</p>
		<p>Best regards,</p>
		<p>The Go Microservices Team</p>
	</body>
</html>
	`
	// UserVerifiedTemplate is the email template for sending User verified emails.
	UserVerifiedTemplate = `
<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
	</head>
	<body>
		<h1>User Verified</h1>
		<p>Dear User,</p>
		<p>We are pleased to inform you that your User has been successfully verified.</p>
		<p>You can now access all the features of our service.</p>
		<p>If you have any questions or need further assistance, please do not hesitate to contact us.</p>
		<p>Best regards,</p>
		<p>The Go Microservices Team</p>
	</body>
</html>
	`
)
