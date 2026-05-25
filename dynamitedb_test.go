package dynamitedb

import "time"

type Test struct {
	PartID       KeyField                     `pk:"part" json:"-"`
	SortID       KeyField                     `sk:"sort" json:"-"`
	Nested       *NestedTest                  `json:"nested,omitempty"`
	TestString   DataField[string]            `json:"test_string,omitempty"`
	TestInt      DataField[int]               `json:"test_int,omitempty"`
	TestFloat    DataField[float64]           `json:"test_float,omitempty"`
	TestSlice    DataField[[]string]          `json:"test_slice,omitempty"`
	TestMap      DataField[map[string]string] `json:"test_map,omitempty"`
	TestBool     DataField[bool]              `json:"test_bool,omitempty"`
	TestTime     DataField[time.Time]         `json:"test_time,omitempty"`
	TestDuration DataField[time.Duration]     `json:"test_duration,omitempty"`

	TestUnmodified DataField[string]            `json:"test_unmodified,omitempty"`
	TestNil        DataField[string]            `json:"test_nil,omitempty"`
	TestNilMap     DataField[map[string]string] `json:"test_nil_map,omitempty"`
}

type NestedTest struct {
	TestString DataField[string] `json:"test_string,omitempty"`
	Nested     NestedNestedTest  `json:"nested"`
}

type NestedNestedTest struct {
	TestString DataField[string] `json:"test_string,omitempty"`
}
