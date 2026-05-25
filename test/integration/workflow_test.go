package integration

import (
	"testing"
	"time"

	"github.com/megakuul/dynamitedb"
)

func checkWorkflow(t *testing.T, bucket *dynamitedb.Bucket) {
	now := time.Now()
	err := dynamitedb.Create(t.Context(), bucket, &Test{
		PartID:     dynamitedb.Key("workflow"),
		SortID:     dynamitedb.Key("69"),
		TestString: dynamitedb.Set("Bombaclad"),
		Nested: &NestedTest{
			TestString: dynamitedb.Set("Nested Test"),
			Nested:     NestedNestedTest{TestString: dynamitedb.Set("Nested Nested Test")},
		},
		TestInt:      dynamitedb.Set(69),
		TestFloat:    dynamitedb.Set(42.0),
		TestSlice:    dynamitedb.Append("something", "anotherthing"),
		TestMap:      dynamitedb.Emplace(map[string]string{"test": "true"}),
		TestTime:     dynamitedb.Set(now),
		TestDuration: dynamitedb.Set(time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = dynamitedb.Update(t.Context(), bucket, &Test{
		PartID:     dynamitedb.Key("workflow"),
		SortID:     dynamitedb.Key("69"),
		TestString: dynamitedb.Set("Updated Bombaclad"),
		Nested: &NestedTest{
			TestString: dynamitedb.Set("Updated Nested Test"),
			Nested:     NestedNestedTest{TestString: dynamitedb.Set("Updated Nested Nested Test")},
		},
		TestInt:      dynamitedb.Increment(-70),
		TestFloat:    dynamitedb.Multiply(2.0),
		TestSlice:    dynamitedb.Append("updatedthing"),
		TestMap:      dynamitedb.Emplace(map[string]string{"test": "false"}),
		TestTime:     dynamitedb.Add(time.Hour),
		TestDuration: dynamitedb.Increment(time.Hour),
		TestBool:     dynamitedb.Toggle(),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = dynamitedb.Get(t.Context(), bucket, &Test{
		PartID:     dynamitedb.Key("workflow"),
		SortID:     dynamitedb.Key("69"),
		TestString: dynamitedb.Includes("pdated Bombacla"),
		Nested: &NestedTest{
			TestString: dynamitedb.Eq("Updated Nested Test"),
			Nested:     NestedNestedTest{TestString: dynamitedb.Eq("Updated Nested Nested Test")},
		},
		TestInt:      dynamitedb.LessThan(1),
		TestFloat:    dynamitedb.NotEq(42.0),
		TestSlice:    dynamitedb.Contains("updatedthing", "something"),
		TestMap:      dynamitedb.Has("test", "false"),
		TestBool:     dynamitedb.Eq(true),
		TestTime:     dynamitedb.After(now),
		TestDuration: dynamitedb.LessOrEqThan(time.Hour * 2),
		TestNil:      dynamitedb.Eq(""),
		TestNilMap:   dynamitedb.Eq(map[string]string{}),
	})
	if err != nil {
		t.Fatal(err)
	}
}
