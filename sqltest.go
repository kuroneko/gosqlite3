package main

import "fmt"
import "sqlite"

func main() {
	sqlite.Initialize();
	fmt.Printf("Sqlite3 Version: %s\n", sqlite.LibVersion());

	dbh := new(sqlite.Handle);
	dbh.Open("test.db");

	dbh.Close();
	sqlite.Shutdown();
}