package perms

import "github.com/bwmarrin/discordgo"

// Request describes one moderation action for the hierarchy check.
type Request struct {
	Guild *discordgo.Guild // fetched fresh, carries roles and owner
	Need  int64            // permission the action requires

	// Automated actions run as the bot and skip the invoker checks.
	Automated    bool
	InvokerID    string
	InvokerRoles []string
	InvokerPerms int64 // resolved for the channel, as sent with the interaction

	BotID    string
	BotRoles []string
	BotPerms int64 // resolved for the channel

	TargetID     string
	TargetRoles  []string
	TargetMember bool // false when the target is not in the guild
}

// Check runs the checks in a fixed order and returns the first failure, or nil.
func Check(r Request) *DenialError {
	if !r.Automated && !Has(r.InvokerPerms, r.Need) {
		return &DenialError{Code: MissingPermission, Permission: r.Need}
	}
	switch {
	case r.TargetID == r.Guild.OwnerID:
		return &DenialError{Code: TargetOwner}
	case !r.Automated && r.TargetID == r.InvokerID:
		return &DenialError{Code: TargetSelf}
	case r.TargetID == r.BotID:
		return &DenialError{Code: TargetBot}
	}
	if r.TargetMember {
		target := TopPosition(r.Guild, r.TargetRoles)
		invokerIsOwner := r.InvokerID == r.Guild.OwnerID
		if !r.Automated && !invokerIsOwner && TopPosition(r.Guild, r.InvokerRoles) <= target {
			return &DenialError{Code: InvokerTooLow}
		}
		if TopPosition(r.Guild, r.BotRoles) <= target {
			return &DenialError{Code: BotTooLow}
		}
	}
	if !Has(r.BotPerms, r.Need) {
		return &DenialError{Code: BotMissingPermission, Permission: r.Need}
	}
	return nil
}
