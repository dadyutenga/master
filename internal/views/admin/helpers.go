package admin

import (
	"fmt"
	"strings"
	"time"
)

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
	return fmt.Sprintf("%,.0f TZS", amount)
}

func DomainDisplay(domain string) string {
	if strings.HasPrefix(domain, "pending-") {
		return "Not set"
	}
	return domain
}

func IsPendingDomain(domain string) bool {
	return strings.HasPrefix(domain, "pending-")
}
