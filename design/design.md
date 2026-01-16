# Design: Change Tracker

**Source Spec:** specs/main.md

## Intent

A Go package providing variable management with automatic change detection. Supports hierarchical variables with path-based navigation, pluggable resolvers, weak object references, and priority-based change sorting.

## Cross-cutting Concerns

- **Resolver Interface**: Used by Tracker, Variable for path navigation (crc-Resolver.md)
- **Priority Type**: Used by Variable for value and property priorities (crc-Priority.md)

## Artifacts

### CRC Cards
- [x] crc-Tracker.md → `tracker.go`
- [x] crc-Variable.md → `tracker.go`
- [x] crc-Priority.md → `tracker.go`
- [x] crc-Change.md → `tracker.go`
- [x] crc-Resolver.md → `tracker.go`
- [x] crc-ObjectRef.md → `tracker.go`
- [x] crc-ObjectRegistry.md → `tracker.go`

### Sequences
- [x] seq-create-variable.md → `tracker.go:CreateVariable`
- [x] seq-destroy-variable.md → `tracker.go:DestroyVariable`
- [x] seq-detect-changes.md → `tracker.go:DetectChanges`
- [x] seq-get-value.md → `tracker.go:Variable.Get`
- [x] seq-set-value.md → `tracker.go:Variable.Set`
- [x] seq-set-property.md → `tracker.go:Variable.SetProperty`
- [x] seq-to-value-json.md → `tracker.go:ToValueJSON`

### Test Designs
- [x] test-Tracker.md → `tracker_test.go`
- [x] test-Variable.md → `tracker_test.go`
- [x] test-Priority.md → `tracker_test.go`
- [x] test-Change.md → `tracker_test.go`
- [x] test-Resolver.md → `tracker_test.go`
- [x] test-ObjectRegistry.md → `tracker_test.go`
- [x] test-ValueJSON.md → `tracker_test.go`
- [x] test-Wrapper.md → `tracker_test.go`

## Gaps

**Status: GREEN** - All critical gaps resolved.

### Oversights (On)
- [ ] O1: Benchmark tests
  - [ ] CreateVariable
  - [ ] DetectChanges (tree traversal)
  - [ ] ToValueJSON
  - [ ] Call/CallWith reflection
- [ ] O2: Typed errors (ErrWriteOnly, ErrReadOnly, ErrPathNotFound)
