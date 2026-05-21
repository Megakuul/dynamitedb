package dynamitedb

import (
	"reflect"
	"testing"
)

func TestSerialization(t *testing.T) {
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

	// act
	rawOriginal, err := serialize(original)
	if err != nil {
		t.Fatalf("failed to serialize structure: %v", err)
	}
	firstUpdate, err := deserialize[Test](rawOriginal)
	if err != nil {
		t.Fatalf("failed to deserialize structure: %v", err)
	}
	rawFirst, err := serialize(firstUpdate)
	if err != nil {
		t.Fatalf("failed to serialize structure: %v", err)
	}
	secondUpdate, err := deserialize[Test](rawFirst)
	if err != nil {
		t.Fatalf("failed to deserialize structure: %v", err)
	}

	// assert
	if !reflect.DeepEqual(firstUpdate, secondUpdate) {
		t.Fatalf("serialization changed the structure")
	}
}
