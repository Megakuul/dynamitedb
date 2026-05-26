package integration

import (
	"errors"
	"testing"

	"github.com/megakuul/dynamitedb"
)

func checkDeletes(t *testing.T, bucket *dynamitedb.Bucket) {
	err := dynamitedb.Create(t.Context(), bucket, &Test{
		PartID:     dynamitedb.Key("delete"),
		SortID:     dynamitedb.Key("1337"),
		TestString: dynamitedb.Set("Santa Clause"),
		TestBool:   dynamitedb.Set(false),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = dynamitedb.Delete(t.Context(), bucket, &Test{
		PartID:   dynamitedb.Key("delete"),
		SortID:   dynamitedb.Key("1337"),
		TestBool: dynamitedb.Eq(true),
	})
	if err != nil && !errors.Is(err, dynamitedb.ErrFilterMismatch) {
		t.Fatal(err)
	}

	_, err = dynamitedb.Get(t.Context(), bucket, &Test{
		PartID: dynamitedb.Key("delete"),
		SortID: dynamitedb.Key("1337"),
	})
	if err != nil {
		t.Fatalf("deletion with blocking filter succeeded: %v", err)
	}

	err = dynamitedb.Delete(t.Context(), bucket, &Test{
		PartID:     dynamitedb.Key("delete"),
		SortID:     dynamitedb.Key("1337"),
		TestString: dynamitedb.Eq("Santa Clause"),
		TestBool:   dynamitedb.Eq(false),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = dynamitedb.Get(t.Context(), bucket, &Test{
		PartID: dynamitedb.Key("delete"),
		SortID: dynamitedb.Key("1337"),
	})
	if !errors.Is(err, dynamitedb.ErrNotFound) {
		t.Fatalf("deletion with passing filter failed")
	}
}
