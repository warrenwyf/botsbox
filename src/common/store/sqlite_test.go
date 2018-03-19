package store

import (
	"testing"
)

var sqliteStore Store = &SqliteStore{
	FilePath: "/tmp/botsbox-test.db",
}

func Test_Init(t *testing.T) {
	err := sqliteStore.Init()
	if err != nil {
		t.Fatalf("SqliteStore.Init() failed: %v", err.Error())
	}
}

func Test_CreateDataset(t *testing.T) {
	fieldNames := []string{"key", "value"}
	fieldTypes := []string{"text", "integer"}
	err := sqliteStore.CreateDataset("unittest", fieldNames, fieldTypes)
	if err != nil { // Maybe dataset is already exists, do not fatal
		t.Logf("SqliteStore.CreateTable() failed: %v", err.Error())
	}
}

func Test_InsertObject(t *testing.T) {
	fields := []string{"key", "value"}
	values := []interface{}{"foo", 2, "none"}
	oid, err := sqliteStore.InsertObject("unittest", fields, values)
	if err != nil {
		t.Fatalf("SqliteStore.CreateObject() failed: %v", err.Error())
	} else {
		t.Logf("Autogenerated ID: %v", oid)
	}
}

func Test_Destroy(t *testing.T) {
	err := sqliteStore.Destroy()
	if err != nil {
		t.Fatalf("SqliteStore.Destroy() failed: %v", err.Error())
	}
}
