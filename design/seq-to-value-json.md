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
  |                    | ConvertToValueJSON |
  |                    | (value) via        |
  |                    | Resolver           |
  |                    |--------.           |
  |                    |<-------'           |
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
  |                    | [if ptr/map/func]  |
  |                    | RegisterObject     |
  |                    | (value)            |
  |                    |  [if registered]   |
  |                    | return ObjectRef   |
  |                    |-------.            |
  |                    |<------'            |
  |                    |                    |
  |                    |    [if struct]     |
  |                    | return nil         |
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
  |                    |    [if primitive]  |
  |                    |    (string,number, |
  |                    |     bool)          |
  |                    | return value       |
  |                    |-------.            |
  |                    |<------'            |
  |                    |                    |
  |       json         |                    |
  |<-------------------|                    |
  |                    |                    |
```

## Notes

### Processing Order
1. Resolver.ConvertToValueJSON() — allows custom resolvers to transform domain-specific types
2. nil check — returns nil
3. RegisterObject attempt — pointers, maps, and funcs become ObjectRef{Obj: id}
4. Struct check — returns nil (structs are not serializable as Value JSON)
5. Slice/array — recursive ToValueJSON on each element
6. Fallthrough — primitives (string, number, bool) returned as-is

### Value JSON Types
| Go Type | Value JSON |
|---------|------------|
| nil | nil |
| string | string |
| int, float, etc | number |
| bool | bool |
| slice, array | array (recursive) |
| struct | nil |
| pointer, map, func | ObjectRef{Obj: id} |

### Registration
- ToValueJSON auto-registers unregistered pointers, maps, and funcs via RegisterObject
- RegisterObject is also available as a public method for explicit registration
- Auto-registration applies to: variable values (during CreateVariable/DetectChanges), wrapper objects, and nested objects in arrays
- After registration, the object can be looked up via LookupObject

### Object Reference Format
```go
type ObjectRef struct {
    Obj int64 `json:"obj"`
}
```
JSON: `{"obj": 123}`
