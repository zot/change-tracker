# Wrapper Specification

Wrappers allow a custom resolver to provide an alternative object for child variable navigation. When a variable has a wrapper, child variables navigate through the wrapper instead of the original value.

## Wrapper Property

Setting the `wrapper` property on a variable enables wrapper creation:

```go
// Enable wrapper via path query
v := tracker.CreateVariable(data, 0, "?wrapper=true", nil)

// Or via properties map
v := tracker.CreateVariable(data, 0, "", map[string]string{"wrapper": "true"})

// Or via SetProperty
v.SetProperty("wrapper", "true")
```

The wrapper property value can be any non-empty string. The actual value is passed to `CreateWrapper` via the variable's properties, allowing different wrapper configurations.

## CreateWrapper Method

The `Resolver` interface includes `CreateWrapper`:

```go
type Resolver interface {
    // ... existing methods ...

    // CreateWrapper creates a wrapper object for the given variable.
    // Called when the variable has a "wrapper" property and ValueJSON is non-nil.
    // Returns:
    //   - A new wrapper object
    //   - The same wrapper object (v.WrapperValue) to preserve state
    //   - nil if no wrapper is needed
    CreateWrapper(variable *Variable) any
}
```

The default `Tracker` implementation returns `nil` (no wrapper). Custom resolvers implement this method to create wrappers.

## Wrapper Lifecycle

### Creation

A wrapper is created when:
1. The variable has a non-empty `wrapper` property, AND
2. The variable's `ValueJSON` is non-nil

When these conditions are met, `Resolver.CreateWrapper(variable)` is called.

### Update on Value Change

When `ValueJSON` changes (during `DetectChanges` or when the wrapper property changes):

1. `CreateWrapper(v)` is called
2. If the returned wrapper is the **same pointer** as `v.WrapperValue`:
   - No action taken (wrapper preserved with its state)
3. If the returned wrapper is **different** (including nil):
   - Old wrapper is unregistered from the object registry
   - New wrapper is registered (if non-nil)
   - `WrapperValue` is updated
   - `WrapperJSON` is recomputed

This allows `CreateWrapper` to:
- Return the same wrapper object (modified in place) to preserve persistent state
- Return a new wrapper when replacement is needed
- Return `nil` to remove the wrapper

### Destruction

A wrapper is destroyed (unregistered and cleared) when:
- The `wrapper` property is set to empty string
- The variable's `ValueJSON` becomes nil
- The variable is destroyed via `DestroyVariable`

## Child Navigation

Child variables use `NavigationValue()` to get the starting point for path resolution:

```go
func (v *Variable) NavigationValue() any {
    if v.WrapperValue != nil {
        return v.WrapperValue
    }
    return v.Value
}
```

Both `Get()` and `Set()` operations on child variables navigate from the parent's `NavigationValue()`.

## Use Cases

### Provide Different Interface

Expose a different set of fields/methods to children than the underlying value:

```go
type DataWrapper struct {
    ComputedField string
    FormattedDate string
}

func (r *MyResolver) CreateWrapper(v *Variable) any {
    data := v.Value.(*MyData)
    return &DataWrapper{
        ComputedField: computeField(data),
        FormattedDate: data.Date.Format("2006-01-02"),
    }
}
```

### Maintain Persistent State

Preserve wrapper state across value changes by returning the same wrapper:

```go
type StatefulWrapper struct {
    cache map[string]any
    data  *MyData
}

func (r *MyResolver) CreateWrapper(v *Variable) any {
    // Reuse existing wrapper if present
    if w, ok := v.WrapperValue.(*StatefulWrapper); ok {
        w.data = v.Value.(*MyData)  // Update data reference
        return w                     // Same pointer preserves cache
    }
    // Create new wrapper
    return &StatefulWrapper{
        cache: make(map[string]any),
        data:  v.Value.(*MyData),
    }
}
```

### Implement Adapters

Adapt an object to a different interface expected by child variables:

```go
type LegacyAdapter struct {
    modern *ModernAPI
}

func (a *LegacyAdapter) GetName() string {
    return a.modern.FullName()
}

func (r *MyResolver) CreateWrapper(v *Variable) any {
    return &LegacyAdapter{modern: v.Value.(*ModernAPI)}
}
```

## Example: Custom Resolver with Wrappers

```go
type WrapperData struct {
    WrappedName string
    Extra       int
}

type myResolver struct {
    *changetracker.Tracker
}

func (r *myResolver) CreateWrapper(v *changetracker.Variable) any {
    if v.Value == nil {
        return nil
    }
    if p, ok := v.Value.(*Person); ok {
        // Check for existing wrapper to preserve state
        if w, ok := v.WrapperValue.(*WrapperData); ok {
            w.WrappedName = "Wrapped:" + p.Name
            return w  // Same pointer, state preserved
        }
        return &WrapperData{
            WrappedName: "Wrapped:" + p.Name,
            Extra:       42,
        }
    }
    return nil
}

// Usage
tr := changetracker.NewTracker()
tr.Resolver = &myResolver{tr}

person := &Person{Name: "Alice"}
parent := tr.CreateVariable(person, 0, "?wrapper=true", nil)

// Children navigate through WrapperData, not Person
child := tr.CreateVariable(nil, parent.ID, "WrappedName", nil)
val, _ := child.Get()  // Returns "Wrapped:Alice"

extraChild := tr.CreateVariable(nil, parent.ID, "Extra", nil)
extra, _ := extraChild.Get()  // Returns 42
```

## Variable Fields

Wrapper support adds two fields to `Variable`:

| Field | Type | Description |
|-------|------|-------------|
| `WrapperValue` | `any` | The wrapper object (nil if no wrapper) |
| `WrapperJSON` | `any` | Serialized form of WrapperValue via `ToValueJSON` |

## Registration

Wrappers are registered in the object registry automatically via `ToValueJSON()`:
- When `WrapperJSON` is computed, `ToValueJSON()` auto-registers the wrapper with a unique ID
- Each wrapper gets its own ID (not the variable's ID)
- Unregistered when replaced or destroyed
- Can be looked up via `LookupObject`

See value-json.md for details on how automatic registration works.
