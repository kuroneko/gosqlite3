#ifndef GOSQLITE3_WRAPPER_H_
#define GOSQLITE3_WRAPPER_H_

#include <sqlite3.h>

int gosqlite3_bind_text(sqlite3_stmt* s, int p, const char* q, int n);

#endif  // GOSQLITE3_WRAPPER_H_

