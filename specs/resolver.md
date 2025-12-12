# Value Resolver Specification

The resolver interface enables navigating into and modifying values within variables.

## Interface

```go
type Resolver interface {
    Get(obj any, pathElement any) (any, error)
    Set(obj any, pathElement any, value any) error
}
```

## Default Resolver (Tracker)

The `Tracker` type implements `Resolver` using Go reflection. A tracker's `Resolver` field defaults to itself, but can be set to a custom resolver.

Variables use their tracker's `Resolver` field for `Get` and `Set` operations.

### Supported Path Elements

| Path Element | Type | Description |
|--------------|------|-------------|
| `"fieldName"` | string | Struct field access |
| `"mapKey"` | string | Map key lookup |
| `"methodName()"` | string (with parens) | Method call |
| `0`, `1`, `2`... | int | Slice/array index |

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

**Method calls:**
```go
type Counter struct { value int }
func (c *Counter) Value() int { return c.value }

c := &Counter{value: 5}
root := tracker.CreateVariable(c, 0, nil)
valVar := tracker.CreateVariable(nil, root.ID, map[string]string{"path": "Value()"})
val, _ := valVar.Get()  // returns 5
```

Method requirements:
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

### Error Conditions

**Get errors:**
- `obj` is nil
- Struct field not found or unexported
- Map key not found
- Method not found or requires arguments
- Method returns no values
- Index out of bounds
- Unsupported type for navigation

**Set errors:**
- `obj` is nil
- Struct field not found, unexported, or not settable
- Need pointer for struct field modification
- Index out of bounds
- Type mismatch between value and target

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
