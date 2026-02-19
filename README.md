# Change Tracker

A Go package for variable management with automatic change detection. Track values in nested data structures and detect changes with priority-based sorting.

**No observer pattern, no event emitters, no interfaces to implement.** Just point the tracker at your existing data structures and detect what changed.

## Features

- **Variable Tracking** - Track values in nested object hierarchies with parent-child relationships
- **Change Detection** - Automatic detection via value comparison with `DetectChanges()`
- **Priority Sorting** - Changes returned sorted by priority (high → medium → low)
- **Object Registry** - Consistent identity for objects via weak references (Go 1.24+)
- **Path Navigation** - Navigate nested structures: `"Address.City"`, `"items.0"`, `"GetName()"`
- **Pluggable Resolvers** - Custom navigation strategies for complex domains
- **Wrappers** - Provide alternative objects for child navigation via custom resolvers
- **Recomputation Timing** - Per-variable `ComputeTime` and `MaxComputeTime` for profiling
- **Diagnostics** - Per-variable diagnostic collection via `Diag()` for debugging resolvers
- **Structured Errors** - Typed `VariableError` with categorized error types
- **Zero Coupling** - Domain objects require no modification

## Installation

```bash
go get github.com/zot/change-tracker
```

Requires Go 1.24+ (uses `weak` package).

## Quick Start

```go
tracker := changetracker.NewTracker()

// Create a root variable holding your data
data := &MyData{Name: "Alice", Count: 42}
root := tracker.CreateVariable(data, 0, "", nil)

// Create child variables to track nested values
name := tracker.CreateVariable(nil, root.ID, "Name", nil)
count := tracker.CreateVariable(nil, root.ID, "Count?priority=high", nil)

// Make changes to your data...
data.Name = "Bob"
data.Count = 100

// Detect what changed (sorted by priority: high → medium → low)
changes := tracker.DetectChanges()
for _, change := range changes {
    fmt.Printf("Variable %d: value=%v props=%v\n",
        change.VariableID, change.ValueChanged, change.PropertiesChanged)
}
```

## Access Modes

| Mode | Read | Write | Change Detection | Initial Value |
|------|------|-------|------------------|---------------|
| `rw` (default) | ✓ | ✓ | ✓ | computed |
| `r` | ✓ | ✗ | ✓ | computed |
| `w` | ✗ | ✓ | ✗ | computed |
| `action` | ✗ | ✓ | ✗ | skipped |

Set via path query: `"field?access=r"` or properties map.

The `action` mode is for variables that trigger side effects (like `AddContact(_)`) where computing the initial value would invoke the action prematurely.

## Path Navigation

Paths navigate from a parent variable's value to a nested value:

```go
// Struct fields
tracker.CreateVariable(nil, root.ID, "Address.City", nil)

// Slice indices
tracker.CreateVariable(nil, root.ID, "Items.0.Name", nil)

// Zero-arg method calls (getters)
tracker.CreateVariable(nil, root.ID, "GetName()", map[string]string{"access": "r"})

// One-arg method calls (setters)
tracker.CreateVariable(nil, root.ID, "SetName(_)", map[string]string{"access": "action"})

// URL-style query parameters set properties
tracker.CreateVariable(nil, root.ID, "Count?priority=high&label=counter", nil)
```

## Priority Levels

| Priority | Value | Use Case |
|----------|-------|----------|
| High | 1 | Critical changes to process first |
| Medium | 0 | Default priority |
| Low | -1 | Background/deferred changes |

Set via path properties: `"field?priority=high"`

Properties can also have independent priorities via suffix: `v.SetProperty("label:high", "Important")`

## Active Flag

Variables can be deactivated to skip change detection for an entire subtree:

```go
v.SetActive(false)  // v and all descendants skipped during DetectChanges
v.SetActive(true)   // re-enable
```

## Wrappers

Custom resolvers can provide alternative objects for child navigation:

```go
type myResolver struct { *changetracker.Tracker }

func (r *myResolver) CreateWrapper(v *changetracker.Variable) any {
    if p, ok := v.Value.(*Person); ok {
        return &PersonView{DisplayName: p.First + " " + p.Last}
    }
    return nil
}

tr := changetracker.NewTracker()
tr.Resolver = &myResolver{tr}
parent := tr.CreateVariable(person, 0, "?wrapper=true", nil)
// Children now navigate through PersonView, not Person
child := tr.CreateVariable(nil, parent.ID, "DisplayName", nil)
```

Wrappers support state preservation — returning the same pointer from `CreateWrapper` keeps the wrapper's internal state intact across value changes.

## Recomputation Timing

Each variable tracks how long its own path navigation takes:

```go
child := tracker.CreateVariable(nil, root.ID, "Address.City", nil)
val, _ := child.Get()

fmt.Println(child.ComputeTime)    // duration of most recent recompute
fmt.Println(child.MaxComputeTime) // peak duration across all recomputes
```

Timing measures only the variable's own path navigation, excluding parent value retrieval.

## Diagnostics

Custom resolvers can emit per-variable diagnostics during path navigation:

```go
tracker.DiagLevel = 1  // enable level-1 diagnostics

// In your custom resolver's Get method:
func (r *myResolver) Get(obj any, pathElement any) (any, error) {
    r.Tracker.Diag(1, "resolving %v on %T", pathElement, obj)
    return r.Tracker.Get(obj, pathElement)
}

// After Get() or DetectChanges(), check variable.Diags
val, _ := child.Get()
for _, msg := range child.Diags {
    fmt.Println(msg)
}
```

Diagnostics are cleared at the start of each recompute. `DiagLevel = 0` (default) disables collection.

## Structured Errors

Operations return `*VariableError` with typed error categories:

```go
_, err := variable.Get()
if ve, ok := variable.Error.(*changetracker.VariableError); ok {
    switch ve.ErrorType {
    case changetracker.PathError:   // path navigation failed
    case changetracker.NotFound:    // variable or parent not found
    case changetracker.BadAccess:   // access mode violation
    case changetracker.NilPath:     // nil value in path
    case changetracker.BadCall:     // method call failed
    }
}
```

## Object Registry

The tracker maintains a weak map from Go objects (pointers and maps) to IDs, providing consistent identity without modifying domain types:

```go
alice := &Person{Name: "Alice"}
root := tracker.CreateVariable(alice, 0, "", nil)

// Same object always serializes to the same reference
json := tracker.ToValueJSON(alice)  // {"obj": 1}

// Weak references — objects can be garbage collected normally
id, ok := tracker.LookupObject(alice)
obj := tracker.GetObject(id)
```

## Concurrency

Change detection happens in a single thread. If your data structures are accessed from multiple goroutines, provide a custom Resolver implementation that implements safe access.

## Documentation

- [Specifications](specs/main.md) - Design principles and concepts
- [API Reference](specs/api.md) - Complete API documentation
- [Resolver Spec](specs/resolver.md) - Value resolver interface
- [Wrapper Spec](specs/wrapper.md) - Wrapper support
- [Value JSON Spec](specs/value-json.md) - Value JSON serialization format

## License

MIT
