package admin

import "time"

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
