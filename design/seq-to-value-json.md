# Sequence: To Value JSON
**Source Spec:** value-json.md, api.md

## Participants
- Client: caller requesting serialization
- Tracker: performs serialization
- Registry: object registry for lookups and registration

## Sequence

```
Client              Tracker             Registry
  |                    |                    |
  |  ToValueJSON       |                    |
  |  (value)           |                    |
  |------------------->|                    |
  |                    |                    |
  |                    | check type         |
  |                    |--------.           |
  |                    |<-------'           |
  |                    |                    |
  |                    |    [if nil]        |
  |                    | return nil         |
  |                    |-------.            |
  |                    |<------'            |
  |                    |                    |
  |                    |    [if primitive]  |
  |                    |    (string,number, |
  |                    |     bool)          |
  |                    | return value       |
  |                    |-------.            |
  |                    |<------'            |
  |                    |                    |
  |                    |    [if slice/array]|
  |                    |  [for each elem]   |
  |                    | ToValueJSON(elem)  |
  |                    |--------.           |
  |                    |<-------' elemJSON  |
  |                    |  [end for]         |
  |                    | return []elemJSON  |
  |                    |-------.            |
  |                    |<------'            |
  |                    |                    |
  |                    |    [if pointer/map]|
  |                    | LookupObject       |
  |                    | (value)            |
  |                    |------------------->|
  |                    |                    |
  |                    |    [if found]      |
  |                    |      id, true      |
  |                    |<-------------------|
  |                    | return ObjectRef   |
  |                    | {Obj: id}          |
  |                    |-------.            |
  |                    |<------'            |
  |                    |                    |
  |                    |    [if not found]  |
  |                    |      0, false      |
  |                    |<-------------------|
  |                    |                    |
  |                    | allocate nextID    |
  |                    |-------.            |
  |                    |<------'  id        |
  |                    |                    |
  |                    | register (internal)|
  |                    | (value, id)        |
  |                    |------------------->|
  |                    |                    |
  |                    |      registered    |
  |                    |<-------------------|
  |                    |                    |
  |                    | return ObjectRef   |
  |                    | {Obj: id}          |
  |                    |-------.            |
  |                    |<------'            |
  |                    |                    |
  |       json         |                    |
  |<-------------------|                    |
  |                    |                    |
```

## Notes

### Value JSON Types
| Go Type | Value JSON |
|---------|------------|
| nil | nil |
| string | string |
| int, float, etc | number |
| bool | bool |
| slice, array | array (recursive) |
| pointer, map | ObjectRef{Obj: id} |

### Auto-Registration (Only Registration Mechanism)
- ToValueJSON is the **only** way objects get registered - there is no public RegisterObject method
- When an unregistered pointer or map is encountered, it is automatically registered
- The next available ID is allocated and assigned to the object
- This applies to: variable values (during CreateVariable/DetectChanges), wrapper objects, and nested objects in arrays
- After auto-registration, the object can be looked up via LookupObject

### Object Reference Format
```go
type ObjectRef struct {
    Obj int64 `json:"obj"`
}
```
JSON: `{"obj": 123}`
