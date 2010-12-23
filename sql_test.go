package sql_test

import "fmt"
import "testing"
import "sqlite3"

func TestGeneral(t *testing.T) {
	filename := ":memory:"

	sqlite3.Initialize()
	defer sqlite3.Shutdown()
	t.Logf("Sqlite3 Version: %v\n", sqlite3.LibVersion())
	
	db, e := sqlite3.Open(filename)
	if e != sqlite3.OK {
		t.Fatalf("Open %v: %v", filename, e)
	}
	defer db.Close()
	t.Logf("Database opened: %v [flags: %v]", db.Filename, int(db.Flags))
	t.Logf(fmt.Sprintf("Returning status: %v", e))

	st, e := db.Prepare("CREATE TABLE foo (i INTEGER, s VARCHAR(20));")
	if e != sqlite3.OK {
		t.Fatalf("Create Table: %v", e)
	}
	defer st.Finalize()
	st.Step()
	
	st, e = db.Prepare("DROP TABLE foo;")
	if e != sqlite3.OK {
		t.Fatalf("Drop Table: %v", e)
	}
	defer st.Finalize()
	st.Step()
}

func runQuery(t *testing.T, db *sqlite3.Database, sql string, params... interface{}) {
	if st, e := db.Prepare(sql); e == sqlite3.OK {
		t.Logf("successfully compiled %v\n", st.Source())
		st.Bind(1, params...)
		st.Step()
		st.Finalize()
	} else {
		t.Errorf("Error: failed to compile %v", st.Source())
	}
	t.Logf("last insert id: %v\n", db.LastInsertRowID())
	t.Logf("%v changes\n", db.TotalChanges())
}

func TestSession(t *testing.T) {
	sqlite3.Session(":memory:", func(db *sqlite3.Database) {
		t.Logf("Sqlite3 Version: %v\n", sqlite3.LibVersion())

		runQuery(t, db, "DROP TABLE IF EXISTS foo;")
		runQuery(t, db, "CREATE TABLE foo (number INTEGER, text VARCHAR(20));")
		runQuery(t, db, "INSERT INTO foo values (1, 'this is a test')")
		runQuery(t, db, "INSERT INTO foo values (?, ?)", 2, "holy moly")
		runQuery(t, db, "INSERT INTO foo values (?, ?)", []interface{}{ 3, "holy moly guacomole" })

		if st, e := db.Prepare("SELECT * from foo limit 5;"); e == sqlite3.OK {
			for i := 0; ; i++ {
				switch st.Step() {
				case sqlite3.DONE:
					return
				case sqlite3.ROW:
					t.Logf("%v: %v, %v: %v\n", st.ColumnName(0), st.Column(0), st.ColumnName(1), st.Column(1))
				default:
					t.Errorf("SELECT * from foo limit 5; failed on step %v: %v", i, db.Error())
					return
				}
			}
			st.Finalize()
		} else {
			t.Errorf("SELECT * from foo limit 5; failed to return results %v", db.Error())
		}
	})
}
