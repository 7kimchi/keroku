package config

import (
	"errors"
	"fmt"
	"strings"
)

const maxSecretBytes = 4096

// loadSecret reads NAME or NAME_FILE. Exactly one must be set.
func loadSecret(src Source, name string) (Secret, error) {
	direct := src.Getenv(name)
	path := src.Getenv(name + "_FILE")
	switch {
	case direct != "" && path != "":
		return Secret{}, fmt.Errorf("config: set %s or %s_FILE, not both", name, name)
	case direct != "":
		return checkSecret(name, direct)
	case path != "":
		raw, err := src.ReadFile(path)
		if err != nil {
			return Secret{}, fmt.Errorf("config: read %s_FILE: %w", name, err)
		}
		if len(raw) > maxSecretBytes {
			return Secret{}, fmt.Errorf("config: %s_FILE is too large", name)
		}
		// Secret files usually end with a newline. Nothing else is trimmed.
		return checkSecret(name, strings.TrimRight(string(raw), "\r\n"))
	}
	return Secret{}, fmt.Errorf("config: %s or %s_FILE is required", name, name)
}

func checkSecret(name, v string) (Secret, error) {
	if v == "" || len(v) > maxSecretBytes {
		return Secret{}, fmt.Errorf("config: %s is empty or too long", name)
	}
	for i := 0; i < len(v); i++ {
		if v[i] <= ' ' || v[i] > '~' {
			return Secret{}, fmt.Errorf("config: %s has whitespace or non printable characters", name)
		}
	}
	return NewSecret(v), nil
}

func loadToken(src Source) (Secret, error) {
	s, err := loadSecret(src, "DISCORD_TOKEN")
	if err != nil {
		return Secret{}, err
	}
	// A bot token is three non empty dot separated parts. The "Bot " prefix is added later.
	parts := strings.Split(s.Reveal(), ".")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return Secret{}, errors.New("config: DISCORD_TOKEN must be the raw bot token")
	}
	return s, nil
}
