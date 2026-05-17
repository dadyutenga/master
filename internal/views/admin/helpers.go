package admin

import (
	"fmt"
	"strings"
	"time"

	"github.com/dadyutenga/hms-control/internal/db/generated"
)

type DockerTemplateOption struct {
	ID   int64
	Name string
}

func FormatDate(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02")
}

func OptionalStringValue(value *string) string {
	if value != nil {
		return *value
	}
	return ""
}

func FormatFileSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}

func DocTypeLabel(docType string) string {
	switch docType {
	case "brela_certificate":
		return "BRELA"
	case "tra_certificate":
		return "TRA"
	default:
		return docType
	}
}

func FormatAmount(amount float64) string {
	if amount == 0 {
		return "0 TZS"
	}
	return fmt.Sprintf("%.0f TZS", amount)
}

func DomainDisplay(domain string) string {
	if strings.HasPrefix(domain, "pending-") {
		return "Not set"
	}
	return domain
}

func ActiveLabel(active bool) string {
	if active {
		return "Deactivate"
	}
	return "Activate"
}

func PMBadgeStyle(methodType string) string {
	switch methodType {
	case "card":
		return "background:hsl(217, 91%, 95%); color:hsl(217, 91%, 40%);"
	case "lipa_namba":
		return "background:hsl(140, 60%, 95%); color:hsl(140, 60%, 35%);"
	default:
		return "background:hsl(25, 80%, 95%); color:hsl(25, 80%, 40%);"
	}
}

func IsPendingDomain(domain string) bool {
	return strings.HasPrefix(domain, "pending-")
}

type PaymentSummary struct {
	TotalCount  int
	TotalAmount float64
	ClientCount int
	LatestDate  string
}

func calcPaymentSummary(payments []generated.ListPaymentsRow) PaymentSummary {
	var s PaymentSummary
	seen := make(map[string]bool)
	for _, p := range payments {
		s.TotalCount++
		s.TotalAmount += p.Amount
		if !seen[p.TenantID] {
			seen[p.TenantID] = true
			s.ClientCount++
		}
	}
	if s.TotalCount > 0 {
		s.LatestDate = payments[0].CreatedAt.Format("02 Jan 2006")
	}
	return s
}

func ExtractReceiptPath(description string) string {
	prefix := " | File: "
	if idx := strings.LastIndex(description, prefix); idx != -1 {
		return description[idx+len(prefix):]
	}
	return ""
}

func PaymentDescriptionShort(description string) string {
	// Remove the file path from description for cleaner display
	if idx := strings.LastIndex(description, " | File: "); idx != -1 {
		return description[:idx]
	}
	return description
}
