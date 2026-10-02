// Package vecprobe is a build-capability regression test — it contains only a
// _test.go file and no production code, so importing modernc.org/sqlite/vec is
// confined to this test binary (its sqlite3_auto_extension side effect, which
// rules out RegisterPageCache, therefore never touches the app).
//
// It exists to keep ADR-0012's central claim honest: that sqlite-vec runs in
// this repo's pure-Go (CGO_ENABLED=0) modernc.org/sqlite stack. If a future
// driver upgrade drops the bundled vec extension, this test fails loudly
// before someone rediscovers it in production. Run it under CGO_ENABLED=0
// (as the CI cross-build does) to prove the zero-CGO property holds.
package vecprobe
