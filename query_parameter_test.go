package sqlite3

import "testing"

func TestQueryParameterBinding(t *testing.T) {
	Session("test.db", func(db *Database) {
		if db == nil {
			t.Fatal("unable to acquire DB handle")
		}
		BAR.Create(db)
		sql := "INSERT INTO bar values (?, ?);"
		st, e := db.Prepare(sql, func(s *Statement) {
			fatalOnError(t, QueryParameter(1).Bind(s, nil), "unable to bind NULL to column 1")	
		})
		fatalOnError(t, e, "unable to prepare query: %v", sql)

		for _, v := range []interface{}{1.1, "hello", TwoItems{ "a", "b" }, []int{13, 27} } {
			fatalOnError(t, QueryParameter(1).Bind(st, v), "erroneously bound %v to column 1", v)
		}

		fatalOnError(t, QueryParameter(1).Bind(st, 1), "unable to bind integer to column 1")
		fatalOnError(t, QueryParameter(2).Bind(st, TwoItems{ "a", "b" }), "unable to bind blob to column 2")

		st, e = db.Prepare("INSERT INTO bar values (1, ?)")
		fatalOnError(t, e, "unable to prepare query: %v", st.SQLSource())
		fatalOnError(t, QueryParameter(1).Bind(st, TwoItems{ "a", "b" }), "unable to bind blob to column 2")
	})
}