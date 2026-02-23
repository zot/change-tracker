# Change Tracker Specification

A Go package (`github.com/zot/change-tracker`) that provides variable management with automatic change detection.

## Overview

The package provides:
- A **change tracker** that manages variables and detects changes
- **Variables** that hold values and track parent-child relationships
- **Object registry** with weak references for consistent object identity
- **Value JSON** serialization with object references
- **Change detection** via value comparison
- **Pluggable value resolution** for navigating into objects

## Design Principles

### Simplicity First
- Pure data structure library
- No thread safety - callers are responsible for synchronization if needed
- Minimal dependencies

### Explicit Over Automatic
- Changes are detected only when `DetectChanges()` is called
- `DetectChanges()` returns `bool`; `GetChanges()` returns sorted changes and clears internal state

## Core Concepts

### Tracker

The tracker is the central object that:
- Creates and manages variables
- Maintains a set of root variable IDs for efficient tree traversal
- Maintains an object registry (weak map from objects to variable IDs)
- Tracks which variables have changed
- Has a `Resolver` field (defaults to itself, using Go reflection)
- Serializes values to Value JSON form
- Has a `DiagLevel` field (int, default 0) controlling diagnostic output
- Tracks which variable is currently being computed (`computingVar`)

### Variables

Each variable has:
- **ID** - Unique integer identifier (assigned by the tracker)
- **ParentID** - ID of parent variable (0 = no parent, making it a "root" variable)
- **ChildIDs** - Slice of child variable IDs (maintained automatically by the tracker)
- **Active** - Whether this variable and its children should be checked for changes (default: true)
- **Access** - Access mode: `"rw"` (read-write, default), `"r"` (read-only), `"w"` (write-only), or `"action"` (action trigger)
- **Properties** - Map of string key-value metadata
- **PropertyPriorities** - Map of property names to their priorities
- **Path** - Parsed path elements (e.g., `"Address.City"` becomes `["Address", "City"]`)
- **Value** - Cached value for child navigation
- **ValueJSON** - Cached Value JSON for change detection
- **ValuePriority** - Priority level for the variable's value
- **WrapperValue** - Optional wrapper object for child navigation (created via `Resolver.CreateWrapper` when "wrapper" property is set)
- **WrapperJSON** - Serialized form of WrapperValue (via `ToValueJSON`)
- **ComputeTime** - Duration of the most recent value recomputation (excludes parent value retrieval, measures only this variable's own path navigation)
- **MaxComputeTime** - Maximum ComputeTime observed across all recomputations
- **Diags** - Slice of diagnostic strings collected during the most recent value recomputation (cleared at the start of each recompute)

A variable's value is computed by:
1. Starting at the parent's cached value
2. Applying each path element using the tracker's resolver

Variables form a tree structure via parent-child relationships.

### Active Variables

The **Active** field controls whether a variable participates in change detection:

- When `Active` is true (default), the variable and its children are checked for changes
- When `Active` is false, the variable and all its descendants are skipped during change detection
- Setting a variable to inactive effectively "prunes" that entire subtree from change detection
- The Active field can be toggled at any time; changes take effect on the next `DetectChanges()` call

### Access Modes

The **Access** field controls read/write permissions and initialization behavior:

| Mode     | Get | Set | Change Detection | Initial Value Computed |
|----------|-----|-----|------------------|------------------------|
| `rw`     | ✓   | ✓   | ✓                | ✓                      |
| `r`      | ✓   | ✗   | ✓                | ✓                      |
| `w`      | ✗   | ✓   | ✗                | ✓                      |
| `action` | ✗   | ✓   | ✗                | ✗                      |

- **rw** (default): Full read-write access, included in change detection. Paths may end in `()` but not `(_)`.
- **r**: Read-only, included in change detection but `Set()` fails. Paths may end in `()`.
- **w**: Write-only, `Get()` fails and excluded from change detection. Paths may end in `(_)`.
- **action**: Like write-only, but initial value is NOT computed during `CreateVariable`. Paths may end in `()` or `(_)`.

The `action` mode is designed for variables that trigger side effects (like calling `AddContact(_)`) where navigating the path during creation would invoke the action prematurely.

**Path restrictions:** Paths ending in `(_)` require `access: "w"` or `access: "action"`. Paths ending in `()` are allowed with `rw`, `r`, or `action` access.

### Priorities

Values and properties can have priority levels: **Low**, **Medium** (default), and **High**.

- A value's priority is set via the `priority` property (values: `"low"`, `"medium"`, `"high"`)
- Each property's priority is appended to its name with a colon suffix: `:low`, `:medium`, `:high`
- When setting a property with `SetProperty("name:high", "value")`, the priority is extracted and stored separately
- Properties without a priority suffix default to Medium

### Object Registry

The tracker maintains a **weak map** from Go objects (pointers, maps, and funcs) to IDs:

- Objects are registered automatically via `ToValueJSON()` or explicitly via `RegisterObject()`
- Objects have identity independent of where they appear
- The same object in multiple locations serializes to the same `{"obj": id}`
- Uses Go 1.24+ `weak` package for weak references
- Objects can be garbage collected when no longer referenced by application code
- The registry is automatically cleaned up as objects are collected
- **Frictionless**: domain objects require no modification - no interfaces, no embedded IDs

### Value JSON

**Value JSON** is a serialization format with three value types:

- **Primitives**: strings, numbers, booleans, null
- **Arrays**: with elements in Value JSON form
- **Object references**: `{"obj": ID}` for registered objects (pointers, maps, funcs)
- **Structs**: serialize as `nil` (not directly serializable)

Unregistered pointers, maps, and funcs are auto-registered during `ToValueJSON()`.

This allows:
- Consistent identity for objects appearing in multiple places
- Cycle detection (registered objects break cycles)
- Compact serialization when the same object appears multiple times

### Value Resolvers

A **value resolver** knows how to navigate into values:
- `Get(obj, pathElement)` - Returns the value at the path element within obj
- `Set(obj, pathElement, value)` - Sets a value at the path element within obj
- `Call(obj, methodName)` - Invokes a zero-arg method, returns result
- `CallWith(obj, methodName, value)` - Invokes a one-arg method
- `CreateValue(variable, typ, value)` - Creates a value for a variable with a "create" property
- `CreateWrapper(variable)` - Creates a wrapper object for child navigation
- `GetType(variable, value)` - Returns a type string for the value
- `ConvertToValueJSON(tracker, value)` - Converts domain-specific types for serialization

The tracker itself implements the resolver interface using Go reflection, supporting:
- Struct fields (by name)
- Map keys (string keys)
- Slice/array indices (integer keys)
- Method calls (zero-argument methods that return a value, one-argument void methods)

### Change Detection

Change detection uses priority-ordered graph traversal:

1. The tracker maintains a set of root variable IDs (variables with ParentID == 0)
2. `DetectChanges()` collects active variables via tree traversal (respecting Active flag propagation), then checks them in priority order:
   - **Collection phase**: Walk the tree from roots. Inactive variables prune their entire subtree. Active readable variables are grouped by priority (high, medium, low). Non-readable variables are skipped but their children are still collected.
   - **Check phase**: Check all high-priority variables first, then all medium, then all low. Before checking any variable, ensure its readable ancestors have been checked first (parent-before-child guarantee). A `checked` set prevents double-processing. For each variable, convert current value to Value JSON and compare to stored Value JSON. If different, perform wrapper-aware comparison (see below).
3. On variable creation, the initial value is converted to Value JSON and stored
4. After comparison, current Value JSON becomes the new stored Value JSON
5. `GetChanges()` returns changes sorted by priority and clears internal change records (but the returned slice remains valid)

**Wrapper-aware comparison**: When a variable's ValueJSON changes, the check phase saves the old wrapper-aware JSON (`JsonForUpdate()`), updates Value/ValueJSON, calls `updateWrapper()`, then compares the old wrapper-aware JSON to the new. A change is only reported if the wrapper-aware JSON differs. This means a value change that doesn't affect the external (wrapper) representation is silently absorbed — the cache is updated but no change is reported to consumers.

Each variable stores its last known Value JSON for comparison purposes.

**Priority controls computation order**: High-priority variables are checked (value fetched and compared) before medium-priority variables, which are checked before low-priority variables. This ensures that high-priority changes are detected and their cached values updated before lower-priority variables are processed.

**Parent-before-child guarantee**: Since child variables derive their values from parent `NavigationValue()`, readable ancestors are always checked before their descendants. When a high-priority child has a lower-priority parent, the parent is pulled forward and checked first. The `checked` set ensures each variable is checked at most once.

### Change Counters

The tracker and variables maintain change counters for monitoring:

- **Tracker.ChangeCount** - An `int64` counter incremented each time `DetectChanges()` finds any changes (i.e., when it returns `true`). Not incremented when no changes are detected.
- **Variable.ChangeCount** - An `int64` counter on each variable, incremented each time that variable's value actually changes during `DetectChanges()`. Property-only changes do not increment this counter.

### Diagnostics

The tracker provides per-variable diagnostic support for debugging resolver and path navigation behavior:

- **DiagLevel** - An integer on the tracker (default 0) that controls which diagnostics are collected. A level of 0 means diagnostics are disabled.
- **Diag(level, format, args...)** - A method on the tracker that adds a formatted diagnostic string to the currently-computing variable's `Diags` slice, but only if the tracker's `DiagLevel` is >= the specified level.
- **computingVar** - The tracker internally tracks which variable is currently having its value computed. This is set before path navigation in `GetValue()` and cleared after.
- **Diags** - Each variable has a `Diags []string` field. It is cleared at the start of each value recomputation (in `GetValue()`). Diagnostics accumulate during path navigation and resolver calls.
- If `Diag()` is called when no variable is being computed, the call is ignored.

### Property Changes

Setting a variable's properties via `SetProperty()` also records changes in the tracker:
- Setting `priority` updates the variable's `ValuePriority`
- Setting `path` updates the variable's `Path` field (re-parses the path)
- Property changes are tracked separately from value changes

### Wrapper Support

A variable can have an optional wrapper that stands in for its value when child variables navigate paths. This allows the resolver to provide a different interface to children than the underlying value.

- When the "wrapper" property is set and ValueJSON is non-nil, `Resolver.CreateWrapper(variable)` is called
- Child variables use `NavigationValue()` which returns WrapperValue if present, otherwise Value
- Change detection uses `JsonForUpdate()` which returns WrapperJSON if present, otherwise ValueJSON
- Wrappers can be preserved across value changes by returning the same pointer from `CreateWrapper`

### Change Objects

A change record indicates what changed for a variable:
- **VariableID** - Which variable changed
- **ValueChanged** - Whether the value changed
- **PropertiesChanged** - Which properties changed (list of property names)

### Sorted Changes

`GetChanges()` returns changes sorted by priority:
- Returns a slice of change objects sorted by priority (high → medium → low)
- Changes with mixed priorities are split: a variable may appear multiple times at different priority levels
- The value's priority determines where value changes appear
- Each property's priority determines where that property change appears
- The tracker reuses an internal `sortedChanges` slice to avoid allocations

## Documentation

- [api.md](api.md) - Detailed API documentation
- [resolver.md](resolver.md) - Value resolver specification
- [value-json.md](value-json.md) - Value JSON serialization format

## Use Cases

### Simple Value Tracking
```go
type MyData struct {
    Count int
}

tracker := changetracker.NewTracker()
data := &MyData{Count: 42}
root := tracker.CreateVariable(data, 0, "", nil)

// Create a variable for the Count field
countVar := tracker.CreateVariable(nil, root.ID, "Count", nil)

// Modify value externally
data.Count = 100

// Detect the change
tracker.DetectChanges()  // returns true
changes := tracker.GetChanges()
// changes contains countVar.ID with ValueChanged: true
```

### Object Registration
```go
type Person struct {
    Name string
    Age  int
}

tracker := changetracker.NewTracker()
alice := &Person{Name: "Alice", Age: 30}
bob := &Person{Name: "Bob", Age: 25}
tracker.CreateVariable(alice, 0, "", nil)  // ID 1
tracker.CreateVariable(bob, 0, "", nil)    // ID 2

// Serialize to Value JSON - registered objects become {"obj": id}
json := tracker.ToValueJSON(alice)
// json = {"obj": 1}

// Arrays of registered objects serialize as arrays of references
people := []*Person{alice, bob, alice}
json2 := tracker.ToValueJSON(people)
// json2 = [{"obj": 1}, {"obj": 2}, {"obj": 1}]
```

### Hierarchical Data
```go
tracker := changetracker.NewTracker()
root := tracker.CreateVariable(rootObj, 0, "", nil)
child := tracker.CreateVariable(nil, root.ID, "items.1", nil)
```

### Path-Based Navigation
```go
type Person struct {
    Name    string
    Address Address
}
type Address struct {
    City string
}

tracker := changetracker.NewTracker()
person := &Person{Name: "Alice", Address: Address{City: "Boston"}}

// Root variable holds the person object
root := tracker.CreateVariable(person, 0, "", nil)

// Child variable navigates to address.city via path
cityVar := tracker.CreateVariable(nil, root.ID, "Address.City", nil)

// Get the value
city, _ := cityVar.Get()  // returns "Boston"

// Set the value
cityVar.Set("New York")
```

### Path with Properties (URL-style Query Syntax)
```go
tracker := changetracker.NewTracker()
root := tracker.CreateVariable(data, 0, "", nil)

// Path can include properties using URL-style query syntax
// Properties in path override those in the properties map
child := tracker.CreateVariable(nil, root.ID, "items.0?label=First Item&priority=high", nil)
// child.Properties["label"] == "First Item"
// child.ValuePriority == High
```

### Property Priorities
```go
tracker := changetracker.NewTracker()
v := tracker.CreateVariable(data, 0, "", nil)

// Set property with priority suffix
v.SetProperty("label:high", "Important")
// v.Properties["label"] == "Important"
// v.PropertyPriorities["label"] == High

// Set property without suffix (defaults to Medium)
v.SetProperty("name", "Counter")
// v.PropertyPriorities["name"] == Medium
```

### Sorted Changes by Priority
```go
tracker := changetracker.NewTracker()
data := &MyData{Count: 0, Label: ""}
root := tracker.CreateVariable(data, 0, "", map[string]string{"priority": "high"})

// Create child variables with different priorities
countVar := tracker.CreateVariable(nil, root.ID, "Count?priority=high", nil)
labelVar := tracker.CreateVariable(nil, root.ID, "Label?priority=low", nil)

// Make changes
data.Count = 42
labelVar.SetProperty("hint:medium", "A helpful hint")

// DetectChanges returns bool; GetChanges returns sorted changes (high → medium → low)
tracker.DetectChanges()
for _, change := range tracker.GetChanges() {
    // High priority changes come first, then medium, then low
    fmt.Printf("Variable %d (priority %d): value=%v props=%v\n",
        change.VariableID, change.Priority,
        change.ValueChanged, change.PropertiesChanged)
}
```

### Custom Resolver
```go
tracker := changetracker.NewTracker()
tracker.Resolver = &MyCustomResolver{}

// All variables use the custom resolver for path navigation
root := tracker.CreateVariable(myData, 0, "", nil)
child := tracker.CreateVariable(nil, root.ID, "key", nil)
val, _ := child.Get()
```
