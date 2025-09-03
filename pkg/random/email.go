package random

import (
	"fmt"
	"strings"
)

const (
	// DefaultEmailDomains contains common email domains for generating random emails.
	defaultEmailDomains = "gmail.com,yahoo.com,outlook.com,hotmail.com,example.com"
)

// Email generates a random email address using common domains.
// Format: randomstring@randomdomain.
func Email() string {
	domains := strings.Split(defaultEmailDomains, ",")
	username := strings.ToLower(AlphaString(8))
	domainIdx := Int(int64(len(domains)))

	return fmt.Sprintf("%s@%s", username, domains[domainIdx])
}

// EmailWithDomain generates a random email address with a specific domain.
func EmailWithDomain(domain string) string {
	if domain == "" {
		return Email()
	}

	username := strings.ToLower(AlphaString(8))

	return fmt.Sprintf("%s@%s", username, domain)
}

// EmailWithPrefix generates a random email address with a specific username prefix.
func EmailWithPrefix(prefix string) string {
	if prefix == "" {
		return Email()
	}

	domains := strings.Split(defaultEmailDomains, ",")
	suffix := strings.ToLower(AlphaString(4))
	domainIdx := Int(int64(len(domains)))

	return fmt.Sprintf("%s%s@%s", prefix, suffix, domains[domainIdx])
}
