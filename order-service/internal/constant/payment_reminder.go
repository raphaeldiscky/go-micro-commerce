package constant

import "time"

const (
	// FirstReminderSequence represents the first reminder in the sequence.
	FirstReminderSequence = 1
	// SecondReminderSequence represents the second reminder in the sequence.
	SecondReminderSequence = 2
)

const (
	// FirstPaymentReminderDelay is the delay for first task scheduling.
	FirstPaymentReminderDelay = 10 * time.Hour
	// SecondPaymentReminderDelay is the delay for second task scheduling.
	SecondPaymentReminderDelay = 20 * time.Hour
	// ExpireOrderReminderDelay is the delay for cancel task scheduling.
	ExpireOrderReminderDelay = 24 * time.Hour
)

const (
	// FirstPaymentReminderEmailSubject is the subject for the first payment reminder.
	FirstPaymentReminderEmailSubject = "Payment Reminder - Your Order is Waiting"
	// SecondPaymentReminderEmailSubject is the subject for the second payment reminder.
	SecondPaymentReminderEmailSubject = "Final Payment Reminder - Your Order Expires Soon"
	// OrderPaymentExpiredEmailSubject is the subject for the order cancellation email.
	OrderPaymentExpiredEmailSubject = "Your Order Has Expired - Payment Timeout"
)
