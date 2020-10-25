package yamlreg

type ErrInvalidOutputArgument struct {
	Reason string
}

func (t ErrInvalidOutputArgument) Error() string {
	return "invalid parameter 'out' " + t.Reason
}
