package client

import "github.com/dadyutenga/hms-control/internal/db/generated"

func RequestedSubdomainValue(t generated.Tenant) string {
	if t.RequestedSubdomain != nil && *t.RequestedSubdomain != "" {
		return *t.RequestedSubdomain
	}
	return t.Slug
}

func OptionalStringValue(value *string) string {
	if value != nil {
		return *value
	}
	return ""
}
