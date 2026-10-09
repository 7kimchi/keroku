package perms

// Code identifies why a check failed.
type Code int

// Failure codes, in check order.
const (
	MissingPermission Code = iota + 1
	TargetOwner
	TargetSelf
	TargetBot
	InvokerTooLow
	BotTooLow
	BotMissingPermission
)

// DenialError is a failed check. Message is safe to show to users.
type DenialError struct {
	Code       Code
	Permission int64
}

// Message returns short user facing copy.
func (d *DenialError) Message() string {
	switch d.Code {
	case MissingPermission:
		return "Missing permission: " + Name(d.Permission) + "."
	case TargetOwner:
		return "Target is the server owner."
	case TargetSelf:
		return "Target is you."
	case TargetBot:
		return "Target is Keroku."
	case InvokerTooLow:
		return "Target's top role is at or above yours."
	case BotTooLow:
		return "Target's top role is at or above Keroku's."
	case BotMissingPermission:
		return "Keroku is missing permission: " + Name(d.Permission) + "."
	}
	return "Not allowed."
}

// Error lets a denial travel as an error.
func (d *DenialError) Error() string { return d.Message() }
