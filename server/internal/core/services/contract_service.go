package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rmf87/divoene/internal/core/domain"
)

type ContractService struct {
	repo     domain.ContractRepository
	clicksign domain.ClicksignSender
}

func NewContractService(repo domain.ContractRepository, clicksign domain.ClicksignSender) *ContractService {
	return &ContractService{repo: repo, clicksign: clicksign}
}

func (s *ContractService) HasPersistence() bool { return s.repo != nil }

func (s *ContractService) Create(ctx context.Context, actor domain.Actor, req domain.CreateContractRequest) (*domain.Contract, error) {
	if err := validateCreateContractRequest(req); err != nil {
		return nil, err
	}

	contract, err := s.clicksign.CreateAndSend(&req)
	if err != nil {
		return nil, fmt.Errorf("clicksign: %w", err)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	contract.ID = fmt.Sprintf("CT-%d", time.Now().UnixNano())
	contract.Status = "sent"
	contract.SentAt = now
	contract.SentBy = actor.UID

	if s.repo != nil {
		if err := s.repo.Create(ctx, contract); err != nil {
			return nil, fmt.Errorf("persist: %w", err)
		}
	}

	return contract, nil
}

func (s *ContractService) Get(ctx context.Context, actor domain.Actor, id string) (*domain.Contract, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("contrato não encontrado")
	}
	contract, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !actor.HasRole("manager") && contract.SentBy != actor.UID {
		return nil, fmt.Errorf("contrato não encontrado")
	}
	return contract, nil
}

func (s *ContractService) HandleWebhook(ctx context.Context, event domain.ClicksignWebhookEvent) (string, error) {
	if s.repo == nil {
		return "", nil
	}

	contract, err := s.repo.GetByDocKey(ctx, event.DocumentKey)
	if err != nil {
		return "", nil
	}

	if contract.Status == "signed" || contract.Status == "declined" || contract.Status == "expired" {
		return contract.Status, nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	switch event.EventName {
	case "auto_close", "document_closed":
		extra := map[string]string{"signed_at": now}
		if err := s.repo.UpdateStatus(ctx, contract.ID, "signed", extra); err != nil {
			return "", err
		}
		return "signed", nil
	case "refusal":
		extra := map[string]string{"declined_at": now}
		if err := s.repo.UpdateStatus(ctx, contract.ID, "declined", extra); err != nil {
			return "", err
		}
		return "declined", nil
	case "cancel":
		extra := map[string]string{"declined_at": now}
		if err := s.repo.UpdateStatus(ctx, contract.ID, "expired", extra); err != nil {
			return "", err
		}
		return "expired", nil
	default:
		return contract.Status, nil
	}
}

// --- helpers ---

func buildContractMessage(req domain.CreateContractRequest) string {
	amountReais := float64(req.Amount) / 100.0
	msg := fmt.Sprintf(
		"Olá %s,\n\n"+
			"Segue o contrato para o seu evento na Chácara Divoene.\n\n"+
			"📅 Evento: %s\n"+
			"📦 Produto: %s\n"+
			"⏱ Duração: %d horas\n"+
			"👥 Público estimado: %d pessoas\n"+
			"💰 Valor total: R$ %.2f\n"+
			"💳 Condições: %s\n\n",
		req.ContactName, req.EventDate, productLabel(req.Product),
		req.EventDuration, req.EstPeople, amountReais, req.PaymentCond,
	)
	if len(req.AddOns) > 0 {
		msg += "🎯 Adicionais:\n"
		for _, a := range req.AddOns {
			addonReais := float64(a.UnitPrice) / 100.0
			msg += fmt.Sprintf("  - %s (x%d): R$ %.2f\n", a.Name, a.Quantity, addonReais*float64(a.Quantity))
		}
		msg += "\n"
	}
	if req.Notes != "" {
		msg += fmt.Sprintf("📝 Observações: %s\n\n", req.Notes)
	}
	msg += "Por favor, assine digitalmente para confirmar sua reserva.\n\n" +
		"Atenciosamente,\nEquipe Chácara Divoene"
	return msg
}

func productLabel(id string) string {
	labels := map[string]string{
		"ensaio_fotografico": "Ensaio Fotográfico",
		"locacao_eventos":    "Locação para Eventos",
		"corporativo":        "Corporativo",
		"casamentos":         "Casamentos",
		"buffet_infantil":    "Buffet Infantil",
		"passeios_escolares": "Passeios Escolares",
	}
	if l, ok := labels[id]; ok {
		return l
	}
	return id
}

var contractEmailRE = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func validateCreateContractRequest(req domain.CreateContractRequest) error {
	req.LeadID = strings.TrimSpace(req.LeadID)
	req.LeadName = strings.TrimSpace(req.LeadName)
	req.LeadEmail = strings.TrimSpace(req.LeadEmail)
	req.LeadWhatsApp = strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, req.LeadWhatsApp)
	req.Product = strings.TrimSpace(req.Product)
	req.EventDate = strings.TrimSpace(req.EventDate)
	req.PaymentCond = strings.TrimSpace(req.PaymentCond)

	if req.LeadID == "" {
		return fmt.Errorf("lead_id é obrigatório")
	}
	if req.LeadName == "" {
		return fmt.Errorf("nome do lead é obrigatório")
	}
	if req.LeadEmail == "" || !contractEmailRE.MatchString(req.LeadEmail) {
		return fmt.Errorf("email do lead é inválido")
	}
	if req.Amount <= 0 {
		return fmt.Errorf("valor do contrato deve ser maior que zero")
	}
	if req.Product == "" || !domain.IsValidProduct(req.Product) {
		return fmt.Errorf("produto inválido: %s", req.Product)
	}
	if req.EventDate == "" {
		return fmt.Errorf("data do evento é obrigatória")
	}
	if req.PaymentCond == "" {
		return fmt.Errorf("condições de pagamento são obrigatórias")
	}

	return nil
}
