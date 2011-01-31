package sqlite3

import "testing"

func TestSimpleTransaction(t *testing.T) {
	TransientSession(func(db *Database) {
		db.createTestTables(t, FOO, BAR)
		doNothing := func(d TransactionalDatabase) { }
		raiseOK := func(d TransactionalDatabase) { panic(OK) }
		raiseErrno := func(d TransactionalDatabase) { panic(MISUSE) }

		fatalOnError(t, Transaction{}.Execute(db), "empty transaction")	
		fatalOnError(t, Transaction{ doNothing }.Execute(db), "inconsequential transaction")
		fatalOnError(t, Transaction{ raiseOK }.Execute(db), "transaction raises OK")
		fatalOnSuccess(t, Transaction{ raiseErrno }.Execute(db), "transaction raises Errno")
	})
}