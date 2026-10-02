//go:build !windows

package platform

import "syscall"

// SetPrivateUmask narrows the process file-creation mask to owner-only, so
// files the app creates after this point (SQLite WAL sidecars on later
// sessions, exports, logs) are not readable by other local users. The data
// directory itself is created 0750 and db.InitDB additionally pins the
// database files to 0600 — this covers everything created in between.
func SetPrivateUmask() {
	_ = syscall.Umask(0o077)
}
