# Value Resolver Specification

The resolver interface enables navigating into and modifying values within variables.

## Interface

```go
type Resolver interface {
    Get(obj any, pathElement any) (any, error)
    Set(obj any, pathElement any, value any) error
    Call(obj any, methodName string) (any, error)
    CallWith(obj any, methodName string, value any) error
}
```

## Default Resolver (Tracker)

The `Tracker` type implements `Resolver` using Go reflection. A tracker's `Resolver` field defaults to itself, but can be set to a custom resolver.

Variables use their tracker's `Resolver` field for `Get` and `Set` operations.

### Supported Path Elements

| Path Element      | Type                 | Description              |
|-------------------|----------------------|--------------------------|
| `"fieldName"`     | string               | Struct field access      |
| `"mapKey"`        | string               | Map key lookup           |
| `"methodName()"`  | string (with parens) | Zero-arg method (getter) |
| `"methodName(_)"` | string (with `_`)    | One-arg method (setter)  |
| `0`, `1`, `2`...  | int                  | Slice/array index        |

### Path Semantics

**Zero-arg calls `methodName()`:**
- Can appear anywhere in a path (beginning, middle, end)
- Used for navigation (like getters)
- Path ending in `()` is **read-only** (Get succeeds, Set fails)

**One-arg calls `methodName(_)`:**
- Can only appear at the **end** of a path
- Used for writing values (like setters)
- Path ending in `(_)` is **write-only** (Set succeeds, Get fails)

**Examples:**
```go
// Getter in middle of path - Set works on terminal field
path: "Address().City"  // Get: OK, Set: OK (City is settable field)

// Path ends in getter - read-only
path: "Value()"  // Get: OK, Set: ERROR

// Path ends in setter - write-only
path: "SetValue(_)"  // Get: ERROR, Set: OK
```

### Get Operations

**Struct fields:**
```go
type Person struct {
    Name string
    Age  int
}
p := &Person{Name: "Alice", Age: 30}
root := tracker.CreateVariable(p, 0, nil)
nameVar := tracker.CreateVariable(nil, root.ID, map[string]string{"path": "Name"})
name, _ := nameVar.Get()  // returns "Alice"
```

**Map keys:**
```go
m := map[string]int{"count": 42}
root := tracker.CreateVariable(m, 0, nil)
countVar := tracker.CreateVariable(nil, root.ID, map[string]string{"path": "count"})
count, _ := countVar.Get()  // returns 42
```

**Method calls (zero-arg / Call):**
```go
type Counter struct { value int }
func (c *Counter) Value() int { return c.value }

c := &Counter{value: 5}
root := tracker.CreateVariable(c, 0, nil)
valVar := tracker.CreateVariable(nil, root.ID, map[string]string{"path": "Value()"})
val, _ := valVar.Get()  // calls c.Value(), returns 5
```

Method requirements for Call:
- Must be exported
- Must take zero arguments
- Must return at least one value (first return value is used)

**Slice/array indexing:**
```go
items := []string{"a", "b", "c"}
root := tracker.CreateVariable(items, 0, nil)
itemVar := tracker.CreateVariable(nil, root.ID, map[string]string{"path": "1"})
item, _ := itemVar.Get()  // returns "b" (0-based)
```

### Set Operations

**Struct fields:**
```go
p := &Person{Name: "Alice"}
root := tracker.CreateVariable(p, 0, nil)
nameVar := tracker.CreateVariable(nil, root.ID, map[string]string{"path": "Name"})
nameVar.Set("Bob")  // p.Name is now "Bob"
```

Note: For struct fields, the root must hold a pointer to the struct.

**Map keys:**
```go
m := map[string]int{"count": 42}
root := tracker.CreateVariable(m, 0, nil)
countVar := tracker.CreateVariable(nil, root.ID, map[string]string{"path": "count"})
countVar.Set(100)  // m["count"] is now 100
```

**Slice indexing:**
```go
items := []string{"a", "b", "c"}
root := tracker.CreateVariable(items, 0, nil)
itemVar := tracker.CreateVariable(nil, root.ID, map[string]string{"path": "1"})
itemVar.Set("x")  // items is now ["a", "x", "c"]
```

Note: Index must be within bounds.

**Method calls (one-arg / CallWith):**
```go
type Counter struct { value int }
func (c *Counter) SetValue(v int) { c.value = v }

c := &Counter{value: 5}
root := tracker.CreateVariable(c, 0, nil)
valVar := tracker.CreateVariable(nil, root.ID, map[string]string{"path": "SetValue(_)"})
valVar.Set(10)  // calls c.SetValue(10), c.value is now 10
```

Method requirements for CallWith:
- Must be exported
- Must take exactly one argument
- Must not return any values (void only)
- Argument type must be assignable from the passed value

### Error Conditions

**Get errors:**
- `obj` is nil
- Struct field not found or unexported
- Map key not found
- Index out of bounds
- Unsupported type for navigation

**Set errors:**
- `obj` is nil
- Struct field not found, unexported, or not settable
- Need pointer for struct field modification
- Index out of bounds
- Type mismatch between value and target

**Call errors:**
- `obj` is nil
- Method not found or unexported
- Method requires arguments (use CallWith instead)
- Method returns no values

**CallWith errors:**
- `obj` is nil
- Method not found or unexported
- Method doesn't take exactly one argument
- Method returns values (must be void)
- Argument type mismatch

**Path-level errors (Variable Get/Set):**
- Get on path ending in `(_)` → error (write-only path)
- Set on path ending in `()` → error (read-only path)
- `(_)` not at end of path → error (setter must be terminal)

**Access property errors (Variable Get/Set):**
- Get on variable with `access: "w"` → error (write-only variable)
- Get on variable with `access: "action"` → error (action variable)
- Set on variable with `access: "r"` → error (read-only variable)
- Invalid `access` value (not `r`, `w`, `rw`, or `action`) → error

**Access/path combination errors (CreateVariable):**
- `access: "r"` or `access: "rw"` with path ending in `(_)` → error (cannot read from setter)
- `access: "w"` or `access: "rw"` with path ending in `()` → error (use `action` for zero-arg methods)

## Custom Resolvers

Implement the `Resolver` interface for custom navigation logic:

```go
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

func (r *JSONResolver) Call(obj any, methodName string) (any, error) {
    return nil, fmt.Errorf("method calls not supported for JSON data")
}

func (r *JSONResolver) CallWith(obj any, methodName string, value any) error {
    return fmt.Errorf("method calls not supported for JSON data")
}
```

To use a custom resolver, set it on the tracker:

```go
tracker := changetracker.NewTracker()
tracker.Resolver = &JSONResolver{}

// All variables use JSONResolver for path navigation
root := tracker.CreateVariable(data, 0, nil)
child := tracker.CreateVariable(nil, root.ID, map[string]string{"path": "key"})
val, _ := child.Get()
```

## Variable Access Property

Variables support an `access` property that controls read/write permissions:

| Value    | Name | Get | Set | Scanned for Changes | Initial Value Computed |
|----------|------|-----|-----|---------------------|------------------------|
| `rw`     | Read-Write (default) | ✓ | ✓ | ✓ | ✓ |
| `r`      | Read-Only | ✓ | Error | ✓ | ✓ |
| `w`      | Write-Only | Error | ✓ | ✗ | ✓ |
| `action` | Action | Error | ✓ | ✗ | ✗ |

**Usage:**
```go
// Read-only variable - scanned for changes, but Set is an error
readOnly := tracker.CreateVariable(nil, root.ID, "Status?access=r", nil)
val, _ := readOnly.Get()    // OK
readOnly.Set("new")         // ERROR: variable is read-only

// Write-only variable - Set works, but Get fails and not scanned
writeOnly := tracker.CreateVariable(nil, root.ID, "Password?access=w", nil)
writeOnly.Set("secret")     // OK
val, _ := writeOnly.Get()   // ERROR: variable is write-only

// Action variable - for methods that trigger side effects
// Initial value is NOT computed during creation (avoids premature invocation)
action := tracker.CreateVariable(nil, root.ID, "AddContact(_)?access=action", nil)
action.Set(newContact)      // calls AddContact(newContact)
val, _ := action.Get()      // ERROR: variable is action

// Default (read-write) - explicit or omitted
readWrite := tracker.CreateVariable(nil, root.ID, "Name?access=rw", nil)
```

**Action vs Write-Only:**

Both `action` and `w` (write-only) prevent reading the variable's value and exclude it from change detection. The key differences:

- **Write-Only (`w`)**: The initial value is computed during creation. Appropriate for settable fields like `Password` where you want to set values but not read them back. Paths must NOT end with zero-arg methods `()` - use `action` access for those. Paths MAY end with one-arg methods `(_)`.

- **Action (`action`)**: The initial value is **not** computed during creation. Designed for action-triggering methods where navigating the path would invoke the action prematurely. Paths may end with zero-arg `()` or one-arg `(_)` methods. The method is only invoked when `Set()` is explicitly called.

**Path Restrictions by Access Mode:**

| Access | Valid Path Endings | Invalid Path Endings |
|--------|-------------------|---------------------|
| `rw`   | fields, indices | `()`, `(_)` |
| `r`    | fields, indices, `()` | `(_)` |
| `w`    | fields, indices, `(_)` | `()` |
| `action` | `()`, `(_)` | (none) |

- `rw` is a union of `r` and `w`, so it inherits restrictions from both: no `()` (from `w`) and no `(_)` (from `r`)
- Paths ending in `(_)` require `access: "w"` or `access: "action"`
- Paths ending in `()` require `access: "r"` or `access: "action"`
- The difference between `w` and `action` for `(_)` paths: `w` computes the initial value (navigates to the method's receiver), while `action` skips this

**Change Detection:**

- Variables with `access: "r"` or `access: "rw"` are included in change detection scans
- Variables with `access: "w"` or `access: "action"` are excluded from scans (their values cannot be read)

## Path Resolution Patterns

The resolver handles single path elements. Variables use the `path` property (a dot-separated string) to specify multi-level paths:

```go
// Variable with multi-level path
root := tracker.CreateVariable(person, 0, nil)
cityVar := tracker.CreateVariable(nil, root.ID, map[string]string{"path": "Address.City"})
city, _ := cityVar.Get()  // navigates person -> Address -> City
```

The variable's `Get()` method internally parses the path and applies each element:

```go
// Equivalent to:
address, _ := tracker.Get(person, "Address")
city, _ := tracker.Get(address, "City")
```

This keeps the resolver simple (single path elements) while the variable handles multi-level paths.
