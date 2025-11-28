package specifier

type Spec interface {
	getSpecifierKey() string
	getSpecifierValue() string
}

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
