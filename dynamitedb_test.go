package dynamitdb

import (
	"testing"

	"github.com/megakuul/dynamitdb/filter"
	"github.com/megakuul/dynamitdb/query"
)

type User struct {
	UserId       KeyField          `pk:"user" json:"user_id"`
	Organization DataField[string] `json:"organization"`
}

type Order struct {
	UserId  KeyField `pk:"user" json:"user_id"`
	OrderId KeyField `sk:"order" json:"order_id"`
}

func TestQuery(t *testing.T) {
	user := &User{
		UserId:       query.Eq("123"),
		Organization: filter.Eq("bombaclad"),
	}
	user.UserId.Value() // panicks
}
