package commands

import (
	"math"

	"github.com/bwmarrin/discordgo"

	"github.com/7kimchi/keroku/internal/validate"
)

func (r *Request) raw(name string, t discordgo.ApplicationCommandOptionType) (any, bool, error) {
	o, ok := r.opts[name]
	if !ok {
		return nil, false, nil
	}
	if o.Type != t || o.Value == nil {
		return nil, false, invalid(name)
	}
	return o.Value, true, nil
}

// Text returns a trimmed, cleaned string option of at most maxRunes.
func (r *Request) Text(name string, maxRunes int) (string, bool, error) {
	v, ok, err := r.raw(name, discordgo.ApplicationCommandOptionString)
	if !ok || err != nil {
		return "", ok, err
	}
	s, isString := v.(string)
	if !isString {
		return "", false, invalid(name)
	}
	clean, err := validate.Text(s, maxRunes)
	if err != nil {
		return "", false, Fail("Invalid option", name+" is too long or has invalid characters.")
	}
	return clean, true, nil
}

// Snowflake returns a user, channel or role option as a validated id.
func (r *Request) Snowflake(name string, t discordgo.ApplicationCommandOptionType) (string, bool, error) {
	v, ok, err := r.raw(name, t)
	if !ok || err != nil {
		return "", ok, err
	}
	s, isString := v.(string)
	if !isString {
		return "", false, invalid(name)
	}
	if _, err := validate.Snowflake(s); err != nil {
		return "", false, invalid(name)
	}
	return s, true, nil
}

// User returns a user option id.
func (r *Request) User(name string) (string, bool, error) {
	return r.Snowflake(name, discordgo.ApplicationCommandOptionUser)
}

// Channel returns a channel option id.
func (r *Request) Channel(name string) (string, bool, error) {
	return r.Snowflake(name, discordgo.ApplicationCommandOptionChannel)
}

// Int returns an integer option within [minimum, maximum].
func (r *Request) Int(name string, minimum, maximum int64) (int64, bool, error) {
	v, ok, err := r.raw(name, discordgo.ApplicationCommandOptionInteger)
	if !ok || err != nil {
		return 0, ok, err
	}
	f, isNumber := v.(float64)
	if !isNumber || f != math.Trunc(f) || f < float64(minimum) || f > float64(maximum) {
		return 0, false, Fail("Invalid option", name+" must be a whole number from "+itoa(minimum)+" to "+itoa(maximum)+".")
	}
	return int64(f), true, nil
}
