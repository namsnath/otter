package specifier

const RoleSpecifierKey = "role"

type RoleSpecifierValue string

const (
	RoleSpecifierAdmin RoleSpecifierValue = "admin"
)

// var RoleAdmin = NewSpecifier(RoleSpecifierKey, "admin")

type RoleSpecifier struct {
	Value RoleSpecifierValue
}

func NewRoleSpecifier(value RoleSpecifierValue) RoleSpecifier {
	return RoleSpecifier{
		Value: value,
	}
}

func (r RoleSpecifier) getSpecifierKey() string {
	return RoleSpecifierKey
}

func (r RoleSpecifier) getSpecifierValue() RoleSpecifierValue {
	return r.Value
}
