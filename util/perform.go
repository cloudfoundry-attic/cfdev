package util

func Perform(attempts int, task func() error) (err error) {
	for {
		switch attempts {
		case 0:
			return
		default:
			attempts -= 1

			err = task()
			if err == nil {
				return
			}
		}
	}
}
