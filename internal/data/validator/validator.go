package validator

import (
	"cmp"
	"regexp"
	"slices"
	"strings"
)

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(k, msg string) {
	if _, ok := v.Errors[k]; !ok {
		v.Errors[k] = msg
	}
}

func (v *Validator) Check(ok bool, k, msg string) {
	if !ok {
		v.AddError(k, msg)
	}
}

func PermittedValue[T comparable](v T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, v)
}

func Matches(v string, rx *regexp.Regexp) bool {
	return rx.MatchString(v)
}

func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool, len(values))

	for _, v := range values {

		if v, ok := any(v).(string); ok && !ValidString(v, 500) {
			return false
		}

		uniqueValues[v] = true
	}

	return len(values) == len(uniqueValues)
}

func WithinRange[T cmp.Ordered](v, low, high T) bool {
	return low <= v && v <= high
}

func ValidString(s string, maxBytes int) bool {
	return strings.TrimSpace(s) != "" && len(s) <= maxBytes
}
