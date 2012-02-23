package sqlite3

import "testing"

func TestResultColumn(t *testing.T) {
	Session("test.db", func(db *Database) {
		BAR.Create(db)
		t.Logf("Test cases needed for ResultColumn")
	})
}