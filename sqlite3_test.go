package sqlite3

import (
	"os"
	"testing"
)

func TestSession(t *testing.T) {
	Session("test.db", func(db *Database) {
		FOO.Drop(db)
		FOO.Create(db)
		db.runQuery(t, "INSERT INTO foo values (1, 'this is a test')")
		db.runQuery(t, "INSERT INTO foo values (?, ?)", 2, "holy moly")
		db.stepThroughRows(t, FOO)
	})
}

func TestTransientSession(t *testing.T) {
	TransientSession(func(db *Database) {
		FOO.Drop(db)
		FOO.Create(db)
		db.runQuery(t, "INSERT INTO foo values (1, 'this is a test')")
		db.runQuery(t, "INSERT INTO foo values (?, ?)", 2, "holy moly")
		db.stepThroughRows(t, FOO)
	})
}

func TestOpen(t *testing.T) {
	//	Test for issue #13

	//	Create a new database
	filename := "new.db"
	os.Remove(filename)
	db, e := Open(filename)
	if e != nil {
		t.Fatalf("Creating %v failed with error: %v", db, e)
	}
	if _, e = db.Execute( "CREATE TABLE foo (id INTEGER PRIMARY KEY ASC, name VARCHAR(10));" ); e != nil {
		t.Fatalf("Create Table foo failed with error: %v", e)
	}
	db.Close()

	if _, e := os.Stat(filename); e != nil {
		t.Fatalf("Checking %v existence failed with error: %v", filename, e)
	}

	//	If new.db already exists and is a valid SQLite3 database this should succeed
	if db, e = Open(filename); e != nil {
		t.Fatalf("Reopening %v failed with error: %v", db, e)
	}
	defer db.Close()
	if _, e = db.Execute( "INSERT INTO foo (id,name) VALUES ('1', 'John');" ); e != nil {
		t.Fatalf("Insert into foo failed with error: %v", e)
	}
}