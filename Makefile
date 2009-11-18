include $(GOROOT)/src/Make.$(GOARCH)

TARG=sqlite3

CGOFILES=\
	sqlite3.go

CGO_LDFLAGS=gosqlite3_wrapper.o -lsqlite3

%: gosqlite3_wrapper.o install %.go
	$(GC) $*.go
	$(LD) -o $* $*.$O

gosqlite3_wrapper.o: gosqlite3_wrapper.c
	gcc -fPIC -O2 -o gosqlite3_wrapper.o -c gosqlite3_wrapper.c

include $(GOROOT)/src/Make.pkg

sqltest: install sqltest.go
	$(GC) sqltest.go
	$(LD) -o $@ sqltest.$O
