package errors

type safeError struct {
	err error
	msg string
}

func SafeWrap(err error, msg string) error {
	return &safeError{
		err: err,
		msg: msg,
	}
}

func (se *safeError) Error() string {
	if se.err == nil {
		return se.msg
	}
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
