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

Registered objects (pointers, maps, and funcs) serialize as:

```json
{"obj": 123}
```

Where `123` is the variable ID associated with the object.

### Structs

Structs serialize as `nil`. They are not directly serializable as Value JSON — use a pointer to the struct instead, which will be registered and serialized as an object reference.

## Registration Rules

Objects can be registered via:

- **`ToValueJSON()`**: Automatically registers unregistered pointers, maps, and funcs during serialization
- **`RegisterObject()`**: Explicit registration for cases where an object needs an ID before serialization

Pointers, maps, and funcs can be registered. Registration happens when `ValueJSON` or `WrapperJSON` is computed.

This registration ensures that:
- Variable values are registered when their ValueJSON is computed during CreateVariable or DetectChanges
- Wrapper objects are registered when their WrapperJSON is computed
- Nested objects in arrays are properly converted to object references

## Serialization Algorithm

```
ToValueJSON(value):
    value = Resolver.ConvertToValueJSON(tracker, value)
    if value is nil:
        return nil
    if value is pointer, map, or func:
        id, ok = RegisterObject(value)
        if ok: return ObjectRef{Obj: id}
    if value is struct:
        return nil
    if value is slice/array:
        return [ToValueJSON(elem) for elem in value]
    return value  // primitives pass through
```

Auto-registration assigns the next available ID to unregistered objects. This enables serialization of arrays containing objects that weren't explicitly registered.

Nested arrays are not supported — `ToValueJSON` panics if an array element is itself an array.

## Examples

### Array of Registered Objects

```go
type Person struct {
    Name string
}

tracker := changetracker.NewTracker()

alice := &Person{Name: "Alice"}
bob := &Person{Name: "Bob"}

tracker.CreateVariable(alice, 0, "", nil)  // ID 1
tracker.CreateVariable(bob, 0, "", nil)    // ID 2

// ToValueJSON for the array
people := []*Person{alice, bob, alice}

// Value JSON result:
// [{"obj": 1}, {"obj": 2}, {"obj": 1}]
```

### Auto-Registration of Unregistered Objects

```go
tracker := changetracker.NewTracker()

// Objects not explicitly registered
alice := &Person{Name: "Alice"}
bob := &Person{Name: "Bob"}

// ToValueJSON auto-registers them
people := []*Person{alice, bob}
json := tracker.ToValueJSON(people)

// Value JSON result (IDs assigned automatically):
// [{"obj": 1}, {"obj": 2}]

// Objects are now registered and can be looked up
id, _ := tracker.LookupObject(alice)  // returns 1, true
```

### Nested Arrays

Nested arrays are **not supported** — `ToValueJSON` panics if an array element is itself an array (e.g., `[][]any`). Arrays should contain only primitives and object references.

### Registered Map

```go
m := map[string]*Person{"alice": alice, "bob": bob}
tracker.CreateVariable(m, 0, "", nil)  // ID 3

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
