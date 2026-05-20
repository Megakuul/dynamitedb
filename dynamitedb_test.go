package dynamitdb

type Test struct {
	PartId     KeyField                     `pk:"part" json:"part_id"`
	SortId     KeyField                     `sk:"sort" json:"sort_id"`
	Nested     *NestedTest                  `json:"nested"`
	TestString DataField[string]            `json:"test_string"`
	TestInt    DataField[int]               `json:"test_int"`
	TestFloat  DataField[float64]           `json:"test_float"`
	TestSlice  DataField[[]string]          `json:"test_slice"`
	TestMap    DataField[map[string]string] `json:"test_map"`
	TestBool   DataField[bool]              `json:"test_bool"`

	TestUnmodified DataField[string]            `json:"test_unmodified"`
	TestNil        DataField[string]            `json:"test_nil"`
	TestNilMap     DataField[map[string]string] `json:"test_nil_map"`
}

type NestedTest struct {
	TestString DataField[string] `json:"test_string"`
	Nested     NestedNestedTest  `json:"nested"`
}

type NestedNestedTest struct {
	TestString DataField[string] `json:"test_string"`
}
