# Design Documentation

<!-- Source: design/architecture.md, design/crc-*.md, design/seq-*.md -->

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [System Components](#system-components)
3. [Design Patterns](#design-patterns)
4. [Data Flow](#data-flow)
5. [Key Design Decisions](#key-design-decisions)

---

## Architecture Overview

<!-- Source: architecture.md -->

The Change Tracker package is organized into four interconnected systems:

```
+-------------------------+
|   Core Tracking System  |
|  (Tracker, Variable,    |
|   Priority, Change)     |
+-----------+-------------+
            |
    +-------+-------+-------+
    |               |       |
    v               v       v
+-------+    +----------+  +---------+
| Value |    | Serial-  |  | Object  |
| Resol-|    | ization  |  | Registry|
| ution |    | System   |  | System  |
+-------+    +----------+  +---------+
```

### System Dependencies

```
Core Tracking System
    |
    +---> Value Resolution System (for path navigation)
    |
    +---> Serialization System (for change detection)
    |
    +---> Object Registry System (for object identity)
```

---

## System Components

### Core Tracking System

**Purpose**: Variable management and change detection

#### Tracker
<!-- CRC: crc-Tracker.md -->

**Purpose**: Central coordinator for variable management and change detection.

**Responsibilities**:
- Create and manage variables with unique IDs
- Maintain object registry (weak map from objects to variable IDs)
- Track which variables have changed (value and property changes)
- Provide sorted access to changes by priority
- Serialize values to Value JSON form
- Implement default resolver using Go reflection

**Collaborates With**: Variable, Resolver, ObjectRef, Change, Priority

**Key Fields**:
- `variables`: map[int64]*Variable - all tracked variables
- `nextID`: int64 - next ID to assign (starts at 1)
- `valueChanges`: map[int64]bool - variables with value changes
- `propertyChanges`: map[int64][]string - property changes per variable
- `sortedChanges`: []Change - reusable slice for sortChanges output
- `objectRegistry`: weak map from object pointers to variable IDs
- `Resolver`: pluggable resolver (defaults to self)

#### Variable
<!-- CRC: crc-Variable.md -->

**Purpose**: Represents a tracked value with metadata and parent-child relationships.

**Responsibilities**:
- Navigate to current value using path from parent
- Set values via path navigation
- Manage properties with priority support
- Cache value and ValueJSON for change detection

**Collaborates With**: Tracker, Resolver, Priority

**Key Fields**:
- `ID`: int64 - unique identifier
- `ParentID`: int64 - parent variable (0 = root)
- `Properties`: map[string]string - metadata
- `PropertyPriorities`: map[string]Priority - priority per property
- `Path`: []any - parsed path elements
- `Value`: any - cached value for child navigation
- `ValueJSON`: any - cached Value JSON for change detection
- `ValuePriority`: Priority - priority level for value changes
- `tracker`: *Tracker - reference to owning tracker

#### Priority
<!-- CRC: crc-Priority.md -->

**Purpose**: Define priority levels for values and properties.

**Design Pattern**: Enumeration type

**Values**:
- `PriorityLow` (-1): Low priority
- `PriorityMedium` (0): Medium priority (default)
- `PriorityHigh` (1): High priority

**Collaborates With**: Variable

#### Change
<!-- CRC: crc-Change.md -->

**Purpose**: Represent a change record for sorted change reporting.

**Responsibilities**: Data container for change information.

**Key Fields**:
- `VariableID`: int64 - which variable changed
- `Priority`: Priority - priority level of this change entry
- `ValueChanged`: bool - whether value changed
- `PropertiesChanged`: []string - property names that changed at this priority

**Note**: A single variable may produce multiple Change entries if its value and properties have different priorities.

---

### Value Resolution System

**Purpose**: Navigate and modify values via paths

#### Resolver (Interface)
<!-- CRC: crc-Resolver.md -->

**Purpose**: Define interface for navigating into values.

**Interface Methods**:
- `Get(obj any, pathElement any) (any, error)`: Retrieve value at path element
- `Set(obj any, pathElement any, value any) error`: Assign value at path element

**Path Element Types**:
| Type | Usage |
|------|-------|
| string | Struct field, map key, or method name (with "()" suffix) |
| int | Slice/array index (0-based) |

**Default Implementation**: Tracker implements Resolver using Go reflection

**Collaborates With**: Tracker, Variable

---

### Serialization System

**Purpose**: Convert values to Value JSON format

#### ObjectRef
<!-- CRC: crc-ObjectRef.md -->

**Purpose**: Represent registered objects in Value JSON form.

**Responsibilities**: Data container for object reference.

**Structure**:
```go
type ObjectRef struct {
    Obj int64 `json:"obj"`
}
```

**JSON Representation**: `{"obj": 123}`

**Helper Functions**:
- `IsObjectRef(value any) bool`: Check if value is an ObjectRef
- `GetObjectRefID(value any) (int64, bool)`: Extract ID from ObjectRef

**Collaborates With**: Tracker

---

### Object Registry System

**Purpose**: Weak reference tracking for object identity

#### ObjectRegistry
<!-- CRC: crc-ObjectRegistry.md -->

**Purpose**: Maintain weak references from objects to variable IDs.

**Note**: Internal to Tracker, not a standalone type.

**Responsibilities**:
- Register objects with variable IDs
- Unregister objects
- Look up variable IDs for objects
- Retrieve objects by variable ID
- Automatic cleanup when objects are garbage collected

**Implementation Details**:
- Uses Go 1.24+ `weak.Pointer` for weak references
- Internal structure:
```go
type weakEntry struct {
    weak  weak.Pointer[any]  // weak reference to object
    objID int64              // object ID for ObjectRef serialization
}
```

**Collaborates With**: Tracker, Variable

---

## Design Patterns

### Strategy Pattern: Resolver
<!-- CRC: crc-Resolver.md -->

The Resolver interface allows pluggable navigation strategies:
- Default: Tracker's reflection-based implementation
- Custom: Any implementation of the Resolver interface

```go
tracker := changetracker.NewTracker()
tracker.Resolver = &MyCustomResolver{}  // Inject custom strategy
```

### Observer Pattern: Change Tracking

Variables don't notify observers directly. Instead:
1. Changes accumulate in the tracker (value changes detected on call, property changes recorded immediately)
2. Clients call `DetectChanges()` to detect value changes, get sorted changes, and clear internal state
3. Returned changes are sorted by priority (High -> Medium -> Low)

### Composite Pattern: Variable Hierarchy

Variables form a tree structure:
- Root variables (ParentID = 0) hold external values
- Child variables navigate from parent's value using path
- Multi-level hierarchies are supported

### Flyweight Pattern: Object Registry

The same object appearing in multiple locations shares a single identity:
- Objects are registered once with a variable ID
- All references to the same object serialize to the same `{"obj": id}`
- Prevents duplication and enables cycle detection

---

## Data Flow

### Variable Creation Flow
<!-- Sequence: seq-create-variable.md -->

```
Client -> Tracker.CreateVariable(value, parentID, path, props)
    |
    +-> Assign next ID
    +-> Parse path and query parameters
    +-> Merge properties (props map, then query params)
    +-> Parse path string into elements
    +-> Set ValuePriority from "priority" property
    +-> If root: cache value directly
    +-> If child: navigate from parent, cache result
    +-> If pointer/map: register in object registry
    +-> Convert to ValueJSON, cache for change detection
    +-> Store variable in tracker
    |
    <- Return *Variable
```

### Change Detection Flow
<!-- Sequence: seq-detect-changes.md -->

```
Client -> Tracker.DetectChanges()
    |
    +-> For each variable:
    |       +-> Get current value (navigate if child)
    |       +-> Convert to ValueJSON
    |       +-> Compare to cached ValueJSON
    |       +-> If different: mark as value change
    |       +-> Update cached Value and ValueJSON
    +-> Sort changes by priority (internal sortChanges)
    |       +-> Collect value and property changes
    |       +-> Group by priority level
    |       +-> Sort High -> Medium -> Low
    +-> Clear internal change records
    |
    <- Return []Change (sorted)
```

### Value Get Flow
<!-- Sequence: seq-get-value.md -->

```
Client -> Variable.Get()
    |
    +-> If root: return cached Value
    +-> If child:
    |       +-> Get parent variable
    |       +-> Start with parent's cached Value
    |       +-> For each path element:
    |       |       +-> Call Resolver.Get(val, element)
    |       |       +-> val = result
    |       +-> Cache final value in Variable.Value
    |
    <- Return value, error
```

### Value Set Flow
<!-- Sequence: seq-set-value.md -->

```
Client -> Variable.Set(newValue)
    |
    +-> If root: error (cannot set root directly)
    +-> If no path: error
    +-> Get parent variable
    +-> Start with parent's cached Value
    +-> Navigate all but last path element
    +-> Call Resolver.Set(val, lastElement, newValue)
    |
    <- Return error or nil
```

### Property Set Flow
<!-- Sequence: seq-set-property.md -->

```
Client -> Variable.SetProperty(name, value)
    |
    +-> Parse name for priority suffix (:low/:medium/:high)
    +-> If value empty: delete property and priority
    +-> Else: set property and priority
    +-> If name is "priority": update ValuePriority
    +-> If name is "path": re-parse Path field
    +-> Record property change in tracker
    +-> Add variable to changed set
    |
    <- Return (void)
```

### Internal Sort Changes

The sorting logic is internal to DetectChanges:
- Reset sortedChanges slice (reuse capacity)
- For each changed variable:
  - If value changed: Create Change at ValuePriority
  - For each changed property: Group by property priority
  - For each priority group: Create/update Change with property names
- Sort by Priority descending (High -> Medium -> Low)
- Return sorted slice

### Value JSON Serialization Flow
<!-- Sequence: seq-to-value-json.md -->

```
Client -> Tracker.ToValueJSON(value)
    |
    +-> If nil: return nil
    +-> If primitive: return value unchanged
    +-> If slice/array:
    |       +-> Recursively convert each element
    |       +-> Return array of converted elements
    +-> If pointer/map:
    |       +-> Look up in object registry
    |       +-> If found: return ObjectRef{Obj: id}
    |       +-> If not found: error (must be registered)
    |
    <- Return converted value or error
```

---

## Key Design Decisions

### Decision 1: Weak References for Object Registry
<!-- Source: main.md (Object Registry) -->

**Context**: Objects need consistent identity without preventing garbage collection.

**Decision**: Use Go 1.24+ `weak.Pointer` for object registry entries.

**Rationale**:
- Objects can be garbage collected when no longer referenced by application
- No memory leaks from long-running tracking
- Registry automatically cleans up collected objects

**Trade-off**: Requires Go 1.24+; registry lookups may return nil for collected objects.

### Decision 2: Explicit Change Detection
<!-- Source: main.md (Design Principles) -->

**Context**: Applications need to know when values change.

**Decision**: Changes are only detected when `DetectChanges()` is explicitly called.

**Rationale**:
- Predictable performance (no hidden operations)
- Application controls when detection happens
- No threading complexity

**Trade-off**: Application must remember to call DetectChanges().

### Decision 3: Value JSON for Comparison
<!-- Source: value-json.md -->

**Context**: Need to compare complex Go values for change detection.

**Decision**: Convert values to Value JSON form, then compare Value JSON.

**Rationale**:
- Object references break cycles
- Same object in multiple places compares equal
- Compact representation

**Trade-off**: Pointers/maps must be registered before serialization.

### Decision 4: Tracker as Default Resolver
<!-- Source: resolver.md -->

**Context**: Variables need to navigate into values.

**Decision**: Tracker implements Resolver interface using reflection.

**Rationale**:
- Works out of the box for common Go types
- No additional dependencies needed
- Can be overridden with custom resolver

**Trade-off**: Reflection has runtime overhead; limited to exported fields.

### Decision 5: Priority Suffixes in Property Names
<!-- Source: api.md (SetProperty) -->

**Context**: Properties need priority levels for sorted reporting.

**Decision**: Priority is specified via suffix on property name (e.g., "label:high").

**Rationale**:
- Compact syntax
- Priority and value set in single call
- Stored separately for efficient lookup

**Trade-off**: Property names cannot contain colons at end.

### Decision 6: Reusable sortedChanges Slice
<!-- Source: api.md (DetectChanges) -->

**Context**: DetectChanges may be called frequently.

**Decision**: Reuse internal slice to avoid allocations.

**Rationale**:
- Better performance for hot paths
- Reduced GC pressure

**Trade-off**: Returned slice only valid until next DetectChanges().
