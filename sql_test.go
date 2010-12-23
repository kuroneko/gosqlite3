package sql_test

import "testing"
import "sqlite3"

func TestGeneral(t *testing.T) {
	sqlite3.Initialize()
	defer sqlite3.Shutdown()
	t.Logf("Sqlite3 Version: %v\n", sqlite3.LibVersion())

	if db, e := sqlite3.Open(":memory:"); e != sqlite3.OK {
		t.Fatalf("Open :memory:: %v", e)
	} else {
		defer db.Close()

		t.Logf("Database opened: %v [flags: %v]", db.Filename, int(db.Flags))
		if st, e := db.Prepare("CREATE TABLE foo (i INTEGER, s VARCHAR(20));"); e != sqlite3.OK {
			t.Errorf("Create Table: %v", e)
		} else {
			defer st.Finalize()
			st.Step()
			if st, e := db.Prepare("DROP TABLE foo;"); e != sqlite3.OK {
				t.Errorf("Drop Table: %v", e)
			} else {
				defer st.Finalize()
				st.Step()
			}
		}
	}
}

var queries = []struct {
	sql     string
	params  [][]interface{}
	verbose bool
}{
	{"DROP TABLE IF EXISTS foo;", nil, false},
	{"CREATE TABLE foo (i INTEGER, s VARCHAR(20));", nil, false},
	{"INSERT INTO foo values (2, 'this is a test')", nil, false},
	{"INSERT INTO foo values (?, ?)", [][]interface{}{{3}, {"holy moly"}}, true},
	{"INSERT INTO foo values (?, ?)", [][]interface{}{{4, "holy moly guacamole"}}, true},
}

func TestSession(t *testing.T) {
	sqlite3.Session(":memory:", func(db *sqlite3.Database) {
		t.Logf("Sqlite3 Version: %v\n", sqlite3.LibVersion())

		for q, query := range queries {
			if st, e := db.Prepare(query.sql); e == sqlite3.OK {
				if query.params != nil {
					for p, params := range query.params {
						st.Bind(p+1, params)
					}
				}
				st.Step()
				st.Finalize()
			} else {
				t.Errorf("queries[%v] \"%v\" failed to compile:\n\t%v",
					q, query.sql, e)
			}

			if query.verbose {
				t.Logf("%v changes\n", db.TotalChanges())
				t.Logf("last insert id: %v\n", db.LastInsertRowID())
			}
		}

		if st, e := db.Prepare("SELECT * from foo limit 5;"); e == sqlite3.OK {
			for i := 0; ; i++ {
				switch st.Step() {
				case sqlite3.DONE:
					return
				case sqlite3.ROW:
					t.Logf("data: %v, %v\n", st.Column(0), st.Column(1))
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
