package integration

import (
	"testing"

	"github.com/megakuul/dynamitedb"
)

func checkMigration(t *testing.T, bucket *dynamitedb.Bucket) {
	err := dynamitedb.Create(t.Context(), bucket, &Test{
		PartID:           dynamitedb.Key("workflow"),
		SortID:           dynamitedb.Key("67"),
		TestMigrationMap: dynamitedb.Set(map[string]string{"Hallooo": "Grüessech"}),
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := dynamitedb.Query(t.Context(), bucket, &TestMigrated{
		PartID: dynamitedb.Key("workflow"),
		SortID: dynamitedb.Key("67"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(results[0].TestMigrationMap.Value()) != 1 {
		t.Fatal("test map not properly migrated")
	}
	if len(results[0].TestNewMigrationMap.Value()) != 0 {
		t.Fatal("new test map not properly initialized")
	}
}
