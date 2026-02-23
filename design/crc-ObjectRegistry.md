# ObjectRegistry
**Source Spec:** main.md, api.md, value-json.md
**Requirements:** R25, R26, R27, R28, R29, R34

## Responsibilities

### Knows
- entries: map[uintptr]weakEntry - maps object address to weak reference and ID
- (internal to Tracker - not a separate type)

### Does
- RegisterObject(obj): registers an object and returns (id, true); returns existing ID if already registered; returns (0, false) if not registerable
- UnregisterObject(obj): removes object from registry
- LookupObject(obj): finds ID for registered object
- GetObject(id): retrieves object by ID (returns nil if collected)
- cleanup(): removes entries for garbage-collected objects (automatic via weak refs)

## Collaborators
- Tracker: owns and manages the registry
- ToValueJSON: auto-registers objects during serialization (most common registration path)

## Sequences
- seq-to-value-json.md: auto-registers unregistered pointers/maps/funcs during serialization

## Registration Mechanism

Objects can be registered via:
- **ToValueJSON()**: Automatically registers unregistered objects during serialization (most common path)
- **RegisterObject()**: Explicit registration for cases where an object needs an ID before serialization

## Notes

### Implementation Details
- Uses Go 1.24+ `weak.Pointer` for weak references
- Objects can be garbage collected when no longer referenced by application code
- Registry entries are automatically cleaned up when objects are collected
- Pointers, maps, and funcs can be registered

### Internal Structure
```go
type weakEntry struct {
    weak  weak.Pointer[any]  // weak reference to object
    objID int64              // object ID for ObjectRef serialization
}
```

This is not a standalone type but internal to Tracker. Documented here for design clarity.
