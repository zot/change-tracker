# ObjectRef
**Source Spec:** value-json.md, api.md
**Requirements:** R29, R30, R31, R32, R33

## Responsibilities

### Knows
- Obj: int64 - variable ID of the referenced object

### Does
- (struct type - data only, no behavior)

## Collaborators
- Tracker: created by ToValueJSON for registered objects

## Sequences
- seq-to-value-json.md: created during serialization

## Notes

ObjectRef is a simple struct used in Value JSON representation:
```go
type ObjectRef struct {
    Obj int64 `json:"obj"`
}
```

When serialized to JSON: `{"obj": 123}`

### Helper Functions
- IsObjectRef(value): checks if value is an ObjectRef
- GetObjectRefID(value): extracts ID from ObjectRef
