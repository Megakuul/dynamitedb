package dynamitedb

import (
	"reflect"
	"testing"
)

func TestModelFilter(t *testing.T) {
	// prepare
	original := reflect.New(reflect.TypeFor[Test]())
	initModel(original)
	applyUpdate(original, reflect.ValueOf(&Test{
		PartID: Key("69"),
		SortID: Key("187"),
		Nested: &NestedTest{
			TestString: Set("Nested Test"),
			Nested: NestedNestedTest{
				TestString: Set("Nested Nested Test"),
			},
		},
		TestString: Set("Test"),
		TestInt:    Set(1337),
		TestFloat:  Set(4.20),
		TestBool:   Set(false),
		TestSlice:  Set([]string{"bombaclad", "ananas", "banana"}),
		TestMap:    Set(map[string]string{"bombaclad": "yes", "ananas": "absolutely", "banana": "yessir"}),

		TestUnmodified: Set("unmodified"),
	}))

	passEmptyFilter := &Test{
		PartID: Key("69"),
		SortID: Key("187"),
	}

	passEqFilter := &Test{
		PartID: Key("69"),
		SortID: Key("187"),
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
		PartID:         Key("69"),
		SortID:         Key("187"),
		TestString:     Includes("es"),
		TestInt:        GreaterThan(1336),
		TestFloat:      LessOrEqThan(4.20),
		TestSlice:      Contains("bombaclad", "ananas"),
		TestMap:        Has("ananas", "absolutely"),
		TestUnmodified: In("ananas", "unmodified", "anotherone"),
	}

	failNestedFilter := &Test{
		PartID: Key("69"),
		SortID: Key("187"),
		Nested: &NestedTest{
			TestString: Eq("Nested Test"),
			Nested: NestedNestedTest{
				TestString: Eq("Nesteddd Nested Test"),
			},
		},
	}
	failInFilter := &Test{
		PartID:    Key("69"),
		SortID:    Key("187"),
		TestSlice: In([]string{"test1"}, []string{"bombacladdd", "ananas", "banana"}, []string{"test2"}),
	}
	failSliceFilter := &Test{
		PartID:     Key("69"),
		SortID:     Key("187"),
		TestString: Eq("Test"),
		TestSlice:  Eq([]string{"bombacladdd", "ananas", "banana"}),
	}
	failContainsFilter := &Test{
		PartID:     Key("69"),
		SortID:     Key("187"),
		TestString: Eq("Test"),
		TestSlice:  Contains("bombacladdd"),
	}
	failMapFilter := &Test{
		PartID:     Key("69"),
		SortID:     Key("187"),
		TestString: Eq("Test"),
		TestMap:    Eq(map[string]string{"bombaclad": "no", "ananas": "absolutely", "banana": "yessir"}),
	}
	failHasFilter := &Test{
		PartID:     Key("69"),
		SortID:     Key("187"),
		TestString: Eq("Test"),
		TestMap:    Has("bombaclad", "no"),
	}

	// assert
	if !checkFilter(original, reflect.ValueOf(passEmptyFilter)) {
		t.Fatalf("empty filter that should pass didn't pass")
	}

	if !checkFilter(original, reflect.ValueOf(passEqFilter)) {
		t.Fatalf("eq filter that should pass didn't pass")
	}

	if !checkFilter(original, reflect.ValueOf(passOpFilter)) {
		t.Fatalf("operation filter that should pass didn't pass")
	}

	if checkFilter(original, reflect.ValueOf(failNestedFilter)) {
		t.Fatalf("nested eq filter that shouldn't pass did pass")
	}

	if checkFilter(original, reflect.ValueOf(failInFilter)) {
		t.Fatalf("in filter that shouldn't pass did pass")
	}

	if checkFilter(original, reflect.ValueOf(failSliceFilter)) {
		t.Fatalf("slice filter that shouldn't pass did pass")
	}

	if checkFilter(original, reflect.ValueOf(failContainsFilter)) {
		t.Fatalf("contains filter that shouldn't pass did pass")
	}

	if checkFilter(original, reflect.ValueOf(failMapFilter)) {
		t.Fatalf("map filter that shouldn't pass did pass")
	}

	if checkFilter(original, reflect.ValueOf(failHasFilter)) {
		t.Fatalf("has filter that shouldn't pass did pass")
	}
}
