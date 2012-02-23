package sqlite3

import (
	"testing"
	"time"
)

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
			t.Logf("Database opened: %v [flags: %v]", target.Filename, target.DBFlag)
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
				t.Logf("Database opened: %v [flags: %v]", backup.Filename, backup.DBFlag)
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

func (r *Reporter) finished(t *testing.T) bool {
	report, ok := <- (*r)
	if report != nil {
		switch e := report.Error; {
		case e == nil:
			if report.Verbose {
				t.Logf("Backup still has %v pages of %v to copy to %v", report.Remaining, report.Total, report.Target)
			}
		case e != DONE:
			t.Fatalf("Backup error %v", e)
		}
	}
	return !ok
}

func TestBackup(t *testing.T) {
	var messages	int

	Session("test.db", func(db *Database) {
		db.createTestTables(t, FOO, BAR)
		db.createTestData(t, 1000)

		TestSyncReporter := func(pages int, interval time.Duration) {
			if reporter, e := db.Backup(BackupParameters{Target: "sync.db", PagesPerStep: pages, QueueLength: 1, Interval: interval, Verbose: false}); e == nil {
				d := time.Now().UnixNano()
				for messages = 0; !reporter.finished(t); messages++ {}
				if interval == 0 {
					t.Logf("backup of %v generated %v synchronous messages and took %vns", db.Filename, messages, time.Now().UnixNano() - d)
				} else {
					t.Logf("backup of %v generated %v synchronous messages and took %vns with interval %v", db.Filename, messages, time.Now().UnixNano() - d, interval)
				}
			}
		}

		TestSyncReporter(3, 0)
		TestSyncReporter(3, 100000)

		TestSyncReporter(10, 0)
		TestSyncReporter(10, 100000)


		TestASyncReporter := func(queuelength, pages int, interval time.Duration) {
			if reporter, e := db.Backup(BackupParameters{Target: "sync.db", PagesPerStep: pages, QueueLength: 1, Interval: interval, Verbose: false}); e == nil {
				d := time.Now().UnixNano()
				for messages = 0; !reporter.finished(t); messages++ {}
				if interval == 0 {
					t.Logf("backup of %v generated %v synchronous messages and took %vns", db.Filename, messages, time.Now().UnixNano() - d)
				} else {
					t.Logf("backup of %v generated %v synchronous messages and took %vns with interval %v", db.Filename, messages, time.Now().UnixNano() - d, interval)
				}
			}
		}

		TestASyncReporter(1, 3, 0)
		TestASyncReporter(1, 10, 0)
		TestASyncReporter(1, 3, 100000)
		TestASyncReporter(1, 10, 100000)

		TestASyncReporter(8, 3, 0)
		TestASyncReporter(8, 10, 0)
		TestASyncReporter(8, 3, 100000)
		TestASyncReporter(8, 10, 100000)

		TestASyncReporter(12, 3, 0)
		TestASyncReporter(12, 10, 0)
		TestASyncReporter(12, 3, 100000)
		TestASyncReporter(12, 10, 100000)
	})
}

func TestExecute(t *testing.T) {
	t.Log("Test case for issue #11")
	db := TransientDatabase()
	if e := db.Open(O_FULLMUTEX, O_READWRITE, O_CREATE); e != nil {
		t.Logf("Open failed: %v", e)
	}
	defer db.Close()
	if _, e := db.Execute("CREATE TABLE t(id INTEGER PRIMARY KEY ASC, unique_int INTEGER UNIQUE);"); e != nil {
		t.Logf("Create table failed: %v", e)
	}
	if _, e := db.Execute("INSERT INTO t ( unique_int ) VALUES ( 1 );"); e != nil {
		t.Logf("1. Insert failed: %v", e)
	}
	if _, e := db.Execute("INSERT INTO t ( unique_int ) VALUES ( 2 );"); e != nil {
		t.Logf("2. Insert failed: %v", e)
	}
	if _, e := db.Execute("INSERT INTO t ( unique_int ) VALUES ( 2 );"); e == nil {
		t.Logf("3. Insert succeeded: %v", e)
	}
}