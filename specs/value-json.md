# Value JSON Specification

Value JSON is a serialization format that represents Go values with object references for registered objects.

## Purpose

Value JSON solves several problems:

1. **Change Detection**: Variables store their last known Value JSON and compare against current Value JSON to detect changes
2. **Object Identity**: The same object appearing in multiple places serializes to the same reference
3. **Cycle Handling**: Registered objects break reference cycles
4. **Compact Output**: Objects referenced multiple times appear once as data, elsewhere as references

## Format

Value JSON has exactly three value types:

### Primitives

Primitives serialize as standard JSON:

| Go Type | JSON |
|---------|------|
| `string` | `"value"` |
| `int`, `int64`, etc. | `123` |
| `float64`, etc. | `1.23` |
| `bool` | `true` / `false` |
| `nil` | `null` |

### Arrays

Arrays and slices serialize as JSON arrays with elements in Value JSON form:

```go
// Go
items := []*Person{alice, bob}  // both registered

// Value JSON
[{"obj": 1}, {"obj": 2}]
```

### Object References

Registered objects (pointers and maps) serialize as:

```json
{"obj": 123}
```

Where `123` is the variable ID associated with the object.

## Registration Rules

Objects are registered in the object registry when:

1. **CreateVariable with pointer/map**: `tracker.CreateVariable(val, parentID, props)` automatically registers pointers and maps
2. **Manual registration**: `tracker.RegisterObject(val, varID)` explicitly registers

Only pointers and maps can be registered and serialized. Unregistered pointers/maps cause an error during serialization.

## Serialization Algorithm

```
ToValueJSON(value):
    if value is nil:
        return nil
    if value is primitive (string, number, bool):
        return value
    if value is slice/array:
        return [ToValueJSON(elem) for elem in value]
    if value is pointer or map:
        if registered(value):
            return ObjectRef{Obj: lookupID(value)}
        else:
            error: pointers and maps must be registered
```

## Examples

### Array of Registered Objects

```go
type Person struct {
    Name string
}

tracker := changetracker.NewTracker()

alice := &Person{Name: "Alice"}
bob := &Person{Name: "Bob"}

tracker.CreateVariable(alice, 0, nil)  // ID 1
tracker.CreateVariable(bob, 0, nil)    // ID 2

// ToValueJSON for the array
people := []*Person{alice, bob, alice}

// Value JSON result:
// [{"obj": 1}, {"obj": 2}, {"obj": 1}]
```

### Nested Arrays

```go
// Arrays can be nested
matrix := [][]*Person{
    {alice, bob},
    {bob, alice},
}

// Value JSON result:
// [[{"obj": 1}, {"obj": 2}], [{"obj": 2}, {"obj": 1}]]
```

### Registered Map

```go
m := map[string]*Person{"alice": alice, "bob": bob}
tracker.CreateVariable(m, 0, nil)  // ID 3

// ToValueJSON(m) returns:
// {"obj": 3}
```

### Mixed Primitives and References

```go
data := []any{"hello", 42, alice, true, bob}

// Value JSON result:
// ["hello", 42, {"obj": 1}, true, {"obj": 2}]
```

## Decoding Object References

To work with Value JSON that contains object references:

```go
// Check if a value is an object reference
if IsObjectRef(value) {
    id, _ := GetObjectRefID(value)
    obj := tracker.GetObject(id)
    // use obj...
}
```
