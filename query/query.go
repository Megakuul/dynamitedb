// query KeyFields are used for indexed server side lookups.
// Query fields should only be used in the context of lookup operations (Get, List, etc.). Calling Value() on them panics.
package query

// Eq performs an exact match on the specified key id.
func Eq(id string) *eqQuery {
	return &eqQuery{id: id}
}

// BeginsWith reads all entries with the specified prefix.
// Using BeginsWith on PK and SK converts the PK to an exact match.
func BeginsWith(prefix string) *beginsWithQuery {
	return &beginsWithQuery{prefix: prefix}
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
