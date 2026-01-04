# ObjectRegistry
**Source Spec:** main.md, api.md, value-json.md

## Responsibilities

### Knows
- entries: map[uintptr]weakEntry - maps object address to weak reference and ID
- (internal to Tracker - not a separate type)

### Does
- register(obj, id): (internal) stores weak reference to object with associated ID
- unregister(obj): removes object from registry
- lookup(obj): finds ID for object
- getObject(id): retrieves object by ID (returns nil if collected)
- cleanup(): removes entries for garbage-collected objects (automatic via weak refs)

## Collaborators
- Tracker: owns and manages the registry; ToValueJSON performs registration
- ToValueJSON: the only mechanism that registers objects (automatic during serialization)

## Sequences
- seq-to-value-json.md: auto-registers unregistered pointers/maps during serialization

## Registration Mechanism

Objects are registered **only** via `ToValueJSON()`:
- When ToValueJSON encounters an unregistered pointer or map, it allocates the next available ID and registers the object
- There is no public RegisterObject method - registration is internal only
- This applies to: variable values (during CreateVariable/DetectChanges), wrapper objects, and nested objects in arrays

## Notes

### Implementation Details
- Uses Go 1.24+ `weak.Pointer` for weak references
- Objects can be garbage collected when no longer referenced by application code
- Registry entries are automatically cleaned up when objects are collected
- Only pointers and maps can be registered

### Internal Structure
```go
type weakEntry struct {
    weak  weak.Pointer[any]  // weak reference to object
    objID int64              // object ID for ObjectRef serialization
}
```

This is not a standalone type but internal to Tracker. Documented here for design clarity.
