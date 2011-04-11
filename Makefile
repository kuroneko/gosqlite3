include $(GOROOT)/src/Make.inc

TARG=github.com/kuroneko/sqlite3

CGOFILES=\
	sqlite3.go

CGO_LDFLAGS=-lsqlite3

include $(GOROOT)/src/Make.pkg
