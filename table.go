package sqlite3

import "C"

import "fmt"
import "os"

type Table struct {
	Name		string
	ColumnSpec	string
}

func (t *Table) Create(db *Database) (e os.Error) {
	sql := fmt.Sprintf("CREATE TABLE %v (%v);", t.Name, t.ColumnSpec)
	_, e = db.Execute(sql, nil)
	return
}

func (t *Table) Drop(db *Database) (e os.Error) {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS %v;", t.Name, t.ColumnSpec)
	_, e = db.Execute(sql, nil)
	return
}

func (t *Table) Rows(db *Database) (c int, e os.Error) {
	sql := fmt.Sprintf("SELECT Count(*) FROM %v;", t.Name)
	_, e = db.Execute(sql, func(s *Statement, values ...interface{}) {
		c = int(values[0].(int64))
	})
	return
}