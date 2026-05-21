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

	passOpFilter := &Test{
		PartId:         Key("69"),
		SortId:         Key("187"),
		TestString:     Includes("es"),
		TestInt:        GreaterThan(1336),
		TestFloat:      LessOrEqThan(4.20),
		TestSlice:      Contains("bombaclad", "ananas"),
		TestMap:        Has("ananas", "absolutely"),
		TestUnmodified: In("ananas", "unmodified", "anotherone"),
	}

	failNestedFilter := &Test{
		PartId: Key("69"),
		SortId: Key("187"),
		Nested: &NestedTest{
			TestString: Eq("Nested Test"),
			Nested: NestedNestedTest{
				TestString: Eq("Nesteddd Nested Test"),
			},
		},
	}
	failInFilter := &Test{
		PartId:    Key("69"),
		SortId:    Key("187"),
		TestSlice: In([]string{"test1"}, []string{"bombacladdd", "ananas", "banana"}, []string{"test2"}),
	}
	failSliceFilter := &Test{
		PartId:     Key("69"),
		SortId:     Key("187"),
		TestString: Eq("Test"),
		TestSlice:  Eq([]string{"bombacladdd", "ananas", "banana"}),
	}
	failContainsFilter := &Test{
		PartId:     Key("69"),
		SortId:     Key("187"),
		TestString: Eq("Test"),
		TestSlice:  Contains("bombacladdd"),
	}
	failMapFilter := &Test{
		PartId:     Key("69"),
		SortId:     Key("187"),
		TestString: Eq("Test"),
		TestMap:    Eq(map[string]string{"bombaclad": "no", "ananas": "absolutely", "banana": "yessir"}),
	}
	failHasFilter := &Test{
		PartId:     Key("69"),
		SortId:     Key("187"),
		TestString: Eq("Test"),
		TestMap:    Has("bombaclad", "no"),
	}

	// assert
	if !checkFilter(reflect.ValueOf(original), reflect.ValueOf(passEmptyFilter)) {
		t.Fatalf("empty filter that should pass didn't pass")
	}

	if !checkFilter(reflect.ValueOf(original), reflect.ValueOf(passEqFilter)) {
		t.Fatalf("eq filter that should pass didn't pass")
	}

	if !checkFilter(reflect.ValueOf(original), reflect.ValueOf(passOpFilter)) {
		t.Fatalf("operation filter that should pass didn't pass")
	}

	if checkFilter(reflect.ValueOf(original), reflect.ValueOf(failNestedFilter)) {
		t.Fatalf("nested eq filter that shouldn't pass did pass")
	}

	if checkFilter(reflect.ValueOf(original), reflect.ValueOf(failInFilter)) {
		t.Fatalf("in filter that shouldn't pass did pass")
	}

	if checkFilter(reflect.ValueOf(original), reflect.ValueOf(failSliceFilter)) {
		t.Fatalf("slice filter that shouldn't pass did pass")
	}

	if checkFilter(reflect.ValueOf(original), reflect.ValueOf(failContainsFilter)) {
		t.Fatalf("contains filter that shouldn't pass did pass")
	}

	if checkFilter(reflect.ValueOf(original), reflect.ValueOf(failMapFilter)) {
		t.Fatalf("map filter that shouldn't pass did pass")
	}

	if checkFilter(reflect.ValueOf(original), reflect.ValueOf(failHasFilter)) {
		t.Fatalf("has filter that shouldn't pass did pass")
	}
}
