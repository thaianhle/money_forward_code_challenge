package exception

type Error interface {
	error
	Check(value any)
}
