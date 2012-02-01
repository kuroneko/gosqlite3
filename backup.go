package sqlite3

// #include <sqlite3.h>
import "C"

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
	if cptr := C.sqlite3_backup_init(d.handle, C.CString(ddb), s.handle, C.CString(sdb)); cptr != nil {
		b = &Backup{cptr: cptr, db: d}
	} else {
		if e = d.Error(); e == OK {
			e = nil
		}
	}
	return
}

// Step will copy up to `pages` between the source and destination database.
// If `pages` is negative, all remaining source pages are copied.
func (b *Backup) Step(pages int) error {
	if e := Errno(C.sqlite3_backup_step(b.cptr, C.int(pages))); e != OK {
		return e
	}
	return nil
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
	if e := Errno(C.sqlite3_backup_finish(b.cptr)); e != OK {
		return e
	}
	return nil
}

// Full creates a full backup of the database.
func (b *Backup) Full() error {
	b.Step(-1)
	b.Finish()
	if e := b.db.Error(); e != OK {
		return e
	}
	return nil
}
