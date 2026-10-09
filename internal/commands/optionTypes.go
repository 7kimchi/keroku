package commands

import (
	"time"

	"github.com/7kimchi/keroku/internal/validate"
	"github.com/bwmarrin/discordgo"
)

// Bool returns a boolean option.
func (r *Request) Bool(name string) (bool, bool, error) {
	v, ok, err := r.raw(name, discordgo.ApplicationCommandOptionBoolean)
	if !ok || err != nil {
		return false, ok, err
	}
	b, isBool := v.(bool)
	if !isBool {
		return false, false, invalid(name)
	}
	return b, true, nil
}

// Duration returns a duration option like "1d4h" within [minimum, maximum].
func (r *Request) Duration(name string, minimum, maximum time.Duration) (time.Duration, bool, error) {
	s, ok, err := r.Text(name, 32)
	if !ok || err != nil {
		return 0, ok, err
	}
	d, err := validate.Duration(s, minimum, maximum)
	if err != nil {
		return 0, false, Fail("Invalid option", name+" must be a duration like 1d4h, at most "+formatMax(maximum)+".")
	}
	return d, true, nil
}

// ResolvedMember returns the member data Discord resolved for a user option, if any.
func (r *Request) ResolvedMember(userID string) *discordgo.Member {
	if r.resolved == nil || r.resolved.Members == nil {
		return nil
	}
	return r.resolved.Members[userID]
}
