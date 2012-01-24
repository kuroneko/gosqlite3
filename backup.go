package sqlite3

// #include <sqlite3.h>
import "C"

type Backup struct {
	cptr	*C.sqlite3_backup
	db		*Database
}

func NewBackup(d *Database, ddb string, s *Database, sdb string) (b *Backup, e error) {
	if cptr := C.sqlite3_backup_init(d.handle, C.CString(ddb), s.handle, C.CString(sdb)); cptr != nil {
	 	b = &Backup{ cptr: cptr, db: d }
	} else {
		if e = d.Error(); e == OK {
			e = nil
		}
	}
	return
}

func (b *Backup) Step(pages int) error {
	if e := Errno(C.sqlite3_backup_step(b.cptr, C.int(pages))); e != OK {
		return e
	}
	return nil
}

func (b *Backup) Remaining() int {
	return int(C.sqlite3_backup_remaining(b.cptr))
}

func (b *Backup) PageCount() int {
	return int(C.sqlite3_backup_pagecount(b.cptr))
}

func (b *Backup) Finish() error {
	if e := Errno(C.sqlite3_backup_finish(b.cptr)); e != OK {
		return e
	}
	return nil
}

func (b *Backup) Full() error {
	b.Step(-1)
	b.Finish()
	if e := b.db.Error(); e != OK {
		return e
	}
	return nil
}