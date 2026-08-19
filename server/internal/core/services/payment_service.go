package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rmf87/divoene/internal/core/domain"
)

// PaymentService handles payment business logic.
type PaymentService struct {
	repo   domain.PaymentRepository
	openpix domain.PIXCharger
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(repo domain.PaymentRepository, op domain.PIXCharger) *PaymentService {
	return &PaymentService{repo: repo, openpix: op}
}

// HasPersistence reports whether the service has a working database connection.
func (s *PaymentService) HasPersistence() bool { return s.repo != nil }

// Create creates a PIX charge via OpenPix.
func (s *PaymentService) Create(ctx context.Context, actor domain.Actor, req domain.CreatePaymentRequest) (*domain.Payment, error) {
	if err := validatePaymentCreateRequest(req); err != nil {
		return nil, err
	}

	result, err := s.openpix.CreateCharge(req.Amount, req.Description, req.ChargeType, req.PayerName, req.PayerEmail, req.PayerTaxID, req.PayerPhone, nil)
	if err != nil {
		return nil, fmt.Errorf("openpix: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	payment := &domain.Payment{
		ID:                   fmt.Sprintf("PAY-%d", time.Now().UnixNano()),
		LeadID:               req.LeadID,
		ContractID:           req.ContractID,
		OpenPixTransactionID: result.TransactionID,
		Amount:               req.Amount,
		Description:          req.Description,
		Type:                 req.Type,
		ChargeType:           req.ChargeType,
		Status:               "pending",
		BrCode:               result.BrCode,
		PaymentLinkURL:       result.PaymentLink,
		CreatedAt:            now,
		CreatedBy:            actor.UID,
	}

	if s.repo != nil {
		if err := s.repo.Create(ctx, payment); err != nil {
			return nil, fmt.Errorf("persist: %w", err)
		}
	}

	return payment, nil
}

// Get retrieves a payment by ID, enforcing role visibility.
func (s *PaymentService) Get(ctx context.Context, actor domain.Actor, id string) (*domain.Payment, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("pagamento não encontrado")
	}
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !actor.HasRole("manager") && p.CreatedBy != actor.UID {
		return nil, fmt.Errorf("pagamento não encontrado")
	}
	return p, nil
}

// HandleWebhook processes an OpenPix webhook notification.
func (s *PaymentService) HandleWebhook(ctx context.Context, txID string) (string, error) {
	if s.repo == nil {
		return "", nil
	}

	opStatus, err := s.openpix.GetCharge(txID)
	if err != nil {
		return "", fmt.Errorf("query openpix: %w", err)
	}

	payment, err := s.repo.GetByTransactionID(ctx, txID)
	if err != nil {
		return "", nil
	}

	if payment.Status == "confirmed" || payment.Status == "expired" || payment.Status == "cancelled" {
		return payment.Status, nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	switch opStatus {
	case "COMPLETED":
		extra := map[string]string{"confirmed_at": now}
		if err := s.repo.UpdateStatus(ctx, payment.ID, "confirmed", extra); err != nil {
			return "", err
		}
		return "confirmed", nil
	case "EXPIRED", "CANCELLED":
		extra := map[string]string{}
		if err := s.repo.UpdateStatus(ctx, payment.ID, "cancelled", extra); err != nil {
			return "", err
		}
		return "cancelled", nil
	default:
		return payment.Status, nil
	}
}

var paymentEmailRE = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func validatePaymentCreateRequest(req domain.CreatePaymentRequest) error {
	req.LeadID = strings.TrimSpace(req.LeadID)
	req.ContractID = strings.TrimSpace(req.ContractID)
	req.Description = strings.TrimSpace(req.Description)
	req.Type = strings.TrimSpace(req.Type)
	req.PayerEmail = strings.TrimSpace(req.PayerEmail)
	req.PayerName = strings.TrimSpace(req.PayerName)

	if req.LeadID == "" {
		return fmt.Errorf("lead_id é obrigatório")
	}
	if req.ContractID == "" {
		return fmt.Errorf("contract_id é obrigatório")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("valor deve ser maior que zero")
	}
	if req.Description == "" {
		return fmt.Errorf("descrição é obrigatória")
	}
	if req.Type == "" || !isValidPaymentType(req.Type) {
		return fmt.Errorf("tipo de pagamento inválido: %s (use 'sinal' ou 'total')", req.Type)
	}
	if req.ChargeType != "" && req.ChargeType != "pix" && req.ChargeType != "pix_credit" {
		return fmt.Errorf("charge_type inválido: %s (use 'pix' ou 'pix_credit')", req.ChargeType)
	}
	if req.PayerEmail == "" || !paymentEmailRE.MatchString(req.PayerEmail) {
		return fmt.Errorf("email do pagador é inválido")
	}
	if req.PayerName == "" {
		return fmt.Errorf("nome do pagador é obrigatório")
	}

	return nil
}

func isValidPaymentType(t string) bool {
	return t == "sinal" || t == "total"
}
