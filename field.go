package dynamitdb

type KeyField interface {
	Value() string
	Query() (string, bool)
}

type StringField interface {
	Value() string
	Filter(string) bool
}

type IntField interface {
	Value() int
	Filter(int) bool
}

type FloatField interface {
	Value() float64
	Filter(int) bool
}

type SliceField interface {
	Value() []string
	Filter([]string) bool
}

type MapField interface {
	Value() map[string]string
	Filter(map[string]string) bool
}
