package dynamitedb

// KeyEq performs an exact match on the specified key id.
func KeyEq(id string) *eqQuery {
	return &eqQuery{id: id}
}

// KeyBeginsWith reads all entries with the specified prefix.
// Using BeginsWith on PK and SK converts the PK to an exact match.
func KeyBeginsWith(prefix string) *beginsWithQuery {
	return &beginsWithQuery{prefix: prefix}
}

type eqQuery struct {
	keyFallback
	id string
}

func (q eqQuery) query() (string, bool) {
	return q.id, true
}

type beginsWithQuery struct {
	keyFallback
	prefix string
}

func (q beginsWithQuery) query() (string, bool) {
	return q.prefix, false
}
