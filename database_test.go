package sqlite3

import "testing"

func (db *Database) createTestTables(t *testing.T, tables... *Table) {
	for _, table := range tables {
		table.Drop(db)
		table.Create(db)
		if c, _ := table.Rows(db); c != 0 {
			t.Fatalf("%v already contains data", table.Name)
		}
	}
}

func (db *Database) createTestData(t *testing.T, repeats int) {
	db.runQuery(t, "PRAGMA synchronous=OFF")
	for i := 0; i < repeats; i++ {
		if i % 2 == 0 {
			db.runQuery(t, "INSERT INTO foo values (?, 'holy moly')", i)
			db.runQuery(t, "INSERT INTO bar values (?, ?)", i, TwoItems{ "holy moly", "guacomole" })
		} else {
			db.runQuery(t, "INSERT INTO foo values (?, 'guacomole')", i)
			db.runQuery(t, "INSERT INTO bar values (?, ?)", i, TwoItems{ "guacomole", "holy moly" })
		}
	}
	db.runQuery(t, "PRAGMA synchronous=NORMAL")
}

func TestTransfers(t *testing.T) {
	TransientSession(func(source *Database) {
		source.createTestTables(t, FOO, BAR)
		source.createTestData(t, 1000)
		Session("target.db", func(target *Database) {
			t.Logf("Database opened: %v [flags: %v]", target.Filename, int(target.Flags))
			target.createTestTables(t, FOO, BAR)
			fatalOnError(t, target.Load(source, "main"), "loading from %v[%]", source.Filename, "main")
			for _, table := range []*Table{ FOO, BAR } {
				i, _ := table.Rows(target)
				j, _ := table.Rows(source)
				if i != j {
					t.Fatalf("failed to load data for table %v", table.Name)
				}
			}

			Session("backup.db", func(backup *Database) {
				t.Logf("Database opened: %v [flags: %v]", backup.Filename, int(backup.Flags))
				backup.createTestTables(t, FOO, BAR)
				fatalOnError(t, target.Save(backup, "main"), "saving to %v[%v]", backup.Filename, "main")
				for _, table := range []*Table{ FOO, BAR } {
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

func (r *Reporter) finished(t *testing.T) (finished bool) {
	report := <- (*r)
	if report != nil {
		switch e := report.Error.(type) {
		case Errno:		if e != DONE { t.Fatalf("Backup error %v", e) }
		case nil:		t.Logf("Backup still has %v pages of %v to copy to %v", report.Remaining, report.PageCount, report.Target)
		}
	}
	return closed(*r)
}

func TestBackup(t *testing.T) {
	var callbacks	int

	Session("test.db", func(db *Database) {
		db.createTestTables(t, FOO, BAR)
		db.createTestData(t, 1000)

		if sync_reporter, e := db.Backup(BackupParameters{Target: "sync.db", PagesPerStep: 3, QueueLength: 1}); e == nil {
			for callbacks = 0; !sync_reporter.finished(t); callbacks++ {}
			t.Logf("backup of %v generated %v callbacks", db.Filename, callbacks)
		}

		if async_reporter, e := db.Backup(BackupParameters{Target: "async.db", PagesPerStep: 3, QueueLength: 8}); e == nil {
			for callbacks = 0; !async_reporter.finished(t); callbacks++ {}
			t.Logf("backup of %v generated %v callbacks", db.Filename, callbacks)
		}
	})
}