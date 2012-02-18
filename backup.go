package sqlite3

// #include <sqlite3.h>
// #include <stdlib.h>
import "C"

import (
	"unsafe"
)

// Backup implements the SQLite Online Backup API.
//
// The backup API copies the content of one database to another. It is 
// useful either for creating backups of databases or for copying in-memory 
// databases to or from persistent files.
type Backup struct {
	cptr *C.sqlite3_backup
	db   *Database
}

// NewBackup initializes and returns the handle to a backup.
func NewBackup(d *Database, ddb string, s *Database, sdb string) (b *Backup, e error) {
	dname := C.CString(ddb)
	defer C.free(unsafe.Pointer(dname))
	sname := C.CString(sdb)
	defer C.free(unsafe.Pointer(sname))

	if cptr := C.sqlite3_backup_init(d.handle, dname, s.handle, sname); cptr != nil {
		b = &Backup{cptr: cptr, db: d}
	} else {
		e = d.Error()
	}
	return
}

// Step will copy up to `pages` between the source and destination database.
// If `pages` is negative, all remaining source pages are copied.
func (b *Backup) Step(pages int) error {
	return SQLiteError(C.sqlite3_backup_step(b.cptr, C.int(pages)))
}

// Remaining returns the number of pages still to be backed up.
func (b *Backup) Remaining() int {
	return int(C.sqlite3_backup_remaining(b.cptr))
}

// PageCount returns the total number of pages in the source database.
func (b *Backup) PageCount() int {
	return int(C.sqlite3_backup_pagecount(b.cptr))
}

// Finish should be called when the backup is done, an error occured or when 
// the application wants to abandon the backup operation.
func (b *Backup) Finish() error {
	return SQLiteError(C.sqlite3_backup_finish(b.cptr))
}

// Full creates a full backup of the database.
func (b *Backup) Full() error {
	b.Step(-1)
	b.Finish()
	return b.db.Error()
}