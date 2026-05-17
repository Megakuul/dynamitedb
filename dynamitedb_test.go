package dynamitdb

import "testing"

type User struct {
	UserId Field `pk:"user" json:"user_id"`
}

type Order struct {
	UserId  Field `pk:"user" json:"user_id"`
	OrderId Field `sk:"order" json:"order_id"`
}

func TestQuery(t *testing.T) {
	order := &Order{
		UserId: Eq("bombaclad"),
	}
	_ = order
}
