package dynamitedb

import "time"

type Custom struct {
	Something string
}

type CustomEnum int32

const (
	CustomEnumExample CustomEnum = iota
)

type Test struct {
	PartID         KeyField                      `pk:"part" cbor:"-"`
	SortID         KeyField                      `sk:"sort" cbor:"-"`
	Nested         *NestedTest                   `cbor:"nested,omitempty"`
	TestString     DataField[string]             `cbor:"test_string,omitempty"`
	TestInt        DataField[int]                `cbor:"test_int,omitempty"`
	TestFloat      DataField[float64]            `cbor:"test_float,omitempty"`
	TestSlice      DataField[[]string]           `cbor:"test_slice,omitempty"`
	TestMap        DataField[map[string]string]  `cbor:"test_map,omitempty"`
	TestBool       DataField[bool]               `cbor:"test_bool,omitempty"`
	TestTime       DataField[time.Time]          `cbor:"test_time,omitempty"`
	TestDuration   DataField[time.Duration]      `cbor:"test_duration,omitempty"`
	TestCustom     DataField[map[string]*Custom] `cbor:"test_custom,omitempty"`
	TestCustomEnum DataField[CustomEnum]         `cbor:"test_custom_enum,omitempty"`

	TestUnmodified DataField[string]            `cbor:"test_unmodified,omitempty"`
	TestNil        DataField[string]            `cbor:"test_nil,omitempty"`
	TestNilMap     DataField[map[string]string] `cbor:"test_nil_map,omitempty"`
}

type NestedTest struct {
	TestString DataField[string] `cbor:"test_string,omitempty"`
	Nested     NestedNestedTest  `cbor:"nested"`
}

type NestedNestedTest struct {
	TestString DataField[string] `cbor:"test_string,omitempty"`
}
