package integration

import (
	"errors"
	"fmt"
	"testing"

	"github.com/megakuul/dynamitedb"
)

func checkETag(t *testing.T, bucket *dynamitedb.Bucket) {
	err := dynamitedb.Create(t.Context(), bucket, &Test{
		PartID:     dynamitedb.Key("etag"),
		SortID:     dynamitedb.Key("420"),
		TestString: dynamitedb.Set("Patrick Bang"),
		TestBool:   dynamitedb.Set(false),
	})
	if err != nil {
		t.Fatal(err)
	}

	etaggedTests, err := dynamitedb.Query(t.Context(), bucket, &Test{
		PartID: dynamitedb.Key("etag"),
		SortID: dynamitedb.Key("420"),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = dynamitedb.Update(t.Context(), bucket, &Test{
		PartID:     dynamitedb.Key(etaggedTests[0].PartID.Value()),
		SortID:     dynamitedb.Key(etaggedTests[0].SortID.Value()),
		ETag:       etaggedTests[0].ETag,
		TestString: dynamitedb.Set("SpongeBOZZ"),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = dynamitedb.Delete(t.Context(), bucket, &Test{
		PartID: dynamitedb.Key(etaggedTests[0].PartID.Value()),
		SortID: dynamitedb.Key(etaggedTests[0].SortID.Value()),
		ETag:   etaggedTests[0].ETag,
	})
	if err == nil || !errors.Is(err, dynamitedb.ErrConcurrencyConflict) {
		println(fmt.Sprint(err))
		t.Fatalf("delete operation is supposed to return a optimistic lock failure")
	}

	etaggedTest, err := dynamitedb.Get(t.Context(), bucket, &Test{
		PartID: dynamitedb.Key("etag"),
		SortID: dynamitedb.Key("420"),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = dynamitedb.Delete(t.Context(), bucket, &Test{
		PartID: dynamitedb.Key(etaggedTest.PartID.Value()),
		SortID: dynamitedb.Key(etaggedTest.SortID.Value()),
		ETag:   etaggedTest.ETag,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = dynamitedb.Create(t.Context(), bucket, &Test{
		PartID:     dynamitedb.Key(etaggedTest.PartID.Value()),
		SortID:     dynamitedb.Key(etaggedTest.SortID.Value()),
		TestString: dynamitedb.Set("Es ist Juri"),
	})
	if err != nil {
		t.Fatal(err)
	}

	err = dynamitedb.Update(t.Context(), bucket, &Test{
		PartID:     dynamitedb.Key(etaggedTest.PartID.Value()),
		SortID:     dynamitedb.Key(etaggedTest.SortID.Value()),
		TestString: dynamitedb.Set("DEA"),
		ETag:       etaggedTest.ETag,
	})
	if err == nil || !errors.Is(err, dynamitedb.ErrConcurrencyConflict) {
		t.Fatalf("update operation is supposed to return a optimistic lock failure")
	}

	err = dynamitedb.Put(t.Context(), bucket, &Test{
		PartID:     dynamitedb.Key(etaggedTest.PartID.Value()),
		SortID:     dynamitedb.Key(etaggedTest.SortID.Value()),
		TestString: dynamitedb.Set("Johnny Dünnson"),
		ETag:       etaggedTest.ETag,
	})
	if err == nil || !errors.Is(err, dynamitedb.ErrConcurrencyConflict) {
		t.Fatalf("put operation is supposed to return a optimistic lock failure")
	}
}
