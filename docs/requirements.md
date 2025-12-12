# Requirements Documentation

<!-- Source: main.md, api.md, resolver.md, value-json.md -->

## Table of Contents

1. [Overview](#overview)
2. [Business Requirements](#business-requirements)
3. [Functional Requirements](#functional-requirements)
4. [Non-Functional Requirements](#non-functional-requirements)
5. [Technical Constraints](#technical-constraints)
6. [Out of Scope](#out-of-scope)

---

## Overview

<!-- Source: main.md (Overview) -->

The Change Tracker is a Go package (`github.com/zot/change-tracker`) that provides variable management with automatic change detection. It enables applications to track values, detect when they change, and report those changes in priority order.

**Key Capabilities:**
- A change tracker that manages variables and detects changes
- Variables that hold values and track parent-child relationships
- Object registry with weak references for consistent object identity
- Value JSON serialization with object references
- Change detection via value comparison
- Pluggable value resolution for navigating into objects

---

## Business Requirements

### BR1: Variable Management
<!-- Source: main.md (Core Concepts - Variables) -->
**Description**: The system shall provide a mechanism to create, manage, and destroy tracked variables that represent values in an application.

**Rationale**: Applications need to track specific values within their data structures and be notified when those values change.

### BR2: Change Detection
<!-- Source: main.md (Core Concepts - Change Detection) -->
**Description**: The system shall detect when tracked variable values have changed since the last detection cycle.

**Rationale**: Applications need to know which values have changed so they can respond appropriately (update UI, persist data, notify observers, etc.).

### BR3: Priority-Based Reporting
<!-- Source: main.md (Core Concepts - Sorted Changes) -->
**Description**: The system shall support priority levels for values and properties, and report changes sorted by priority.

**Rationale**: Not all changes are equally urgent; applications need to process high-priority changes before low-priority ones.

### BR4: Object Identity
<!-- Source: main.md (Core Concepts - Object Registry) -->
**Description**: The system shall maintain consistent identity for Go objects (pointers and maps) across the tracking system.

**Rationale**: The same object appearing in multiple places should be recognized as the same object for change detection purposes.

---

## Functional Requirements

### FR1: Tracker Creation
<!-- Source: main.md (Core Concepts - Tracker), api.md (NewTracker) -->
**Description**: Create a new change tracker instance.

**Acceptance Criteria**:
- NewTracker() returns a non-nil Tracker instance
- The tracker's Resolver field defaults to itself
- The tracker starts with no variables and an empty changed set

**Priority**: High

### FR2: Variable Creation
<!-- Source: main.md (Core Concepts - Variables), api.md (CreateVariable) -->
**Description**: Create tracked variables with support for parent-child relationships, paths, and properties.

**Acceptance Criteria**:
- Variables receive unique sequential IDs starting from 1
- Root variables (parentID=0) cache the provided value directly
- Child variables compute their value by navigating from parent using path
- Path can include URL-style query parameters for properties
- Query properties override properties map
- The "priority" property sets the variable's ValuePriority
- Pointers and maps are automatically registered in the object registry
- Initial ValueJSON is cached for change detection
- Nil properties map is initialized to empty map

**Priority**: High

### FR3: Variable Retrieval
<!-- Source: api.md (GetVariable) -->
**Description**: Retrieve variables by their ID.

**Acceptance Criteria**:
- GetVariable(id) returns the variable if it exists
- GetVariable(id) returns nil if the variable does not exist

**Priority**: High

### FR4: Variable Destruction
<!-- Source: api.md (DestroyVariable) -->
**Description**: Remove variables from the tracker.

**Acceptance Criteria**:
- Variable is removed from the tracker
- Variable is removed from the changed set if present
- Associated object is unregistered from the object registry
- Operation is a no-op for non-existent IDs

**Priority**: High

### FR5: Value Navigation - Get
<!-- Source: api.md (Variable.Get), resolver.md -->
**Description**: Get a variable's value by navigating from parent using path.

**Acceptance Criteria**:
- Root variables return their cached Value directly
- Child variables navigate from parent's cached Value using path elements
- Each path element is resolved via the tracker's Resolver
- Result is cached in Variable.Value for child navigation
- Errors propagate if path resolution fails

**Priority**: High

### FR6: Value Navigation - Set
<!-- Source: api.md (Variable.Set), resolver.md -->
**Description**: Set a variable's value by navigating from parent using path.

**Acceptance Criteria**:
- Navigate to parent of target using all but last path element
- Use resolver's Set to assign value at final path element
- Cannot set root variables directly (they hold external values)
- Struct field setting requires pointer to struct
- Slice index must be within bounds

**Priority**: High

### FR7: Change Detection
<!-- Source: main.md (Core Concepts - Change Detection), api.md (DetectChanges) -->
**Description**: Detect which variables have changed since the last detection and return sorted changes.

**Acceptance Criteria**:
- For each variable, get current value and convert to Value JSON
- Compare current Value JSON to stored Value JSON
- If different, mark variable's value as changed
- Update stored Value JSON to current Value JSON
- Sort all changes (value and property) by priority (High -> Medium -> Low)
- Clear internal change records after sorting
- Return the sorted []Change
- A variable may appear multiple times if it has changes at different priority levels
- Properties at the same priority are grouped into one Change entry
- Reuses internal slice to minimize allocations
- Returned slice is valid until the next call to DetectChanges()

**Priority**: High

### FR8: Property Management
<!-- Source: api.md (Variable.SetProperty, GetProperty) -->
**Description**: Get and set variable properties with priority support.

**Acceptance Criteria**:
- GetProperty returns property value or empty string
- SetProperty with empty value removes the property
- Property names can include priority suffix (:low, :medium, :high)
- Setting "priority" property updates ValuePriority
- Setting "path" property re-parses and updates Path field
- Property changes are recorded immediately in the tracker

**Priority**: High

### FR9: Object Registry
<!-- Source: main.md (Core Concepts - Object Registry), api.md (Object Registry Methods) -->
**Description**: Maintain a weak map from Go objects to variable IDs.

**Acceptance Criteria**:
- RegisterObject stores weak reference to object with variable ID
- UnregisterObject removes object from registry
- LookupObject finds variable ID for registered object
- GetObject retrieves object by variable ID (may return nil if collected)
- Uses Go 1.24+ weak references
- Objects can be garbage collected when no longer referenced

**Priority**: High

### FR10: Value JSON Serialization
<!-- Source: value-json.md, api.md (Value JSON Methods) -->
**Description**: Serialize values to Value JSON format.

**Acceptance Criteria**:
- Primitives (string, number, bool, nil) pass through unchanged
- Registered pointers and maps become ObjectRef{Obj: id}
- Slices/arrays become arrays with elements in Value JSON form
- Unregistered pointers/maps cause an error
- ToValueJSONBytes returns JSON-encoded bytes

**Priority**: High

### FR11: Path Resolution
<!-- Source: resolver.md -->
**Description**: Navigate into values using path elements.

**Acceptance Criteria**:
- String path elements: struct field, map key, or method name (with "()" suffix)
- Integer path elements: slice/array index (0-based)
- Methods must be exported, zero-argument, return at least one value
- Appropriate errors for nil objects, missing fields, out of bounds, etc.

**Priority**: High

### FR12: Variable Listing
<!-- Source: api.md (Variables, RootVariables, Children) -->
**Description**: List variables in the tracker.

**Acceptance Criteria**:
- Variables() returns all variables
- RootVariables() returns variables with parentID=0
- Children(parentID) returns child variables of a parent

**Priority**: Medium

---

## Non-Functional Requirements

### NFR1: Simplicity
<!-- Source: main.md (Design Principles - Simplicity First) -->
**Description**: The package shall be a pure data structure library with minimal dependencies.

**Acceptance Criteria**:
- No external dependencies beyond Go standard library
- No thread safety mechanisms (callers handle synchronization)
- Clean, straightforward API

### NFR2: Explicit Operations
<!-- Source: main.md (Design Principles - Explicit Over Automatic) -->
**Description**: Changes shall only be detected when explicitly requested.

**Acceptance Criteria**:
- DetectChanges() must be called to detect value changes
- DetectChanges() clears internal state and returns sorted changes
- No automatic background processing

### NFR3: Memory Efficiency
<!-- Source: main.md (Core Concepts - Object Registry) -->
**Description**: The object registry shall not prevent garbage collection.

**Acceptance Criteria**:
- Uses weak references for registered objects
- Objects can be garbage collected when no longer referenced by application code
- Registry entries are automatically cleaned up when objects are collected

### NFR4: Allocation Efficiency
<!-- Source: api.md (DetectChanges) -->
**Description**: Minimize memory allocations in hot paths.

**Acceptance Criteria**:
- DetectChanges() reuses internal slice for sorted changes
- Returned slice valid until next DetectChanges()

### NFR5: Frictionless Integration
<!-- Source: main.md (Core Concepts - Object Registry) -->
**Description**: Domain objects shall require no modification for tracking.

**Acceptance Criteria**:
- No interfaces to implement
- No embedded IDs required
- Works with existing Go structs, maps, and slices

---

## Technical Constraints

### TC1: Go Version
<!-- Source: main.md (Core Concepts - Object Registry) -->
**Description**: Requires Go 1.24+ for weak reference support.

### TC2: Registered Objects
<!-- Source: value-json.md (Registration Rules) -->
**Description**: All pointers and maps must be registered before serialization to Value JSON.

### TC3: Struct Field Settability
<!-- Source: resolver.md (Set Operations) -->
**Description**: Setting struct fields requires a pointer to the struct.

### TC4: Method Requirements
<!-- Source: resolver.md (Get Operations) -->
**Description**: Methods called via resolver must be exported, take zero arguments, and return at least one value.

---

## Out of Scope

The following features are explicitly not part of this package:

1. **Thread Safety**: Callers are responsible for synchronization if needed
2. **Persistence**: No built-in save/load functionality
3. **Network Communication**: No remote change notification
4. **Automatic Change Detection**: Changes are only detected when explicitly requested
5. **Undo/Redo**: No change history or rollback capability
6. **Struct Serialization**: Value JSON only serializes object references, not struct contents
