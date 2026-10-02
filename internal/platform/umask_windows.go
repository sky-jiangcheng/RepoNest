//go:build windows

package platform

// SetPrivateUmask is a no-op on Windows: file ACLs there are governed by the
// profile directory's inherited permissions, not a creation mask.
func SetPrivateUmask() {}
