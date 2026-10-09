package config

import "log/slog"

const redacted = "[redacted]"

// Secret holds a credential. Every print, log and JSON path shows a placeholder.
type Secret struct {
	value string
}

// NewSecret wraps a raw credential.
func NewSecret(v string) Secret {
	return Secret{value: v}
}

// Reveal returns the raw value. Only pass it to the code that needs it.
func (s Secret) Reveal() string { return s.value }

// Empty reports whether no value is set.
func (s Secret) Empty() bool { return s.value == "" }

// String hides the value from fmt.
func (s Secret) String() string { return redacted }

// GoString hides the value from %#v.
func (s Secret) GoString() string { return redacted }

// LogValue hides the value from slog.
func (s Secret) LogValue() slog.Value { return slog.StringValue(redacted) }

// MarshalJSON hides the value from encoding/json.
func (s Secret) MarshalJSON() ([]byte, error) { return []byte(`"` + redacted + `"`), nil }

// MarshalText hides the value from text encoders.
func (s Secret) MarshalText() ([]byte, error) { return []byte(redacted), nil }
