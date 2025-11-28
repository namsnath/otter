package specifier

type Specifier struct {
	Identifier string
	Value      string
}

func NewSpecifier(identifier string, value string) *Specifier {
	return &Specifier{
		Identifier: identifier,
		Value:      value,
	}
}
