package dynamitedb

import (
	"reflect"
	"testing"
)

func TestModelUpdate(t *testing.T) {
	// prepare
	original := reflect.New(reflect.TypeFor[Test]())
	initModel(original)
	applyUpdate(original, reflect.ValueOf(&Test{
		PartId: Key("69"),
		SortId: Key("187"),
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

	update := &Test{
		PartId: Key("000"),
		SortId: Key("000"),
		Nested: &NestedTest{
			TestString: Set("Updated Nested Test"),
			Nested: NestedNestedTest{
				TestString: Set("Updated Nested Nested Test"),
			},
		},
		TestString: Set("Updated Test"),
		TestInt:    Mul(2),
		TestFloat:  Inc(1.0),
		TestBool:   Toggle(),
		TestSlice:  Append([]string{"update"}),
		TestMap:    Set(map[string]string{"updated": "true"}),

		TestNil: Set("modified"),
		TestNilMap: Emplace(map[string]string{
			"modified": "true",
		}),
	}

	// act
	rawOriginal, err := serialize(original.Interface().(*Test))
	if err != nil {
		t.Fatalf("failed to serialize original structure: %v", err)
	}
	rawUpdated, err := updateObject(rawOriginal, update)
	if err != nil {
		t.Fatalf("failed to update structure: %v", err)
	}
	updated, err := deserialize[Test](rawUpdated)
	if err != nil {
		t.Fatalf("failed to deserialize updated structure: %v", err)
	}

	// assert
	if updated.TestString.Value() != "Updated Test" {
		t.Fatalf("string update does not work properly (got '%v' expected '%v')!",
			updated.TestString.Value(),
			"Updated Test",
		)
	}
	if updated.TestInt.Value() != 1337*2 {
		t.Fatalf("int update does not work properly (got '%v' expected '%v')!",
			updated.TestInt.Value(),
			1337*2,
		)
	}
	if updated.TestFloat.Value() != 4.20+1 {
		t.Fatalf("float update does not work properly (got '%v' expected '%v')!",
			updated.TestFloat.Value(),
			4.20+1,
		)
	}
	if !updated.TestBool.Value() {
		t.Fatalf("bool update does not work properly (got 'false' expected 'true')!")
	}
	if updated.TestSlice.Value()[len(updated.TestSlice.Value())-1] != "update" {
		t.Fatalf("slice update does not work properly (got '%v' expected '%v')!",
			updated.TestSlice.Value()[0],
			"update",
		)
	}
	if updated.TestMap.Value()["updated"] != "true" {
		t.Fatalf("map update does not work properly (got '%v' expected '%v')!",
			updated.TestMap.Value()["updated"],
			"true",
		)
	}

	if updated.TestUnmodified.Value() != "unmodified" {
		t.Fatalf("original field that should not be updated was updated (got '%v' expected '%v')!",
			updated.TestUnmodified.Value(),
			"unmodified",
		)
	}
	if updated.TestNil.Value() != "modified" {
		t.Fatalf("original nil field that should be updated was not updated (got '%v' expected '%v')!",
			updated.TestNil.Value(),
			"modified",
		)
	}
	if updated.TestNilMap.Value()["modified"] != "true" {
		t.Fatalf("map update does not work properly (got '%v' expected '%v')!",
			updated.TestNilMap.Value()["modified"],
			"true",
		)
	}

	if updated.Nested.Nested.TestString.Value() != "Updated Nested Nested Test" {
		t.Fatalf("nested structural data was not updated properly (got '%v' expected '%v')!",
			updated.Nested.Nested.TestString.Value(),
			"Updated Nested Nested Test",
		)
	}
}
