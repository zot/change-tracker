# Developer Guide

<!-- Source: design/architecture.md, design/crc-*.md, specs/api.md -->

## Table of Contents

1. [Getting Started](#getting-started)
2. [Project Structure](#project-structure)
3. [Architecture](#architecture)
4. [Development Workflow](#development-workflow)
5. [Adding Features](#adding-features)
6. [Testing](#testing)
7. [Build and Deployment](#build-and-deployment)

---

## Getting Started

### Prerequisites

- **Go 1.24+**: Required for weak reference support (`weak.Pointer`)
- **Git**: For version control

### Installation

```bash
# Clone the repository
git clone https://github.com/zot/change-tracker.git
cd change-tracker

# Verify Go version
go version  # Should be 1.24 or higher

# Download dependencies
go mod download

# Verify installation
go build ./...
```

### Running the Project

The change-tracker is a library package, not an executable. To use it in your project:

```go
import "github.com/zot/change-tracker"
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

---

## Project Structure

```
change-tracker/
|-- specs/              # Level 1: Human specifications
|   |-- main.md         # Main specification
|   |-- api.md          # API documentation
|   |-- resolver.md     # Resolver specification
|   `-- value-json.md   # Value JSON specification
|
|-- design/             # Level 2: Design models
|   |-- architecture.md # Architecture overview
|   |-- traceability.md # Spec-to-design mapping
|   |-- crc-*.md        # CRC cards for components
|   |-- seq-*.md        # Sequence diagrams
|   |-- test-*.md       # Test design specifications
|   `-- gap-analysis.md # Design completeness analysis
|
|-- docs/               # Generated documentation
|   |-- requirements.md # Requirements documentation
|   |-- design.md       # Design overview
|   |-- developer-guide.md # This file
|   `-- user-manual.md  # User documentation
|
|-- *.go                # Level 3: Implementation
`-- *_test.go           # Test implementation
```

### Three-Tier Design System

The project uses a CRC (Class-Responsibility-Collaborator) modeling workflow:

```
Level 1: Human specs (specs/*.md)
   |
   v
Level 2: Design models (design/*.md)
   |
   v
Level 3: Implementation (*.go)
```

**Important**: Do NOT generate code directly from Level 1 specs. Always create Level 2 design first.

---

## Architecture

### System Overview

The package is organized into four interconnected systems:

| System           | Purpose                                     | Key Components                      |
|------------------|---------------------------------------------|-------------------------------------|
| Core Tracking    | Variable management and change detection    | Tracker, Variable, Priority, Change |
| Value Resolution | Navigate and modify values via paths        | Resolver interface                  |
| Serialization    | Convert values to Value JSON format         | ObjectRef                           |
| Object Registry  | Weak reference tracking for object identity | Internal to Tracker                 |

### Key Components

#### Tracker
<!-- CRC: crc-Tracker.md -->

Central coordinator that:
- Creates and manages variables
- Maintains the object registry
- Tracks changes (value and property)
- Provides sorted change access
- Implements the Resolver interface

#### Variable
<!-- CRC: crc-Variable.md -->

Represents a tracked value with:
- Unique ID and parent-child relationships
- Path-based navigation
- Properties with priorities
- Cached value and ValueJSON

#### Resolver
<!-- CRC: crc-Resolver.md -->

Interface for navigating into values:
```go
type Resolver interface {
    Get(obj any, pathElement any) (any, error)
    Set(obj any, pathElement any, value any) error
}
```

### Design Patterns Used

| Pattern   | Usage                               |
|-----------|-------------------------------------|
| Strategy  | Pluggable Resolver implementations  |
| Observer  | Change accumulation and retrieval   |
| Composite | Variable parent-child hierarchy     |
| Flyweight | Object registry for shared identity |

---

## Development Workflow

### Reading Design Documents

Before implementing or modifying features:

1. **Read architecture.md first**: Understand the system structure
2. **Find relevant CRC cards**: Identify affected components
3. **Review sequence diagrams**: Understand workflows
4. **Check test designs**: Know what needs testing

### Traceability Comments

All source files should include traceability comments:

```go
// Tracker manages variables and change detection.
// CRC: crc-Tracker.md
// Spec: main.md (Core Concepts - Tracker)
type Tracker struct {
    // ...
}

// CreateVariable creates a new tracked variable.
// Sequence: seq-create-variable.md
func (t *Tracker) CreateVariable(value any, parentID int64, path string, properties map[string]string) *Variable {
    // ...
}
```

**Format Rules**:
- Use simple filenames WITHOUT directory paths
- Correct: `CRC: crc-Person.md`, `Spec: main.md`
- Wrong: `CRC: design/crc-Person.md`

### Finding Implementations

To find where a design element is implemented:

```bash
grep "seq-get-file.md" *.go
grep "crc-Tracker.md" *.go
```

---

## Adding Features

### Step 1: Update Specifications

If the feature requires new capabilities, update the relevant specs in `specs/`:

- `main.md` for core concepts
- `api.md` for API changes
- `resolver.md` for resolver changes
- `value-json.md` for serialization changes

### Step 2: Create Level 2 Design

Use the designer agent or manually create:

1. **CRC Cards**: `design/crc-NewComponent.md`
   - Responsibilities (Knows/Does)
   - Collaborators
   - Sequence references

2. **Sequence Diagrams**: `design/seq-new-workflow.md`
   - ASCII art diagram
   - Participants
   - Notes

3. **Test Designs**: `design/test-NewComponent.md`
   - Test scenarios
   - Error scenarios
   - Expected inputs/outputs

### Step 3: Update Architecture

Add the new component to `design/architecture.md`:

```markdown
### System Name
**Purpose**: What it does
**Design Elements**: crc-NewComponent.md, seq-new-workflow.md, test-NewComponent.md
```

### Step 4: Update Traceability

Add mappings to `design/traceability.md`:

- Level 1 to Level 2 (spec to design)
- Level 2 to Level 3 (design to implementation)

### Step 5: Implement

1. Read the test design first
2. Implement the feature following the CRC card and sequences
3. Add traceability comments to the code
4. Write tests following the test design

### Step 6: Run Gap Analysis

After implementation, run the gap-analyzer agent to verify completeness.

---

## Testing

### Test Design Methodology

Tests are designed at Level 2 before implementation:

1. **Test designs** (`design/test-*.md`) specify what to test
2. **Test code** (`*_test.go`) implements the specifications
3. Every test scenario in design must have corresponding test code

### Test File Structure

```go
// Test Design: test-Tracker.md

package changetracker

import "testing"

// TestNewTracker tests tracker creation.
// Test IDs: T1.1, T1.2
func TestNewTracker(t *testing.T) {
    // T1.1: Create new tracker
    tracker := NewTracker()
    if tracker == nil {
        t.Error("expected non-nil tracker")
    }

    // T1.2: Initial state
    if tracker.Resolver != tracker {
        t.Error("expected Resolver to default to self")
    }
}
```

### Test Categories

| Category | Description | Example |
|----------|-------------|---------|
| Unit Tests | Test individual methods | `TestCreateVariable` |
| Error Tests | Test error conditions | `TestGetInvalidPath` |
| Integration Tests | Test workflows | `TestFullChangeLifecycle` |

### Test Data Structures

Common test types:

```go
type TestPerson struct {
    Name    string
    Age     int
    Address TestAddress
}

type TestAddress struct {
    City    string
    Country string
}

type TestCounter struct {
    value int
}

func (c *TestCounter) Value() int {
    return c.value
}
```

### Running Specific Tests

```bash
# Run tests matching pattern
go test -run TestTracker ./...

# Run tests in a specific file
go test -v ./... -run "Test.*Variable"

# Run with race detector
go test -race ./...
```

---

## Build and Deployment

### Building

```bash
# Build the package
go build ./...

# Check for issues
go vet ./...

# Format code
go fmt ./...
```

### Code Quality

```bash
# Run staticcheck (install first: go install honnef.co/go/tools/cmd/staticcheck@latest)
staticcheck ./...

# Run golangci-lint (install first)
golangci-lint run
```

### Publishing

The package is published to pkg.go.dev automatically when tagged:

```bash
# Tag a release
git tag v1.0.0
git push origin v1.0.0
```

### API Stability

- Public API is defined in `specs/api.md`
- Breaking changes require major version bump
- Internal implementation can change freely

---

## Code Examples

### Basic Usage

```go
package main

import (
    ct "github.com/zot/change-tracker"
)

func main() {
    // Create tracker
    tracker := ct.NewTracker()

    // Create root variable
    data := &MyData{Count: 0}
    root := tracker.CreateVariable(data, 0, "", nil)

    // Create child variable for Count field
    countVar := tracker.CreateVariable(nil, root.ID, "Count", nil)

    // Modify value
    data.Count = 42

    // Detect changes - returns sorted changes and clears internal state
    changes := tracker.DetectChanges()

    // Process changes
    for _, change := range changes {
        fmt.Printf("Variable %d changed: value=%v\n",
            change.VariableID, change.ValueChanged)
    }
}
```

### Custom Resolver

```go
// JSONResolver navigates JSON-like map structures
type JSONResolver struct{}

func (r *JSONResolver) Get(obj any, pathElement any) (any, error) {
    m, ok := obj.(map[string]any)
    if !ok {
        return nil, fmt.Errorf("expected map[string]any")
    }
    key, ok := pathElement.(string)
    if !ok {
        return nil, fmt.Errorf("expected string key")
    }
    val, ok := m[key]
    if !ok {
        return nil, fmt.Errorf("key %q not found", key)
    }
    return val, nil
}

func (r *JSONResolver) Set(obj any, pathElement any, value any) error {
    m, ok := obj.(map[string]any)
    if !ok {
        return fmt.Errorf("expected map[string]any")
    }
    key, ok := pathElement.(string)
    if !ok {
        return fmt.Errorf("expected string key")
    }
    m[key] = value
    return nil
}

// Usage
tracker := ct.NewTracker()
tracker.Resolver = &JSONResolver{}
```

### Priority-Based Processing

```go
// Create variables with different priorities
highPriority := tracker.CreateVariable(nil, root.ID, "Critical?priority=high", nil)
lowPriority := tracker.CreateVariable(nil, root.ID, "Optional?priority=low", nil)

// Set property with priority
highPriority.SetProperty("status:high", "active")
lowPriority.SetProperty("hint:low", "some hint")

// DetectChanges returns changes sorted by priority
changes := tracker.DetectChanges()
for _, change := range changes {
    // High priority changes come first
    if change.Priority == ct.PriorityHigh {
        // Handle urgently
    }
}
```

---

## Troubleshooting

### Common Issues

**Issue**: Unregistered pointer/map error during ToValueJSON

**Solution**: Ensure all pointers and maps are registered via CreateVariable before serialization:
```go
// Register the pointer by creating a variable for it
tracker.CreateVariable(myPointer, 0, "", nil)
```

**Issue**: Cannot set struct field

**Solution**: The parent must hold a pointer to the struct:
```go
// Wrong: data is not a pointer
data := MyStruct{}
root := tracker.CreateVariable(data, 0, "", nil)

// Right: data is a pointer
data := &MyStruct{}
root := tracker.CreateVariable(data, 0, "", nil)
```

**Issue**: Method call fails

**Solution**: Methods must be exported, take zero arguments, and return at least one value:
```go
// Wrong: unexported method
func (m *MyType) getValue() int { return m.value }

// Right: exported method
func (m *MyType) GetValue() int { return m.value }
```

**Issue**: Changes not detected

**Solution**: Remember to call DetectChanges():
```go
data.Value = newValue
changes := tracker.DetectChanges()  // Must call this! Returns sorted changes
```
