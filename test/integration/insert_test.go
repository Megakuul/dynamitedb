package integration

import (
	"testing"

	"github.com/megakuul/dynamitedb"
)

func checkInserts(t *testing.T, bucket *dynamitedb.Bucket) {
	err := dynamitedb.Create(t.Context(), bucket, &Test{
		PartId:     dynamitedb.Key("insert"),
		SortId:     dynamitedb.Key("1337"),
		TestString: dynamitedb.Set("Santa Clause"),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = dynamitedb.Create(t.Context(), bucket, &Test{
		PartId:     dynamitedb.Key("insert"),
		SortId:     dynamitedb.Key("1337"),
		TestString: dynamitedb.Set("The Oompa-Loompas"),
	})
	if err == nil {
		t.Fatalf("double create insert should fail but it didn't fail...")
	}

	err = dynamitedb.Put(t.Context(), bucket, &Test{
		PartId:     dynamitedb.Key("insert"),
		SortId:     dynamitedb.Key("1337"),
		TestString: dynamitedb.Set("Willy Wonka"),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = dynamitedb.Get(t.Context(), bucket, &Test{
		PartId:     dynamitedb.Key("insert"),
		SortId:     dynamitedb.Key("1337"),
		TestString: dynamitedb.Eq("Willy Wonka"),
	})
	if err != nil {
		t.Fatal(err)
	}
}
