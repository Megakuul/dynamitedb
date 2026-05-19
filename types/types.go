package types

type DataConstraint interface {
	string | int | float64 | bool | []string | map[string]string
}
