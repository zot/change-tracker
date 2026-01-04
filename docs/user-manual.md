# User Manual

<!-- Source: specs/main.md, specs/api.md -->

## Table of Contents

1. [Introduction](#introduction)
2. [Getting Started](#getting-started)
3. [Features](#features)
4. [How-To Guides](#how-to-guides)
5. [Troubleshooting](#troubleshooting)

---

## Introduction

### What is Change Tracker?

Change Tracker is a Go package that helps you track values in your application and detect when they change. It provides:

- **Variable tracking**: Monitor specific values within your data structures
- **Change detection**: Know which values have changed
- **Priority sorting**: Process important changes first
- **Object identity**: Track the same object appearing in multiple places

### When to Use Change Tracker

Use Change Tracker when you need to:

- Update a UI when data changes
- Synchronize data between systems
- Log or audit value changes
- Implement undo/redo functionality
- Trigger actions when specific values change

### Key Concepts

| Concept        | Description                                                     |
|----------------|-----------------------------------------------------------------|
| **Tracker**    | The central manager that tracks variables and detects changes   |
| **Variable**   | A tracked value with an ID, parent relationship, and properties |
| **Path**       | A dot-separated string describing how to navigate to a value    |
| **Priority**   | Level (low/medium/high) that determines change processing order |
| **Value JSON** | Internal format for comparing values                            |

---

## Getting Started

### Installation

Add the package to your Go project:

```bash
go get github.com/zot/change-tracker
```

**Requirement**: Go 1.24 or higher

### Basic Example

```go
package main

import (
    "fmt"
    ct "github.com/zot/change-tracker"
)

type Counter struct {
    Value int
}

func main() {
    // 1. Create a tracker
    tracker := ct.NewTracker()

    // 2. Create your data
    counter := &Counter{Value: 0}

    // 3. Track the data
    root := tracker.CreateVariable(counter, 0, "", nil)
    valueVar := tracker.CreateVariable(nil, root.ID, "Value", nil)

    // 4. Modify the data
    counter.Value = 42

    // 5. Detect changes (returns sorted changes and clears internal state)
    changes := tracker.DetectChanges()

    // 6. Check what changed
    for _, change := range changes {
        if change.VariableID == valueVar.ID && change.ValueChanged {
            fmt.Println("Counter value changed!")
        }
    }
}
```

---

## Features

### Creating Variables

<!-- Spec: main.md (Core Concepts - Variables) -->
<!-- CRC: crc-Variable.md -->

**What it does**: Creates a tracked variable that monitors a value.

**How to access**: Call `tracker.CreateVariable(value, parentID, path, properties)`

**Parameters**:
- `value`: The initial value (for root variables)
- `parentID`: ID of parent variable (0 for root variables)
- `path`: Path string to navigate from parent (see Path Navigation)
- `properties`: Optional metadata map

**Example**:
```go
// Root variable - holds the main data object
root := tracker.CreateVariable(myData, 0, "", nil)

// Child variable - navigates to a field
name := tracker.CreateVariable(nil, root.ID, "Name", nil)

// Nested child - navigates through multiple levels
city := tracker.CreateVariable(nil, root.ID, "Address.City", nil)
```

---

### Path Navigation

<!-- Spec: main.md (Use Cases - Path-Based Navigation) -->
<!-- CRC: crc-Resolver.md -->

**What it does**: Navigates into data structures using dot-separated paths.

**Path Types**:

| Path Element | Navigates To  | Example          |
|--------------|---------------|------------------|
| Field name   | Struct field  | `"Name"`         |
| Key          | Map value     | `"key"`          |
| Index        | Slice element | `"0"`, `"1"`     |
| Method       | Method result | `"GetValue()"`   |
| Combined     | Nested path   | `"Address.City"` |

**Example**:
```go
type Person struct {
    Name    string
    Address Address
    Tags    []string
}

type Address struct {
    City    string
    Country string
}

person := &Person{
    Name:    "Alice",
    Address: Address{City: "Boston", Country: "USA"},
    Tags:    []string{"developer", "manager"},
}

root := tracker.CreateVariable(person, 0, "", nil)

// Navigate to nested struct field
cityVar := tracker.CreateVariable(nil, root.ID, "Address.City", nil)
city, _ := cityVar.Get()  // "Boston"

// Navigate to slice element
tagVar := tracker.CreateVariable(nil, root.ID, "Tags.0", nil)
tag, _ := tagVar.Get()  // "developer"
```

---

### Path with Query Parameters

<!-- Spec: main.md (Use Cases - Path with Properties) -->
<!-- UI: None -->

**What it does**: Combines path navigation with property setting using URL-style syntax.

**Format**: `path?key=value&key2=value2`

**Example**:
```go
// Path with properties
child := tracker.CreateVariable(nil, root.ID, "items.0?label=First&priority=high", nil)
// Path navigates to: items[0]
// Properties set: label="First", priority="high"

// Priority in query sets ValuePriority
critical := tracker.CreateVariable(nil, root.ID, "Status?priority=high", nil)
// This variable's value changes will be high priority
```

---

### Priority Levels

<!-- Spec: main.md (Core Concepts - Priorities) -->
<!-- CRC: crc-Priority.md -->

**What it does**: Assigns importance levels to values and properties.

**Priority Values**:
| Priority | Value | Use Case |
|----------|-------|----------|
| High | 1 | Critical changes that need immediate attention |
| Medium | 0 | Normal changes (default) |
| Low | -1 | Background or optional changes |

**Setting Priority**:
```go
// Via path query parameter
highVar := tracker.CreateVariable(nil, root.ID, "Critical?priority=high", nil)

// Via properties map
lowVar := tracker.CreateVariable(nil, root.ID, "Background",
    map[string]string{"priority": "low"})

// Via SetProperty
someVar.SetProperty("priority", "medium")
```

---

### Change Detection

<!-- Spec: main.md (Core Concepts - Change Detection) -->
<!-- CRC: crc-Tracker.md -->

**What it does**: Compares current values to cached values and identifies changes.

**How to use**:
```go
// After modifying data
data.Value = newValue

// Detect changes - returns sorted changes and clears internal state
changes := tracker.DetectChanges()

// Process all changes sorted by priority
for _, change := range changes {
    fmt.Printf("Variable %d: value=%v, props=%v\n",
        change.VariableID,
        change.ValueChanged,
        change.PropertiesChanged)
}
```

---

### Sorted Changes

<!-- Spec: main.md (Core Concepts - Sorted Changes) -->
<!-- CRC: crc-Change.md -->

**What it does**: DetectChanges returns changes sorted by priority (high to low).

**Change Structure**:
```go
type Change struct {
    VariableID        int64     // Which variable changed
    Priority          Priority  // Priority level
    ValueChanged      bool      // Did the value change?
    PropertiesChanged []string  // Which properties changed
}
```

**Example**:
```go
// Process changes in priority order (returned by DetectChanges)
changes := tracker.DetectChanges()
for _, change := range changes {
    switch change.Priority {
    case ct.PriorityHigh:
        // Handle immediately
        processUrgent(change)
    case ct.PriorityMedium:
        // Handle normally
        processNormal(change)
    case ct.PriorityLow:
        // Handle when convenient
        queueForLater(change)
    }
}
```

---

### Properties

<!-- Spec: api.md (Variable.SetProperty) -->
<!-- CRC: crc-Variable.md -->

**What it does**: Attach metadata to variables with optional priority.

**Setting Properties**:
```go
// Basic property
variable.SetProperty("label", "Counter")

// Property with priority suffix
variable.SetProperty("hint:low", "Optional hint")
variable.SetProperty("error:high", "Critical error message")

// Remove property (empty value)
variable.SetProperty("label", "")
```

**Getting Properties**:
```go
label := variable.GetProperty("label")          // Returns value or ""
priority := variable.GetPropertyPriority("label")  // Returns priority level
```

**Special Properties**:
| Property   | Effect                            |
|------------|-----------------------------------|
| `priority` | Sets the variable's ValuePriority |
| `path`     | Re-parses the navigation path     |

---

### Object Registry

<!-- Spec: main.md (Core Concepts - Object Registry) -->
<!-- CRC: crc-ObjectRegistry.md -->

**What it does**: Maintains identity for Go objects (pointers and maps).

**Why It Matters**:
- Same object in multiple places is recognized as same object
- Prevents cycles in serialization
- Enables consistent change detection

**Automatic Registration**:
```go
// Objects are registered automatically via ToValueJSON
person := &Person{Name: "Alice"}
root := tracker.CreateVariable(person, 0, "", nil)
// person is now registered with its own unique object ID
```

**Looking Up Objects**:
```go
// Look up object's ID
if objID, ok := tracker.LookupObject(myObject); ok {
    fmt.Printf("Object is registered with ID %d\n", objID)
}

// Retrieve object by ID
obj := tracker.GetObject(objID)
```

Note: There is no manual registration API. Objects are registered automatically via `ToValueJSON()` when their `ValueJSON` is computed.

---

### Getting and Setting Values

<!-- Spec: api.md (Variable.Get, Variable.Set) -->
<!-- Sequence: seq-get-value.md, seq-set-value.md -->

**What it does**: Read or modify the value a variable points to.

**Getting Values**:
```go
// Get current value
value, err := variable.Get()
if err != nil {
    // Handle error (invalid path, etc.)
}
```

**Setting Values**:
```go
// Set value (child variables only)
err := variable.Set("new value")
if err != nil {
    // Handle error
}

// Change is detected on next DetectChanges()
tracker.DetectChanges()
```

**Note**: Root variables cannot be set directly (they hold external data).

---

## How-To Guides

### Track a Simple Value

**Goal**: Monitor a single field for changes.

```go
type Counter struct {
    Count int
}

func main() {
    tracker := ct.NewTracker()
    counter := &Counter{Count: 0}

    // Step 1: Create root variable
    root := tracker.CreateVariable(counter, 0, "", nil)

    // Step 2: Create variable for the field
    countVar := tracker.CreateVariable(nil, root.ID, "Count", nil)

    // Step 3: Modify the data
    counter.Count = 100

    // Step 4: Detect and process changes
    changes := tracker.DetectChanges()

    for _, change := range changes {
        if change.VariableID == countVar.ID && change.ValueChanged {
            val, _ := countVar.Get()
            fmt.Printf("Count changed to %v\n", val)
        }
    }
}
```

**Tip**: DetectChanges returns sorted changes and clears internal state automatically.

---

### Track Nested Data

**Goal**: Monitor values deep within a data structure.

```go
type Company struct {
    Name string
    CEO  Person
}

type Person struct {
    Name    string
    Address Address
}

type Address struct {
    City    string
    Country string
}

func main() {
    tracker := ct.NewTracker()

    company := &Company{
        Name: "Acme Corp",
        CEO: Person{
            Name: "Alice",
            Address: Address{City: "Boston", Country: "USA"},
        },
    }

    root := tracker.CreateVariable(company, 0, "", nil)

    // Track nested values using dot paths
    ceoCity := tracker.CreateVariable(nil, root.ID, "CEO.Address.City", nil)
    ceoName := tracker.CreateVariable(nil, root.ID, "CEO.Name", nil)

    // Modify nested value
    company.CEO.Address.City = "New York"

    tracker.DetectChanges()

    // ceoCity is in changed set, ceoName is not
}
```

**Tip**: Use the shortest path that uniquely identifies the value.

---

### Process Changes by Priority

**Goal**: Handle important changes before less important ones.

```go
func main() {
    tracker := ct.NewTracker()
    data := &AppData{Critical: 0, Normal: 0, Background: 0}
    root := tracker.CreateVariable(data, 0, "", nil)

    // Create variables with different priorities
    tracker.CreateVariable(nil, root.ID, "Critical?priority=high", nil)
    tracker.CreateVariable(nil, root.ID, "Normal?priority=medium", nil)
    tracker.CreateVariable(nil, root.ID, "Background?priority=low", nil)

    // Modify all values
    data.Critical = 1
    data.Normal = 1
    data.Background = 1

    // DetectChanges returns sorted changes and clears internal state
    changes := tracker.DetectChanges()

    // Process in priority order
    for _, change := range changes {
        variable := tracker.GetVariable(change.VariableID)

        switch change.Priority {
        case ct.PriorityHigh:
            fmt.Println("URGENT:", variable.GetProperty("path"))
        case ct.PriorityMedium:
            fmt.Println("Normal:", variable.GetProperty("path"))
        case ct.PriorityLow:
            fmt.Println("Low:", variable.GetProperty("path"))
        }
    }
}
```

**Tip**: High-priority changes always come first in the sorted list.

---

### Track Array Elements

**Goal**: Monitor specific elements in a slice or array.

```go
type TaskList struct {
    Tasks []Task
}

type Task struct {
    Title    string
    Complete bool
}

func main() {
    tracker := ct.NewTracker()

    list := &TaskList{
        Tasks: []Task{
            {Title: "Task 1", Complete: false},
            {Title: "Task 2", Complete: false},
            {Title: "Task 3", Complete: false},
        },
    }

    root := tracker.CreateVariable(list, 0, "", nil)

    // Track specific elements by index
    task0 := tracker.CreateVariable(nil, root.ID, "Tasks.0.Complete", nil)
    task1 := tracker.CreateVariable(nil, root.ID, "Tasks.1.Complete", nil)

    // Complete first task
    list.Tasks[0].Complete = true

    // DetectChanges returns sorted changes
    changes := tracker.DetectChanges()

    // Only task0 is in changes
    for _, change := range changes {
        if change.VariableID == task0.ID {
            fmt.Println("Task 0 completion status changed")
        }
    }
    _ = task1 // unused in this example
}
```

**Tip**: Index-based paths use 0-based indexing.

---

### Use Property Priorities

**Goal**: Assign different priorities to different properties on the same variable.

```go
func main() {
    tracker := ct.NewTracker()
    root := tracker.CreateVariable(&MyData{}, 0, "", nil)

    // Create a variable
    v := tracker.CreateVariable(nil, root.ID, "Field", nil)

    // Set properties with different priorities
    v.SetProperty("error:high", "An error occurred")
    v.SetProperty("warning:medium", "A warning")
    v.SetProperty("info:low", "Some information")

    // Get sorted changes (DetectChanges returns them sorted and clears state)
    changes := tracker.DetectChanges()
    for _, change := range changes {
        for _, prop := range change.PropertiesChanged {
            fmt.Printf("Priority %d: %s = %s\n",
                change.Priority,
                prop,
                v.GetProperty(prop))
        }
    }
    // Output:
    // Priority 1: error = An error occurred
    // Priority 0: warning = A warning
    // Priority -1: info = Some information
}
```

**Tip**: The same variable can appear multiple times in sorted changes if it has changes at different priority levels.

---

### Use a Custom Resolver

**Goal**: Navigate custom data structures that aren't standard Go types.

```go
// CustomResolver for JSON-like data
type CustomResolver struct{}

func (r *CustomResolver) Get(obj any, pathElement any) (any, error) {
    m, ok := obj.(map[string]any)
    if !ok {
        return nil, fmt.Errorf("expected map[string]any, got %T", obj)
    }
    key := pathElement.(string)
    val, ok := m[key]
    if !ok {
        return nil, fmt.Errorf("key not found: %s", key)
    }
    return val, nil
}

func (r *CustomResolver) Set(obj any, pathElement any, value any) error {
    m, ok := obj.(map[string]any)
    if !ok {
        return fmt.Errorf("expected map[string]any")
    }
    m[pathElement.(string)] = value
    return nil
}

func main() {
    tracker := ct.NewTracker()
    tracker.Resolver = &CustomResolver{}

    data := map[string]any{
        "user": map[string]any{
            "name": "Alice",
            "age":  30,
        },
    }

    // Register the nested map manually
    tracker.RegisterObject(data["user"], 2)

    root := tracker.CreateVariable(data, 0, "", nil)
    name := tracker.CreateVariable(nil, root.ID, "user.name", nil)

    val, _ := name.Get()
    fmt.Println(val)  // "Alice"
}
```

**Tip**: Custom resolvers must implement both Get and Set methods.

---

## Troubleshooting

### "unregistered pointer/map" Error

**Problem**: ToValueJSON fails with an error about unregistered pointers or maps.

**Cause**: A pointer or map in your data structure hasn't been registered.

**Solution**: Create variables for all pointers and maps before serialization:
```go
// Register nested pointer
tracker.CreateVariable(data.NestedPointer, 0, "", nil)

// Or manually register
tracker.RegisterObject(data.NestedPointer, nextID)
```

---

### Cannot Set Value on Root Variable

**Problem**: `variable.Set()` returns an error for a root variable.

**Cause**: Root variables hold external data and cannot be set directly.

**Solution**: Modify the original data object:
```go
// Wrong: trying to set root
root.Set(newValue)  // Error!

// Right: modify the original object
myData.Field = newValue
tracker.DetectChanges()
```

---

### Changes Not Detected

**Problem**: Modified values aren't appearing in the changed set.

**Cause**: `DetectChanges()` wasn't called after modification.

**Solution**: Always call DetectChanges() after modifying data:
```go
data.Value = newValue
changes := tracker.DetectChanges()  // Required! Returns sorted changes
```

---

### Cannot Set Struct Field

**Problem**: `variable.Set()` fails when trying to set a struct field.

**Cause**: The parent variable holds a struct value, not a pointer.

**Solution**: Use a pointer to the struct:
```go
// Wrong: not a pointer
data := MyStruct{Field: "old"}
root := tracker.CreateVariable(data, 0, "", nil)

// Right: use pointer
data := &MyStruct{Field: "old"}
root := tracker.CreateVariable(data, 0, "", nil)

// Now Set() works
fieldVar := tracker.CreateVariable(nil, root.ID, "Field", nil)
fieldVar.Set("new")  // Works!
```

---

### Method Call Fails

**Problem**: Path with method call (e.g., "GetValue()") returns an error.

**Cause**: Method doesn't meet requirements.

**Requirements for methods**:
- Must be exported (uppercase first letter)
- Must take zero arguments
- Must return at least one value

**Solution**: Ensure method meets requirements:
```go
// Wrong: unexported
func (m *MyType) getValue() int { return m.value }

// Wrong: takes arguments
func (m *MyType) GetValue(x int) int { return x }

// Right: exported, zero args, returns value
func (m *MyType) GetValue() int { return m.value }
```

---

### Stale Change Data

**Problem**: Same changes appearing multiple times.

**Cause**: Previous DetectChanges() result being re-used instead of calling DetectChanges() again.

**Solution**: Call DetectChanges() each time you want to check for changes (it clears internal state automatically):
```go
for {
    // DetectChanges returns sorted changes and clears internal state
    changes := tracker.DetectChanges()

    for _, change := range changes {
        // Process change
    }
}
```
