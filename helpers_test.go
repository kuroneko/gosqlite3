package sqlite3

import "fmt"
import "gob"
import "os"
import "testing"


var FOO	*Table
var BAR *Table

func init() {
	FOO = &Table{ "foo", "number INTEGER, text VARCHAR(20)" }
 	BAR = &Table{ "bar", "number INTEGER, value BLOB" }
}

type TwoItems struct {
	Number		string
	Text		string
}

func (t *TwoItems) String() string {
	return "[" + t.Number + " : " + t.Text + "]"
}

func fatalOnError(t *testing.T, e os.Error, message string, parameters... interface{}) {
	if e != nil {
		t.Fatalf("%v : %v", e, fmt.Sprintf(message, parameters...))
	}
}

func (db *Database) stepThroughRows(t *testing.T, table *Table) (c int) {
	var e	os.Error
	sql := fmt.Sprintf("SELECT * from %v;", table.Name)
	c, e = db.Execute(sql, func(st *Statement, values ...interface{}) {
		data := values[1]
		switch data := data.(type) {
		case *gob.Decoder:
			blob := &TwoItems{}
			data.Decode(blob)
			t.Logf("BLOB =>   %v: %v, %v: %v\n", ResultColumn(0).Name(st), ResultColumn(0).Value(st), st.ColumnName(1), blob)
		default:
			t.Logf("TEXT => %v: %v, %v: %v\n", ResultColumn(0).Name(st), ResultColumn(0).Value(st), st.ColumnName(1), st.Column(1))
		}
	})
	fatalOnError(t, e, "%v failed on step %v", sql, c)
	if rows, _ := table.Rows(db); rows != c {
		t.Fatalf("%v: %v rows expected, %v rows found", table.Name, rows, c)
	}
	return
}

func (db *Database) runQuery(t *testing.T, sql string, params... interface{}) {
	st, e := db.Prepare(sql, params...)
	fatalOnError(t, e, st.SQLSource())
	st.Step()
	st.Finalize()
}

func (db *Database) populate(t *testing.T, table *Table) {
	switch table.Name {
	case "foo":
		db.runQuery(t, "INSERT INTO foo values (1, 'this is a test')")
		db.runQuery(t, "INSERT INTO foo values (?, ?)", 2, "holy moly")
		if c, _ := table.Rows(db); c != 2 {
			t.Fatal("Failed to populate %v", table.Name)
		}
	case "bar":
		db.runQuery(t, "INSERT INTO bar values (1, 'this is a test')")
		db.runQuery(t, "INSERT INTO bar values (?, ?)", 2, "holy moly")
		db.runQuery(t, "INSERT INTO bar values (?, ?)", 3, TwoItems{ "holy moly", "guacomole" })
		if c, _ := table.Rows(db); c != 3 {
			t.Fatal("Failed to populate %v", table.Name)
		}
	}
}