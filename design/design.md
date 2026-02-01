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
- [x] crc-VariableError.md → `tracker.go`
- [x] crc-Priority.md → `tracker.go`
- [x] crc-Change.md → `tracker.go`
- [x] crc-Resolver.md → `tracker.go`
- [x] crc-ObjectRef.md → `tracker.go`
- [x] crc-ObjectRegistry.md → `tracker.go`

### Sequences
- [x] seq-create-variable.md → `tracker.go`
- [x] seq-destroy-variable.md → `tracker.go`
- [x] seq-detect-changes.md → `tracker.go`
- [x] seq-get-value.md → `tracker.go`
- [x] seq-set-value.md → `tracker.go`
- [x] seq-set-property.md → `tracker.go`
- [x] seq-to-value-json.md → `tracker.go`

### Test Designs
- [x] test-Tracker.md
- [x] test-Variable.md
- [x] test-Priority.md
- [x] test-Change.md
- [x] test-Resolver.md
- [x] test-ObjectRegistry.md
- [x] test-ValueJSON.md
- [x] test-Wrapper.md

## Gaps

**Status: GREEN** - All critical gaps resolved.

### Oversights (On)
- [ ] O1: Benchmark tests
  - [ ] CreateVariable
  - [ ] DetectChanges (tree traversal)
  - [ ] ToValueJSON
  - [ ] Call/CallWith reflection
- [x] O2: Typed errors (VariableError with VariableErrorType enum)
