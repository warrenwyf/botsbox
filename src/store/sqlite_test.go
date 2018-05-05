package store

import (
	"path"
	"testing"
)

var sqliteStore Store = &SqliteStore{
	FilePath: path.Join("/tmp", "botsbox-test.db"),
}

func Test_Init(t *testing.T) {
	err := sqliteStore.Init()
	if err != nil {
		t.Fatalf("SqliteStore.Init() failed: %v", err.Error())
	}
}

func Test_CreateDataset(t *testing.T) {
	fieldNames := []string{"key", "value", "time"}
	fieldTypes := []string{"text", "integer", "timestamp DEFAULT CURRENT_TIMESTAMP"}
	err := sqliteStore.CreateDataset("$unittest", fieldNames, fieldTypes)
	if err != nil {
		t.Fatalf("SqliteStore.CreateTable() failed: %v", err.Error())
	}
}

func Test_HasDataset(t *testing.T) {
	if !sqliteStore.HasDataset("$unittest") {
		t.Fatalf("SqliteStore.HasDataset(%s) should be true", "$unittest")
	}

	if sqliteStore.HasDataset("$foo") {
		t.Fatalf("SqliteStore.HasDataset(%s) should be false", "$foo")
	}
}

func Test_InsertObject(t *testing.T) {
	fields := []string{"key", "value"}
	values := []interface{}{"foo", 2, "none"}
	oid, err := sqliteStore.InsertObject("$unittest", fields, values)
	if err != nil {
		t.Fatalf("SqliteStore.CreateObject() failed: %v", err.Error())
	} else {
		t.Logf("Autogenerated ID: %v", oid)
	}
}

func Test_EmptyDataset(t *testing.T) {
	err := sqliteStore.EmptyDataset("$unittest")
	if err != nil {
		t.Fatalf("SqliteStore.EmptyDataset() failed: %v", err.Error())
	}
}

func Test_Destroy(t *testing.T) {
	err := sqliteStore.Destroy()
	if err != nil {
		t.Fatalf("SqliteStore.Destroy() failed: %v", err.Error())
	}
}
