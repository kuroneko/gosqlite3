package sqlite3

// #include <sqlite3.h>
import "C"
import "os"

type TransactionalDatabase interface {
	Begin() os.Error
	Rollback() os.Error
	Commit() os.Error
}

type Transaction []func(db TransactionalDatabase)

func (t Transaction) Execute(db TransactionalDatabase) (e os.Error) {
	defer func() {
		switch r := recover().(type) {
		case nil:
			e = db.Commit()
		case Errno:
			if r == OK {
				e = db.Commit()
			} else {
				if db.Rollback() != nil {
					panic(e)
				} else {
					e = r
				}
			}
		case os.Error:
			if db.Rollback() != nil {
				panic(e)
			} else {
				e = r
			}
		default:
			panic(r)
		}
	}()

	if e = db.Begin(); e == nil {
		for _, f := range t {
			f(db)
		}
	}
	return
}