package errors

type safeError struct {
	err error
	msg string
}

func SafeWrap(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &safeError{
		err: err,
		msg: msg,
	}
}

func (se *safeError) Error() string {
	return se.msg + ": " + se.err.Error()
}

func (se *safeError) safeError() string {
	if e, ok := se.err.(*safeError); ok {
		return se.msg + ": " + e.safeError()
	}
	return se.msg
}

func SafeError(err error) string {
	if e, ok := err.(*safeError); ok {
		return e.safeError()
	}
	return ""
}
