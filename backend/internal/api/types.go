package api

type GenerateMnemonicResponse struct {
	Mnemonic  string `json:"mnemonic"`
	Address   string `json:"address"`
	PublicKey string `json:"public_key"`
	Email     string `json:"email"`
}

type DeriveAddressRequest struct {
	Mnemonic string `json:"mnemonic"`
}

type DeriveAddressResponse struct {
	Address   string `json:"address"`
	Email     string `json:"email"`
	PublicKey string `json:"public_key,omitempty"`
	Valid     bool   `json:"valid"`
}

type ValidateAddressResponse struct {
	Address string `json:"address"`
	Valid   bool   `json:"valid"`
	Reason  string `json:"reason,omitempty"`
}

type EmailResponse struct {
	ID         string   `json:"id"`
	From       string   `json:"from"`
	To         []string `json:"to"`
	Subject    string   `json:"subject"`
	Body       string   `json:"body"`
	ReceivedAt string   `json:"received_at"`
}

type InboxResponse struct {
	Address string          `json:"address"`
	Email   string          `json:"email"`
	Count   int             `json:"count"`
	Emails  []EmailResponse `json:"emails"`
}

type DeleteResponse struct {
	Deleted bool   `json:"deleted"`
	ID      string `json:"id,omitempty"`
	Count   int64  `json:"count,omitempty"`
}

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}
