# Change Tracker

A Go package for variable management with automatic change detection. Track values in nested data structures and detect changes with priority-based sorting.

**No observer pattern, no event emitters, no interfaces to implement.** Just point the tracker at your existing data structures and detect what changed.

## Features

- **Variable Tracking** - Track values in nested object hierarchies with parent-child relationships
- **Change Detection** - Automatic detection via value comparison with `DetectChanges()`
- **Priority Sorting** - Changes returned sorted by priority (high → medium → low)
- **Object Registry** - Consistent identity for objects via weak references (Go 1.24+)
- **Path Navigation** - Navigate nested structures: `"Address.City"`, `"items.0"`
- **Pluggable Resolvers** - Custom navigation strategies for complex domains
- **Zero Coupling** - Domain objects require no modification

## Installation

```bash
go get github.com/zot/change-tracker
```

Requires Go 1.24+ (uses `weak` package).

## Quick Start

```go
tracker := change_tracker.NewTracker(nil)

// Create a root variable
root := tracker.NewVariable(nil, myData, "")

// Create child variables to track nested values
name := tracker.NewVariable(root, nil, "Name")
address := tracker.NewVariable(root, nil, "Address.City?priority=high")

// Make changes to myData...
myData.Name = "Updated"

// Detect what changed (sorted by priority)
changes := tracker.DetectChanges()
for _, change := range changes {
    fmt.Printf("Changed: %s\n", change.Variable.Path())
}
```

## Priority Levels

| Priority | Value | Use Case |
|----------|-------|----------|
| High | 1 | Critical changes to process first |
| Medium | 0 | Default priority |
| Low | -1 | Background/deferred changes |

Set via path properties: `"field?priority=high"`

## Access Modes

| Mode | Read | Write | Change Detection |
|------|------|-------|------------------|
| `rw` | ✓ | ✓ | ✓ |
| `r` | ✓ | ✗ | ✓ |
| `w` | ✗ | ✓ | ✗ |
| `action` | - | - | ✗ |

Set via path properties: `"field?access=r"`

## Concurrency

Change detection happens in a single thread. If your data structures are accessed from multiple goroutines, provide a custom Resolver implementation that implements safe access.

## Documentation

- [Specifications](specs/main.md) - Design principles and concepts
- [API Reference](specs/api.md) - Complete API documentation
- [Developer Guide](docs/developer-guide.md) - Architecture and integration
- [User Manual](docs/user-manual.md) - Usage examples and how-tos

## License

MIT
