package dynamitedb

// KeyPrefix reads all entries with the specified prefix.
// Using KeyPrefix on PK and SK converts the PK to an exact match.
func KeyPrefix(prefix string) *beginsWithQuery {
	return &beginsWithQuery{prefix: prefix}
}

type beginsWithQuery struct {
	keyFallback
	prefix string
}

func (q beginsWithQuery) query() (string, bool) {
	return q.prefix, false
}
