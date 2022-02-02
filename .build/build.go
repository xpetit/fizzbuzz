// This package exists only to speed up the building of Docker images by precompiling the SQLite amalgamation.
package build

import _ "github.com/mattn/go-sqlite3"
