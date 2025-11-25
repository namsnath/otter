package specifier

type Specifier struct {
	Identifier string
	Value      string
}

func NewSpecifier(identifier, value string) *Specifier {
	return &Specifier{
		Identifier: identifier,
		Value:      value,
	}
}
