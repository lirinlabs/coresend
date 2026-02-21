package api

type RegisterResponse struct {
	Registered bool   `json:"registered" example:"true" validate:"required"`
	Address    string `json:"address" example:"0x123abc" validate:"required,hexadecimal"`
	ExpiresIn  int    `json:"expires_in" example:"86400" validate:"required,gt=0"`
}
type EmailResponse struct {
	ID         string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" binding:"required"`
	From       string   `json:"from" example:"sender@example.com"`
	To         []string `json:"to" example:"[\"0x123abc@coresend.dev\"]"`
	Subject    string   `json:"subject" example:"Hello World"`
	Body       string   `json:"body" example:"This is the email body content"`
	ReceivedAt string   `json:"received_at" example:"2024-01-01T12:00:00Z"`
}

type InboxResponse struct {
	Address string          `json:"address" example:"0x123abc"`
	Email   string          `json:"email" example:"0x123abc@coresend.dev"`
	Count   int             `json:"count" example:"5"`
	Emails  []EmailResponse `json:"emails"`
}

type DeleteResponse struct {
	Deleted bool   `json:"deleted" example:"true"`
	ID      string `json:"id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Count   int64  `json:"count,omitempty" example:"5"`
}

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code    string                 `json:"code" example:"INVALID_ADDRESS"`
	Message string                 `json:"message" example:"Address is required"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type HealthResponse struct {
	Status   string            `json:"status" example:"connected"`
	Services map[string]string `json:"services" example:"{\"redis\":\"connected\",\"smtp\":\"running\"}"`
}
