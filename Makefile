include $(GOROOT)/src/Make.inc

TARG=sqlite3

CGOFILES=\
	sqlite3.go\
	database.go\
	statement.go\
	query_parameter.go\
	result_column.go\
	table.go\
	backup.go

ifeq ($(GOOS),darwin)
CGO_LDFLAGS=/usr/lib/libsqlite3.0.dylib
else
CGO_LDFLAGS=-lsqlite3
endif

include $(GOROOT)/src/Make.pkg
