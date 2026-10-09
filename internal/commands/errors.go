package commands

// UserError is a failure whose copy is safe to show. Title reads like "Ban failed".
type UserError struct {
	Title  string
	Detail string
}

func (e *UserError) Error() string { return e.Title + ". " + e.Detail }

// Fail builds a UserError.
func Fail(title, detail string) *UserError {
	return &UserError{Title: title, Detail: detail}
}

func invalid(name string) *UserError {
	return Fail("Invalid option", "Value for "+name+" is not valid.")
}
