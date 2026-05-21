package dynamitdb

import (
	"reflect"
	"testing"
)

func TestModelFilter(t *testing.T) {
	// prepare
	original := &Test{
		PartId: Key("69"),
		SortId: Key("187"),
		Nested: &NestedTest{
			TestString: Data("Nested Test"),
			Nested: NestedNestedTest{
				TestString: Data("Nested Nested Test"),
			},
		},
		TestString: Data("Test"),
		TestInt:    Data(1337),
		TestFloat:  Data(4.20),
		TestBool:   Data(false),
		TestSlice:  Data([]string{"bombaclad", "ananas", "banana"}),
		TestMap:    Data(map[string]string{"bombaclad": "yes", "ananas": "absolutely", "banana": "yessir"}),

		TestUnmodified: Data("unmodified"),
	}

	passEmptyFilter := &Test{
		PartId: Key("69"),
		SortId: Key("187"),
	}

	passEqFilter := &Test{
		PartId: Key("69"),
		SortId: Key("187"),
		Nested: &NestedTest{
			TestString: Eq("Nested Test"),
			Nested: NestedNestedTest{
				TestString: Eq("Nested Nested Test"),
			},
		},
		TestString: Eq("Test"),
		TestInt:    Eq(1337),
		TestFloat:  Eq(4.20),
		TestBool:   Eq(false),
		TestSlice:  Eq([]string{"bombaclad", "ananas", "banana"}),
		TestMap:    Eq(map[string]string{"bombaclad": "yes", "ananas": "absolutely", "banana": "yessir"}),

		TestUnmodified: NotEq("modified"),
	}

	failFilterNested := &Test{
		PartId: Key("69"),
		SortId: Key("187"),
		Nested: &NestedTest{
			TestString: Eq("Nested Test"),
			Nested: NestedNestedTest{
				TestString: Eq("Nesteddd Nested Test"),
			},
		},
	}

	failFilterSlice := &Test{
		PartId:     Key("69"),
		SortId:     Key("187"),
		TestString: Eq("Test"),
		TestSlice:  Eq([]string{"bombacladdd", "ananas", "banana"}),
	}

	failFilterMap := &Test{
		PartId:     Key("69"),
		SortId:     Key("187"),
		TestString: Eq("Test"),
		TestMap:    Eq(map[string]string{"bombaclad": "no", "ananas": "absolutely", "banana": "yessir"}),
	}

	// assert
	if !checkFilter(reflect.ValueOf(original), reflect.ValueOf(passEmptyFilter)) {
		t.Fatalf("empty filter that should pass didn't pass")
	}

	if !checkFilter(reflect.ValueOf(original), reflect.ValueOf(passEqFilter)) {
		t.Fatalf("eq filter that should pass didn't pass")
	}

	if checkFilter(reflect.ValueOf(original), reflect.ValueOf(failFilterNested)) {
		t.Fatalf("nested eq filter that shouldn't pass did pass")
	}

	if checkFilter(reflect.ValueOf(original), reflect.ValueOf(failFilterSlice)) {
		t.Fatalf("slice filter that shouldn't pass did pass")
	}

	if checkFilter(reflect.ValueOf(original), reflect.ValueOf(failFilterMap)) {
		t.Fatalf("map filter that shouldn't pass did pass")
	}
}
