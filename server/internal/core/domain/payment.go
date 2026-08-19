package domain

import "context"

// Payment represents a payment in the system.
type Payment struct {
	ID                   string `json:"id"`
	LeadID               string `json:"lead_id"`
	ContractID           string `json:"contract_id"`
	OpenPixTransactionID string `json:"openpix_transaction_id"`
	Amount               int64  `json:"amount"`
	Description          string `json:"description"`
	Type                 string `json:"type"`
	ChargeType           string `json:"charge_type"`
	Status               string `json:"status"`
	BrCode               string `json:"br_code"`
	PaymentLinkURL       string `json:"payment_link_url"`
	CreatedAt            string `json:"created_at"`
	CreatedBy            string `json:"created_by"`
	ConfirmedAt          string `json:"confirmed_at,omitempty"`
}

// CreatePaymentRequest is the payload for POST /api/payments.
type CreatePaymentRequest struct {
	LeadID      string `json:"lead_id"`
	ContractID  string `json:"contract_id"`
	Amount      int64  `json:"amount"`
	Description string `json:"description"`
	Type        string `json:"type"`
	ChargeType  string `json:"charge_type"`
	PayerEmail  string `json:"payer_email"`
	PayerName   string `json:"payer_name"`
	PayerTaxID  string `json:"payer_taxid,omitempty"`
	PayerPhone  string `json:"payer_phone,omitempty"`
}

// WebhookSignature contains the OpenPix webhook signing info.
type WebhookSignature struct {
	XSignature string
	XRequestID string
	DataID     string
}

// PaymentRepository defines persistence for payments.
type PaymentRepository interface {
	Create(ctx context.Context, p *Payment) error
	Get(ctx context.Context, id string) (*Payment, error)
	GetByTransactionID(ctx context.Context, txID string) (*Payment, error)
	UpdateStatus(ctx context.Context, id string, status string, extra map[string]string) error
	ListByLead(ctx context.Context, leadID string) ([]*Payment, error)
}

// PIXCharger abstracts PIX charge operations (OpenPix).
type PIXCharger interface {
	CreateCharge(amount int64, description, chargeType, payerName, payerEmail, payerTaxID, payerPhone string, addr *ChargeAddress) (*PIXChargeResult, error)
	GetCharge(transactionID string) (string, error)
	IsMock() bool
}

// ChargeAddress represents optional payer address for PIX.
type ChargeAddress struct {
	Street       string `json:"street,omitempty"`
	Number       string `json:"number,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	ZipCode      string `json:"zipcode,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	Complement   string `json:"complement,omitempty"`
	Country      string `json:"country,omitempty"`
}

// PIXChargeResult is the result from creating a PIX charge.
type PIXChargeResult struct {
	TransactionID string `json:"transaction_id"`
	BrCode        string `json:"br_code"`
	PaymentLink   string `json:"payment_link"`
	ExpiresIn     int    `json:"expires_in"`
}
