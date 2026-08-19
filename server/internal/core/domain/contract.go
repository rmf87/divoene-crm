package domain

import "context"

// ContractAddOnItem represents an add-on service included in the contract.
type ContractAddOnItem struct {
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	UnitPrice int64  `json:"unit_price"`
}

// Contract represents a digital contract in the system.
type Contract struct {
	ID              string             `json:"id"`
	LeadID          string             `json:"lead_id"`
	ClicksignDocKey string             `json:"clicksign_doc_key"`
	SignerKey       string             `json:"signer_key"`
	RequestSignKey  string             `json:"request_sign_key"`
	Amount          int64              `json:"amount"`
	Product         string             `json:"product"`
	EventType       string             `json:"event_type"`
	EventDate       string             `json:"event_date"`
	EventDuration   int                `json:"event_duration"`
	EstPeople       int                `json:"estimated_people"`
	ContactName     string             `json:"contact_name"`
	ContactWpp      string             `json:"contact_whatsapp"`
	ContactRole     string             `json:"contact_role"`
	AddOns          []ContractAddOnItem `json:"add_ons"`
	PaymentCond     string             `json:"payment_conditions"`
	Notes           string             `json:"notes,omitempty"`
	Status          string             `json:"status"`
	SentAt          string             `json:"sent_at"`
	SentBy          string             `json:"sent_by"`
	SignedAt        string             `json:"signed_at,omitempty"`
	DeclinedAt      string             `json:"declined_at,omitempty"`
	CreatedAt       string             `json:"created_at,omitempty"`
}

// CreateContractRequest is the payload for POST /api/contracts.
type CreateContractRequest struct {
	LeadID        string             `json:"lead_id"`
	LeadName      string             `json:"lead_name"`
	LeadEmail     string             `json:"lead_email"`
	LeadWhatsApp  string             `json:"lead_whatsapp"`
	Amount        int64              `json:"amount"`
	Product       string             `json:"product"`
	EventType     string             `json:"event_type"`
	EventDate     string             `json:"event_date"`
	EventDuration int                `json:"event_duration"`
	EstPeople     int                `json:"estimated_people"`
	ContactName   string             `json:"contact_name"`
	ContactWpp    string             `json:"contact_whatsapp"`
	ContactRole   string             `json:"contact_role"`
	AddOns        []ContractAddOnItem `json:"add_ons"`
	PaymentCond   string             `json:"payment_conditions"`
	Notes         string             `json:"notes,omitempty"`
}

// ClicksignWebhookEvent represents a parsed Clicksign webhook payload.
type ClicksignWebhookEvent struct {
	EventName      string
	DocumentKey    string
	SignerKey      string
	SignerEmail    string
	SignerName     string
	RefusalReasons []string
	RefusalComment string
}

// ContractRepository defines persistence for contracts.
type ContractRepository interface {
	Create(ctx context.Context, c *Contract) error
	Get(ctx context.Context, id string) (*Contract, error)
	GetByDocKey(ctx context.Context, docKey string) (*Contract, error)
	ListByLead(ctx context.Context, leadID string) ([]*Contract, error)
	ListAll(ctx context.Context) ([]*Contract, error)
	UpdateStatus(ctx context.Context, id, status string, extra map[string]string) error
}

// ClicksignDocKey holds the Clicksign document, signer, and request keys.
type ClicksignDocKey struct {
	DocumentKey string
	SignerKey   string
	RequestKey  string
}

// ClicksignSender abstracts the Clicksign API for creating and sending contracts.
type ClicksignSender interface {
	CreateAndSend(contract *CreateContractRequest) (*Contract, error)
	VerifyWebhookSignature(body []byte, contentHmacHeader string) bool
	IsMock() bool
	GetWebhookSecret() string
	PollStatus(documentKey string) (string, error)
}
