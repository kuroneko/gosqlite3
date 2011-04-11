package sqlite3

import "testing"

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