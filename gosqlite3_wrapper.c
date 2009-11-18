#include "gosqlite3_wrapper.h"
#include <sqlite3.h>

int gosqlite3_bind_text(sqlite3_stmt* s, int p, const char* q, int n) {
    return sqlite3_bind_text(s, p, q, n, SQLITE_TRANSIENT);
}
