package dynamitdb

import (
	"reflect"
	"testing"

	"github.com/megakuul/dynamitdb/data"
	"github.com/megakuul/dynamitdb/key"
	"github.com/megakuul/dynamitdb/update"
)

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

func TestGetModel(t *testing.T) {
	// prepare
	original := &Test{
		PartId: key.New("69"),
		SortId: key.New("187"),
		Nested: &NestedTest{
			TestString: data.New("Nested Test"),
			Nested: NestedNestedTest{
				TestString: data.New("Nested Nested Test"),
			},
		},
		TestString: data.New("Test"),
		TestInt:    data.New(1337),
		TestFloat:  data.New(4.20),
		TestBool:   data.New(false),
		TestSlice:  data.New([]string{"bombaclad", "ananas", "banana"}),
		TestMap:    data.New(map[string]string{"bombaclad": "yes", "ananas": "absolutely", "banana": "yessir"}),

		TestUnmodified: data.New("unmodified"),
	}

	update := &Test{
		PartId: key.New("000"),
		SortId: key.New("000"),
		Nested: &NestedTest{
			TestString: update.Set("Updated Nested Test"),
			Nested: NestedNestedTest{
				TestString: update.Set("Updated Nested Nested Test"),
			},
		},
		TestString: update.Set("Updated Test"),
		TestInt:    update.Mul(2),
		TestFloat:  update.Inc(1.0),
		TestBool:   update.Toggle(),
		TestSlice:  update.Append([]string{"update"}),
		TestMap:    update.Set(map[string]string{"updated": "true"}),

		TestNil: update.Set("modified"),
		TestNilMap: update.Emplace(map[string]string{
			"modified": "true",
		}),
	}

	// act
	applyUpdate(reflect.ValueOf(original), reflect.ValueOf(update))

	// assert
	if original.PartId.Value() != "69" || original.SortId.Value() != "187" {
		t.Fatalf("updateModel modified keyFields this is not allowed!")
	}

	if original.TestString.Value() != "Updated Test" {
		t.Fatalf("string update does not work properly (got '%v' expected '%v')!",
			original.TestString.Value(),
			"Updated Test",
		)
	}
	if original.TestInt.Value() != 1337*2 {
		t.Fatalf("int update does not work properly (got '%v' expected '%v')!",
			original.TestInt.Value(),
			1337*2,
		)
	}
	if original.TestFloat.Value() != 4.20+1 {
		t.Fatalf("float update does not work properly (got '%v' expected '%v')!",
			original.TestFloat.Value(),
			4.20+1,
		)
	}
	if !original.TestBool.Value() {
		t.Fatalf("bool update does not work properly (got 'false' expected 'true')!")
	}
	if original.TestSlice.Value()[len(original.TestSlice.Value())-1] != "update" {
		t.Fatalf("slice update does not work properly (got '%v' expected '%v')!",
			original.TestSlice.Value()[0],
			"update",
		)
	}
	if original.TestMap.Value()["updated"] != "true" {
		t.Fatalf("map update does not work properly (got '%v' expected '%v')!",
			original.TestMap.Value()["updated"],
			"true",
		)
	}

	if original.TestUnmodified.Value() != "unmodified" {
		t.Fatalf("original field that should not be updated was updated (got '%v' expected '%v')!",
			original.TestUnmodified.Value(),
			"unmodified",
		)
	}
	if original.TestNil.Value() != "modified" {
		t.Fatalf("original nil field that should be updated was not updated (got '%v' expected '%v')!",
			original.TestNil.Value(),
			"modified",
		)
	}
	if original.TestNilMap.Value()["modified"] != "true" {
		t.Fatalf("map update does not work properly (got '%v' expected '%v')!",
			original.TestNilMap.Value()["modified"],
			"true",
		)
	}

	if original.Nested.Nested.TestString.Value() != "Updated Nested Nested Test" {
		t.Fatalf("nested structural data was not updated properly (got '%v' expected '%v')!",
			original.Nested.Nested.TestString.Value(),
			"Updated Nested Nested Test",
		)
	}
}
