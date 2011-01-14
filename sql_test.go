package sqlite3

import "bytes"
import "gob"
import "testing"


func TestGeneral(t *testing.T) {
	Initialize()
	defer Shutdown()
	t.Logf("Sqlite3 Version: %v\n", LibVersion())

	filename := ":memory:"
	db, e := Open(filename)
	fatalOnError(t, e, "opening %v", filename)

	defer db.Close()
	t.Logf("Database opened: %v [flags: %v]", db.Filename, int(db.Flags))
	t.Logf("Returning status: %v", e)
}


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

func TestBlob(t *testing.T) {
	Session("test.db", func(db *Database) {
		BAR.Drop(db)
		BAR.Create(db)

		buffer := new(bytes.Buffer)
		encoder := gob.NewEncoder(buffer)
		fatalOnError(t, encoder.Encode(TwoItems{ "holy", "moly guacomole" }), "Encoding failed: buffer = %v", buffer)
		t.Logf("Encoded data: %v", buffer.Bytes())

		db.runQuery(t, "INSERT INTO bar values (?, ?)", 1, TwoItems{ "holy moly", "guacomole" })
		db.stepThroughRows(t, BAR)
	})
}

func TestTransfers(t *testing.T) {
	TransientSession(func(source *Database) {
		tables := []*Table{ FOO, BAR }
		for _, table := range tables {
			table.Drop(source)
			table.Create(source)
			if c, _ := table.Rows(source); c != 0 {
				t.Fatalf("%v already contains data", table.Name)
			}
		}
		source.runQuery(t, "INSERT INTO foo values (1, 'this is a test')")
		source.runQuery(t, "INSERT INTO foo values (?, ?)", 2, "holy moly")
		source.runQuery(t, "INSERT INTO bar values (?, ?)", 1, TwoItems{ "holy moly", "guacomole" })
		source.stepThroughRows(t, FOO)
		source.stepThroughRows(t, BAR)

		Session("target.db", func(target *Database) {
			t.Logf("Database opened: %v [flags: %v]", target.Filename, int(target.Flags))
			tables := []*Table{ FOO, BAR }
			for _, table := range tables {
				table.Drop(target)
				table.Create(target)
			}
			fatalOnError(t, target.Load(source, "main"), "loading from %v[%]", source.Filename, "main")
			for _, table := range tables {
				i, _ := table.Rows(target)
				j, _ := table.Rows(source)
				if i != j {
					t.Fatalf("failed to load data for table %v", table.Name)
				}
			}

			Session("backup.db", func(backup *Database) {
				t.Logf("Database opened: %v [flags: %v]", backup.Filename, int(backup.Flags))
				tables := []*Table{ FOO, BAR }
				for _, table := range tables {
					table.Drop(backup)
					table.Create(backup)
				}
				fatalOnError(t, target.Save(backup, "main"), "saving to %v[%]", backup.Filename, "main")
				for _, table := range tables {
					i, _ := table.Rows(target)
					j, _ := table.Rows(backup)
					if i != j {
						t.Fatalf("failed to load data for table %v", table.Name)
					}
				}
			})
		})
	})
}

/*
func TestBackup(t * testing.T) {
	Session("test.db", func(db *Database) {
	})
}
*/