package orderstatus

type Status string

const (
	StatusCreated          Status = "created"
	StatusPaymentPending   Status = "payment_pending"
	StatusPaymentConfirmed Status = "payment_confirmed"
	StatusShipping         Status = "shipping"
	StatusCompleted        Status = "completed"
	StatusFailed           Status = "failed"
	StatusCancelled        Status = "cancelled"
)
