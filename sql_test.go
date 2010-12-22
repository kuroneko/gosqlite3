package sql_test

import "testing"
import "sqlite3"

func TestGeneral(t *testing.T) {
	sqlite3.Initialize()
	defer sqlite3.Shutdown()
	t.Logf("Sqlite3 Version: %v\n", sqlite3.LibVersion())

	if db, e := sqlite3.Open("test.db"); e != sqlite3.OK {
		t.Errorf("Open test.db: %v", e)
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

func TestSession(t *testing.T) {
	sqlite3.Session("test.db", func(db *sqlite3.Database) {
		t.Logf("Sqlite3 Version: %v\n", sqlite3.LibVersion())
		if st, e := db.Prepare("DROP TABLE IF EXISTS foo;"); e == sqlite3.OK {
			st.Step()
			st.Finalize()
		} else {
			t.Errorf("DROP TABLE IF EXISTS foo; failed to compile: %v", e)
		}

		if st, e := db.Prepare("CREATE TABLE foo (i INTEGER, s VARCHAR(20));"); e == sqlite3.OK {
			st.Step()
			st.Finalize()
		} else {
			t.Errorf("CREATE TABLE foo (i INTEGER, s VARCHAR(20)); failed to compile: %v", e)
		}

		if st, e := db.Prepare("INSERT INTO foo values (2, 'this is a test')"); e == sqlite3.OK {
			st.Step()
			st.Finalize()
		} else {
			t.Errorf("INSERT INTO foo values (2, 'this is a test') failed to compile: %v", e)
		}

		if st, e := db.Prepare("INSERT INTO foo values (?, ?)"); e == sqlite3.OK {
			st.Bind(1, 3)
			st.Bind(2, "holy moly")
			st.Step()
			st.Finalize()
		} else {
			t.Errorf("INSERT INTO foo values (3, \"holy moly\") failed to compile: %v", e)
		}

		t.Logf("%v changes\n", db.TotalChanges())
		t.Logf("last insert id: %v\n", db.LastInsertRowID())

		if st, e := db.Prepare("INSERT INTO foo values (?, ?)"); e == sqlite3.OK {
			st.Bind(1, 4, "holy moly guacamole")
			st.Step()
			st.Finalize()
		} else {
			t.Errorf("INSERT INTO foo values (4, \"holy moly guacamole\") failed to compile: %v", e)
		}

		t.Logf("%v changes\n", db.TotalChanges())
		t.Logf("last insert id: %v\n", db.LastInsertRowID())

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