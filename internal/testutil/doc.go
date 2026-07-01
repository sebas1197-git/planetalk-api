// Package testutil provides helpers for INTEGRATION tests: it spins up a real
// PostgreSQL in a throwaway container (via testcontainers-go) and applies all
// migrations. Almost everything here is behind the "integration" build tag, so
// it's excluded from normal builds and unit tests.
//
// Run integration tests with:  go test -tags=integration ./...
// (Docker must be available.)
package testutil
