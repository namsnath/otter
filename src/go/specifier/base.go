package specifier

type Spec interface {
	getSpecifierKey() string
	getSpecifierValue() string
}

type Specifier struct {
	Key   string
	Value string
}

func NewSpecifier(key string, value string) Specifier {
	return Specifier{
		Key:   key,
		Value: value,
	}
}
