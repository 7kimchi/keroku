package validate

// Int checks that v is within [minimum, maximum].
func Int(v, minimum, maximum int64) error {
	if v < minimum || v > maximum {
		return ErrRange
	}
	return nil
}
