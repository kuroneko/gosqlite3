include $(GOROOT)/src/Make.$(GOARCH)

TARG=sqlite

CGOFILES=\
	sqlite.go

CGO_LDFLAGS=-lsqlite3

include $(GOROOT)/src/Make.pkg

sqltest: install sqltest.go
	$(GC) sqltest.go
	$(LD) -o $@ sqltest.$O
