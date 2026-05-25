package dynamitedb

import (
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestModelUpdate(t *testing.T) {
	// prepare
	now := time.Now()

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
		TestString:   Set("Test"),
		TestInt:      Set(1337),
		TestFloat:    Set(4.20),
		TestBool:     Set(false),
		TestTime:     Set(now),
		TestDuration: Set(time.Second * 3),
		TestSlice:    Set([]string{"bombaclad", "ananas", "banana"}),
		TestMap:      Set(map[string]string{"bombaclad": "yes", "ananas": "absolutely", "banana": "yessir"}),

		TestUnmodified: Set("unmodified"),
	}))

	update := &Test{
		PartID: Key("000"),
		SortID: Key("000"),
		Nested: &NestedTest{
			TestString: Set("Updated Nested Test"),
			Nested: NestedNestedTest{
				TestString: Set("Updated Nested Nested Test"),
			},
		},
		TestString:   Set("Updated Test"),
		TestInt:      Multiply(2),
		TestFloat:    Increment(1.0),
		TestBool:     Toggle(),
		TestTime:     Add(time.Hour),
		TestDuration: Increment(time.Hour),
		TestSlice:    Remove("ananas"),
		TestMap:      Set(map[string]string{"updated": "true"}),

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
	if !updated.TestTime.Value().Equal(now.Add(time.Hour)) {
		t.Fatalf("time update does not work properly (got '%v' expected '%v')!",
			updated.TestTime.Value(),
			now.Add(time.Hour),
		)
	}
	if updated.TestDuration.Value() != time.Second*3+time.Hour {
		t.Fatalf("duration update does not work properly (got '%v' expected '%v')!",
			updated.TestDuration.Value(),
			time.Second*3+time.Hour,
		)
	}
	if slices.Contains(updated.TestSlice.Value(), "ananas") {
		t.Fatalf("slice update does not work properly (expected value was not removed)!")
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
