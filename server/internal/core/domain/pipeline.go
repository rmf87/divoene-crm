package domain

// ValidStages is the ordered list of pipeline stages.
var ValidStages = []string{
	"lead",
	"validated",
	"visit_scheduled",
	"visit_done",
	"contract",
	"paid",
	"booked",
	"completed",
	"cancelled",
}

// AllowedTransitions defines which stage changes are valid.
var AllowedTransitions = map[string][]string{
	"lead":            {"validated", "cancelled"},
	"validated":       {"visit_scheduled", "cancelled"},
	"visit_scheduled": {"visit_done", "cancelled"},
	"visit_done":      {"contract", "cancelled"},
	"contract":        {"paid", "cancelled"},
	"paid":            {"booked", "cancelled"},
	"booked":          {"completed", "cancelled"},
	"completed":       {},
	"cancelled":       {},
}

// ValidProducts is the list of products accepted by the system.
var ValidProducts = []string{
	"ensaio_fotografico",
	"locacao_eventos",
	"corporativo",
	"casamentos",
	"buffet_infantil",
	"passeios_escolares",
}

// CanTransition checks if moving from one stage to another is allowed.
func CanTransition(from, to string) bool {
	targets, ok := AllowedTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

// IsValidStage checks if a stage name is valid.
func IsValidStage(s string) bool {
	for _, vs := range ValidStages {
		if vs == s {
			return true
		}
	}
	return false
}

// IsValidProduct checks if a product ID is valid.
func IsValidProduct(p string) bool {
	for _, vp := range ValidProducts {
		if vp == p {
			return true
		}
	}
	return false
}
