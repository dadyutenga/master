package admin

import (
	"fmt"
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
