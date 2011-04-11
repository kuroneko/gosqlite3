package main

import "fmt"
import "github.com/kuroneko/sqlite3"

func main() {
	sqlite3.Initialize();
	defer sqlite3.Shutdown();
	fmt.Printf("Sqlite3 Version: %s\n", sqlite3.LibVersion());

	dbh := new(sqlite3.Handle);
	dbh.Open("test.db");
	defer dbh.Close();

	st,err := dbh.Prepare("DROP TABLE IF EXISTS foo;");
    if err != "" {
        println(err);
    }
    st.Step();
    st.Finalize();

	st,err = dbh.Prepare("CREATE TABLE foo (i INTEGER, s VARCHAR(20));");
	if (err != "") {
		println(err);
	} else {
		defer st.Finalize();
		st.Step();
	}
	
	st,err = dbh.Prepare("INSERT INTO foo values (2, 'this is a test')");
	if (err != "") {
		println(err);
	} else {
		st.Step();
		st.Finalize();
	}

	st,err = dbh.Prepare("INSERT INTO foo values (?, ?)");
	if (err != "") {
		println(err);
	} else {
        st.BindInt(1, 3);
        st.BindText(2, "holy moly");
		st.Step();
		st.Finalize();
	}
	
	fmt.Printf("%d changes\n", dbh.TotalChanges());
	
	fmt.Printf("last insert id: %d\n", dbh.LastInsertRowID());

	st,err = dbh.Prepare("SELECT * from foo limit 5;");
	if (err != "") {
		println(err);
	} else {
		v, c, n:= "", 0, "";
        func () {
            for {
                c = st.Step();
                switch {
                case c==sqlite3.SQLITE_DONE:
                    return;
                case c==sqlite3.SQLITE_ROW:
                    n, v = st.ColumnText(0), st.ColumnText(1);
                    fmt.Printf("data: %s, %s\n", n, v);
                default:
                    println(dbh.ErrMsg());
                    return;
                };
            }
        }();
		st.Finalize();
	}
	
	
}
