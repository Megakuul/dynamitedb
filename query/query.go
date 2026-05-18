package query

func Eq(operand string) *eqQuery {
	return &eqQuery{id: operand}
}

func BeginsWith(operand string) *beginsWithQuery {
	return &beginsWithQuery{prefix: operand}
}

type invalid struct{}

func (invalid) Value() string {
	panic("invalid operation: query operations are not supported in value structs")
}

type eqQuery struct {
	invalid
	id string
}

func (q eqQuery) Query() (string, bool) {
	return q.id, true
}

type beginsWithQuery struct {
	invalid
	prefix string
}

func (q beginsWithQuery) Query() (string, bool) {
	return q.prefix, false
}
