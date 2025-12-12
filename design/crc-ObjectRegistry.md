# ObjectRegistry
**Source Spec:** main.md, api.md

## Responsibilities

### Knows
- entries: map[uintptr]weakEntry - maps object address to weak reference and variable ID
- (internal to Tracker - not a separate type)

### Does
- register(obj, varID): stores weak reference to object with associated variable ID
- unregister(obj): removes object from registry
- lookup(obj): finds variable ID for object
- getObject(varID): retrieves object by variable ID (returns nil if collected)
- cleanup(): removes entries for garbage-collected objects (automatic via weak refs)

## Collaborators
- Tracker: owns and manages the registry
- Variable: objects are registered when variables are created

## Sequences
- seq-create-variable.md: objects registered during variable creation
- seq-to-value-json.md: registry consulted during serialization

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
    varID int64              // associated variable ID
}
```

This is not a standalone type but internal to Tracker. Documented here for design clarity.
