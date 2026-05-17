package client

import (
	"github.com/dadyutenga/hms-control/internal/db/generated"
)

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

func pmIcon(methodType string) string {
	switch methodType {
	case "card":
		return "credit-card"
	case "lipa_namba", "mobile":
		return "smartphone"
	default:
		return "credit-card"
	}
}

func pmStyle(methodType string) string {
	switch methodType {
	case "card":
		return "background:hsl(217, 91%, 95%); color:hsl(217, 91%, 40%);"
	case "lipa_namba":
		return "background:hsl(140, 60%, 95%); color:hsl(140, 60%, 35%);"
	default:
		return "background:hsl(25, 80%, 95%); color:hsl(25, 80%, 40%);"
	}
}

func pmBadgeStyle(methodType string) string {
	switch methodType {
	case "card":
		return "background:hsl(217, 91%, 95%); color:hsl(217, 91%, 40%);"
	case "lipa_namba":
		return "background:hsl(140, 60%, 95%); color:hsl(140, 60%, 35%);"
	default:
		return "background:hsl(25, 80%, 95%); color:hsl(25, 80%, 40%);"
	}
}
