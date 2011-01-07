package sqlite3

import "bytes"
import "fmt"
import "gob"
import "testing"

func TestGeneral(t *testing.T) {
	filename := ":memory:"

	Initialize()
	defer Shutdown()
	t.Logf("Sqlite3 Version: %v\n", LibVersion())
	
	db, e := Open(filename)
	if e != OK {
		t.Fatalf("Open %v: %v", filename, e)
	}
	defer db.Close()
	t.Logf("Database opened: %v [flags: %v]", db.Filename, int(db.Flags))
	t.Logf(fmt.Sprintf("Returning status: %v", e))

	st, e := db.Prepare("CREATE TABLE foo (i INTEGER, s VARCHAR(20));")
	if e != OK {
		t.Fatalf("Create Table: %v", e)
	}
	defer st.Finalize()
	st.Step()
	
	st, e = db.Prepare("DROP TABLE foo;")
	if e != OK {
		t.Fatalf("Drop Table: %v", e)
	}
	defer st.Finalize()
	st.Step()
}

func runQuery(t *testing.T, db *Database, sql string, params... interface{}) {
	if st, e := db.Prepare(sql); e == OK {
		if e, i := st.Bind(1, params...); e == OK {
			st.Step()
		} else {
			t.Errorf("Error: unable to bind column %v resulting in error %v", i , e)
		}
		st.Finalize()
	} else {
		t.Errorf("Error: failed to compile %v", st.SQLSource())
	}
}

func TestSession(t *testing.T) {
	Session(":memory:", func(db *Database) {
		runQuery(t, db, "DROP TABLE IF EXISTS foo;")
		runQuery(t, db, "CREATE TABLE foo (number INTEGER, text VARCHAR(20));")
		runQuery(t, db, "INSERT INTO foo values (1, 'this is a test')")
		runQuery(t, db, "INSERT INTO foo values (?, ?)", 2, "holy moly")

		if st, e := db.Prepare("SELECT * from foo limit 5;"); e == OK {
			for i := 0; ; i++ {
				switch st.Step() {
				case DONE:
					return
				case ROW:
					text := st.Column(1)
					switch text := text.(type) {
					case *gob.Decoder:
						blob := TwoItems{}
						text.Decode(blob)
						t.Logf("%v: %v, %v: %v\n", Column(0).Name(st), Column(0).Value(st), st.ColumnName(1), blob)
					default:
						t.Logf("%v: %v, %v: %v\n", Column(0).Name(st), Column(0).Value(st), st.ColumnName(1), st.Column(1))
					}
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

func TestTransientSession(t *testing.T) {
	TransientSession(func(db *Database) {
		runQuery(t, db, "DROP TABLE IF EXISTS foo;")
		runQuery(t, db, "CREATE TABLE foo (number INTEGER, text VARCHAR(20));")
		runQuery(t, db, "INSERT INTO foo values (1, 'this is a test')")
		runQuery(t, db, "INSERT INTO foo values (?, ?)", 2, "holy moly")

		if st, e := db.Prepare("SELECT * from foo limit 5;"); e == OK {
			for i := 0; ; i++ {
				switch st.Step() {
				case DONE:
					return
				case ROW:
					text := st.Column(1)
					switch text := text.(type) {
					case *gob.Decoder:
						blob := TwoItems{}
						text.Decode(blob)
						t.Logf("%v: %v, %v: %v\n", Column(0).Name(st), Column(0).Value(st), st.ColumnName(1), blob)
					default:
						t.Logf("%v: %v, %v: %v\n", Column(0).Name(st), Column(0).Value(st), st.ColumnName(1), st.Column(1))
					}
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

type TwoItems struct {
	Number		string
	Text		string
}
func (t *TwoItems) String() string {
	return "[" + t.Number + " : " + t.Text + "]"
}

func TestBlobEncoding(t *testing.T) {
	Session("test.db", func(db *Database) {
		runQuery(t, db, "DROP TABLE IF EXISTS foo;")
		runQuery(t, db, "CREATE TABLE foo (number INTEGER, value BLOB);")

		if st, e := db.Prepare("INSERT INTO foo values (3, ?)"); e == OK {
			buffer := new(bytes.Buffer)
			encoder := gob.NewEncoder(buffer)
			if err := encoder.Encode(TwoItems{ "holy moly", "guacomole" }); err != nil {
				t.Errorf("Encoding failed: buffer = %v", )
			} else {
				t.Logf("Encoded data: %v", buffer.Bytes())
				e = Column(1).bind_blob(st, buffer.Bytes())
			}
			if e == OK {
				st.Step()
			} else {
				t.Errorf("Error: unable to bind column 1 resulting in error %v", e)
			}
			st.Finalize()
		} else {
			t.Errorf("Error: failed to compile %v", st.SQLSource())
		}
	})
}

func TestBlob(t *testing.T) {
	Session("test.db", func(db *Database) {
		runQuery(t, db, "DROP TABLE IF EXISTS foo;")
		runQuery(t, db, "CREATE TABLE foo (number INTEGER, value BLOB);")
		runQuery(t, db, "INSERT INTO foo values (1, 'this is a test')")
		runQuery(t, db, "INSERT INTO foo values (?, ?)", 2, "holy moly")
		runQuery(t, db, "INSERT INTO foo values (?, ?)", 3, TwoItems{ "holy moly", "guacomole" })

		if st, e := db.Prepare("SELECT * from foo limit 5;"); e == OK {
			for i := 0; ; i++ {
				switch st.Step() {
				case DONE:
					return
				case ROW:
					text := st.Column(1)
					switch text := text.(type) {
					case *gob.Decoder:
						blob := new(TwoItems)
						if e = text.Decode(blob); e == nil {
							t.Logf("BLOB => %v: %v, %v: %v\n", Column(0).Name(st), Column(0).Value(st), st.ColumnName(1), blob)
						} else {
							t.Logf("BLOB => %v: %v, %v: (decoding failed: %v)\n", Column(0).Name(st), Column(0).Value(st), st.ColumnName(1), blob)
						}
					default:
						t.Logf("TEXT => %v: %v, %v: %v\n", Column(0).Name(st), Column(0).Value(st), st.ColumnName(1), st.Column(1))
					}
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