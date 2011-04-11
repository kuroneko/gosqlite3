package sqlite3

import "testing"

func TestResultColumn(t *testing.T) {
	Session("test.db", func(db *Database) {
		BAR.Create(db)
		st, e := db.Prepare("INSERT INTO bar values (?, ?)")
		fatalOnError(t, e, "unable to prepare query: %v", st.SQLSource())

		fatalOnError(t, QueryParameter(1).Bind(st, nil), "unable to bind NULL to column 1")
		for _, v := range []interface{}{1.1, "hello", TwoItems{ "a", "b" }, []int{13, 27} } {
			fatalOnError(t, QueryParameter(1).Bind(st, v), "erroneously bound %v to column 1", v)
		}
		fatalOnError(t, QueryParameter(1).Bind(st, 1), "unable to bind integer to column 1")
		fatalOnError(t, QueryParameter(2).Bind(st, TwoItems{ "a", "b" }), "unable to bind blob to column 2")
	})
}