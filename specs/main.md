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
- `DetectChanges()` returns sorted changes and clears internal state automatically

## Core Concepts

### Tracker

The tracker is the central object that:
- Creates and manages variables
- Maintains a set of root variable IDs for efficient tree traversal
- Maintains an object registry (weak map from objects to variable IDs)
- Tracks which variables have changed
- Has a `Resolver` field (defaults to itself, using Go reflection)
- Serializes values to Value JSON form

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

The tracker maintains a **weak map** from Go objects (pointers and maps) to variable IDs:

- When a variable is created with a pointer or map value, the object is registered
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
- **Object references**: `{"obj": ID}` for registered objects (pointers, maps)

All pointers and maps must be registered before serialization. Unregistered pointers/maps cause an error.

This allows:
- Consistent identity for objects appearing in multiple places
- Cycle detection (registered objects break cycles)
- Compact serialization when the same object appears multiple times

### Value Resolvers

A **value resolver** knows how to navigate into values:
- `Get(obj, pathElement)` - Returns the value at the path element within obj
- `Set(obj, pathElement, value)` - Sets a value at the path element within obj

The tracker itself implements the resolver interface using Go reflection, supporting:
- Struct fields (by name)
- Map keys (string keys)
- Slice/array indices (integer keys)
- Method calls (zero-argument methods that return a value)

### Change Detection

Change detection uses tree traversal starting from root variables:

1. The tracker maintains a set of root variable IDs (variables with ParentID == 0)
2. `DetectChanges()` performs a depth-first traversal for each root variable:
   - For each active variable, convert current value to Value JSON and compare to stored Value JSON
   - If the variable is inactive, skip it and all its descendants
   - If active, recursively check all children
3. On variable creation, the initial value is converted to Value JSON and stored
4. After comparison, current Value JSON becomes the new stored Value JSON
5. Changes are sorted by priority and returned
6. Internal change records are cleared (but the returned slice remains valid)

Each variable stores its last known Value JSON for comparison purposes.

### Property Changes

Setting a variable's properties via `SetProperty()` also records changes in the tracker:
- Setting `priority` updates the variable's `ValuePriority`
- Setting `path` updates the variable's `Path` field (re-parses the path)
- Property changes are tracked separately from value changes

### Change Objects

A change record indicates what changed for a variable:
- **VariableID** - Which variable changed
- **ValueChanged** - Whether the value changed
- **PropertiesChanged** - Which properties changed (list of property names)

### Sorted Changes

`DetectChanges()` returns changes sorted by priority:
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

// Detect the change - returns sorted changes and clears internal state
changes := tracker.DetectChanges()
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

// DetectChanges returns sorted changes (high → medium → low) and clears internal state
for _, change := range tracker.DetectChanges() {
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
