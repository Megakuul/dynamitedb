package integration

import (
	"fmt"
	"testing"

	"github.com/megakuul/dynamitedb"
)

func checkQueries(t *testing.T, bucket *dynamitedb.Bucket) {
	for i := range 100 {
		prefix := "odd"
		if i%2 == 0 {
			prefix = "even"
		}

		err := dynamitedb.Create(t.Context(), bucket, &Test{
			PartId:     dynamitedb.Key("workflow"),
			SortId:     dynamitedb.Key(fmt.Sprintf("%s-%d", prefix, i)),
			TestString: dynamitedb.Set(fmt.Sprintf("Bombaclad %d", i)),
			Nested: &NestedTest{
				TestString: dynamitedb.Set("Nested Test"),
				Nested:     NestedNestedTest{TestString: dynamitedb.Set("Nested Nested Test")},
			},
			TestInt: dynamitedb.Set(i),
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	results, err := dynamitedb.Query(t.Context(), bucket, &Test{
		PartId: dynamitedb.Key("workflow"),
		SortId: dynamitedb.KeyPrefix("even-"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 50 {
		t.Fatalf("query failed to properly index entries")
	}
	if results[0].TestInt.Value() != 0 && results[49].TestInt.Value() != 49 {
		t.Fatalf("query sorting was incorrect")
	}

	results, err = dynamitedb.Query(t.Context(), bucket, &Test{
		PartId: dynamitedb.Key("workflow"),
		SortId: dynamitedb.KeyPrefix("even-"),
	}, dynamitedb.WithStartAfter(results[49])) // starting after even-49 should yield nothing (because its lexically sorted)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("query failed to properly start after even entries")
	}

	results, err = dynamitedb.Query(t.Context(), bucket, &Test{
		PartId: dynamitedb.Key("workflow"),
		SortId: dynamitedb.KeyPrefix("odd-"),
	}, dynamitedb.WithLimit(20))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 20 {
		t.Fatalf("query failed to properly limit entries")
	}

	results, err = dynamitedb.Query(t.Context(), bucket, &Test{
		PartId:     dynamitedb.Key("workflow"),
		SortId:     dynamitedb.KeyPrefix("even-"),
		TestString: dynamitedb.Includes("Bombaclad"),
		TestInt: dynamitedb.CustomFilter(func(test int) bool {
			return test%2 == 0 // ensure they are even
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 50 {
		t.Fatalf("query failed to properly scan entries")
	}
}
