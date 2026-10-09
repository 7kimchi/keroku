package commands

// UserError is a failure whose copy is safe to show. Title reads like "Ban failed".
// Cause, when set, is logged with the reference id, and the reply shows the id.
type UserError struct {
	Title  string
	Detail string
	Cause  error
}

func (e *UserError) Error() string { return e.Title + ". " + e.Detail }

// Fail builds a UserError.
func Fail(title, detail string) *UserError {
	return &UserError{Title: title, Detail: detail}
}

// FailWith builds a UserError that also carries an internal cause for the log.
func FailWith(title, detail string, cause error) *UserError {
	return &UserError{Title: title, Detail: detail, Cause: cause}
}

func (e *UserError) Unwrap() error { return e.Cause }

func invalid(name string) *UserError {
	return Fail("Invalid option", "Value for "+name+" is not valid.")
}
