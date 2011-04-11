package sql_test

import "testing"
import "github.com/kuroneko/sqlite3"

func TestGeneral(t *testing.T) {
	sqlite3.Initialize();
	defer sqlite3.Shutdown();
	t.Logf("Sqlite3 Version: %s\n", sqlite3.LibVersion());

	dbh := new(sqlite3.Handle);
	err := dbh.Open("test.db");
	if (err != "") {
		t.Errorf("Open test.db: %s", err);
	}
	defer dbh.Close();

	st,err := dbh.Prepare("CREATE TABLE foo (i INTEGER, s VARCHAR(20));");
	if (err != "") {
		t.Errorf("Create Table: %s", err);
	} else {
		defer st.Finalize();
		st.Step();
	}

	st,err = dbh.Prepare("DROP TABLE foo;");
	if (err != "") {
		t.Errorf("Drop Table: %s", err);
	} else {
		defer st.Finalize();
		st.Step();
	}
}
