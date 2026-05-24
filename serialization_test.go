package dynamitedb

import (
	"reflect"
	"testing"
)

func TestSerialization(t *testing.T) {
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

	// act
	rawOriginal, err := serialize(original.Interface().(*Test))
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
	if secondUpdate.TestString.Value() != "Test" {
		t.Fatalf("serialization disrupted data fields")
	}

	if !reflect.DeepEqual(firstUpdate, secondUpdate) {
		t.Fatalf("serialization changed the structure")
	}
}
