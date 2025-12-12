# Sequence: To Value JSON
**Source Spec:** value-json.md, api.md

## Participants
- Client: caller requesting serialization
- Tracker: performs serialization
- Registry: object registry for lookups

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
  |                    | error: unregistered|
  |                    | pointer/map        |
  |                    |-------.            |
  |                    |<------'            |
  |                    |                    |
  |       json/error   |                    |
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

### Error Conditions
- Unregistered pointer causes error
- Unregistered map causes error
- All pointers/maps must be registered before serialization

### Object Reference Format
```go
type ObjectRef struct {
    Obj int64 `json:"obj"`
}
```
JSON: `{"obj": 123}`
