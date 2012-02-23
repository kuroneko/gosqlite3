package sqlite3

// #include <sqlite3.h>
import "C"

type ProgressReport struct {
	Error			error
	Total			int
	Remaining		int
	Source			string
	Target			string
	Verbose			bool
}