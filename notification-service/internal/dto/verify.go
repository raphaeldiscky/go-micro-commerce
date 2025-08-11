package dto

type SendVerificationEvent struct {
	Token string `json:"token"`
}

type AccountVerifiedEvent struct {
	Email string `json:"email"`
}
