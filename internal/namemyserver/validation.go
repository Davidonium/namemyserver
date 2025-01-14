package namemyserver

import "regexp"

// generated names must be dns subdomain compliant just like kubernetes resources, with the added constraint of them being lowercase
// validates that all characters are lowercase, alphanumberic and the name must start and end with alphanumeric characters.
// reference: https://datatracker.ietf.org/doc/html/rfc1123
var nameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

var segmentRegex = regexp.MustCompile(`^[a-z0-9]+$`)

func ValidateName(name string) bool {
	return nameRegex.MatchString(name)
}

func ValidateNameSegment(s string) bool {
	return segmentRegex.MatchString(s)
}
