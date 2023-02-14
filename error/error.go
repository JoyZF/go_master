package error

type MyErr struct {
}

func (m MyErr) Error() string {
	return "error"
}
