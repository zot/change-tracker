# Sequence: Set Value
**Source Spec:** api.md, resolver.md

## Participants
- Client: caller setting a value
- Variable: the variable being modified
- Tracker: provides resolver and parent lookup
- Resolver: navigates and sets value

## Sequence

```
Client              Variable            Tracker             Resolver
  |                    |                    |                   |
  |  Set(newValue)     |                    |                   |
  |------------------->|                    |                   |
  |                    |                    |                   |
  |                    |    [if Access == "r"]                  |
  |                    | error: read-only   |                   |
  |                    | variable           |                   |
  |                    |-------.            |                   |
  |                    |<------'            |                   |
  |                    |                    |                   |
  |                    |    [if Access != "w" AND path ends with "()"]  |
  |                    | error: read-only   |                   |
  |                    | path (getter)      |                   |
  |                    |-------.            |                   |
  |                    |<------'            |                   |
  |                    |                    |                   |
  |                    |    [if "(_)" not at end of path]       |
  |                    | error: setter must |                   |
  |                    | be terminal        |                   |
  |                    |-------.            |                   |
  |                    |<------'            |                   |
  |                    |                    |                   |
  |                    |    [if root: ParentID==0]              |
  |                    | error: cannot set  |                   |
  |                    | root directly      |                   |
  |                    |-------.            |                   |
  |                    |<------'            |                   |
  |                    |                    |                   |
  |                    |    [if no Path]    |                   |
  |                    | error: no path     |                   |
  |                    |-------.            |                   |
  |                    |<------'            |                   |
  |                    |                    |                   |
  |                    | GetVariable        |                   |
  |                    | (ParentID)         |                   |
  |                    |------------------->|                   |
  |                    |      parent        |                   |
  |                    |<-------------------|                   |
  |                    |                    |                   |
  |                    | val = parent.Value |                   |
  |                    |--------.           |                   |
  |                    |<-------'           |                   |
  |                    |                    |                   |
  |                    |  [for each elem in Path[:-1]]          |
  |                    |  (all but last)    |                   |
  |                    |                    |                   |
  |                    |    [if elem ends with "()"]            |
  |                    |                    | Call(val,         |
  |                    |                    |   methodName)     |
  |                    |----------------------------------->|
  |                    |                        nextVal     |
  |                    |<-----------------------------------|
  |                    |                    |                   |
  |                    |    [else: field/key/index]             |
  |                    |                    | Get(val, elem)    |
  |                    |----------------------------------->|
  |                    |                        nextVal     |
  |                    |<-----------------------------------|
  |                    |                    |                   |
  |                    | val = nextVal      |                   |
  |                    |--------.           |                   |
  |                    |<-------'           |                   |
  |                    |  [end for each]    |                   |
  |                    |                    |                   |
  |                    | lastElem =         |                   |
  |                    | Path[len-1]        |                   |
  |                    |--------.           |                   |
  |                    |<-------'           |                   |
  |                    |                    |                   |
  |                    |    [if lastElem ends with "(_)"]       |
  |                    |                    | CallWith(val,     |
  |                    |                    |   methodName,     |
  |                    |                    |   newValue)       |
  |                    |----------------------------------->|
  |                    |                        nil/error   |
  |                    |<-----------------------------------|
  |                    |                    |                   |
  |                    |    [else if (Access == "w" or "action") AND lastElem ends with "()"]
  |                    |                    | Call(val,         |
  |                    |                    |   methodName)     |
  |                    |                    | (side effect)     |
  |                    |----------------------------------->|
  |                    |                        val/error   |
  |                    |<-----------------------------------|
  |                    |                    |                   |
  |                    |    [else: field/key/index]             |
  |                    |                    | Set(val,          |
  |                    |                    |   lastElem,       |
  |                    |                    |   newValue)       |
  |                    |----------------------------------->|
  |                    |                        nil/error   |
  |                    |<-----------------------------------|
  |                    |                    |                   |
  |       nil/error    |                    |                   |
  |<-------------------|                    |                   |
  |                    |                    |                   |
```

## Notes
- Access check is first: `access: "r"` (read-only) returns error immediately
- Cannot set root variables directly (they hold external values)
- Path ending in `()` is read-only for readable variables (access "r" or "rw"); Set returns error
- For write-only or action variables (`access: "w"` or `access: "action"`), path ending in `()` allows Set to call the method for side effects
- Path with `(_)` not at end returns error (setter must be terminal)
- Navigates to parent of target using all but last path element
- Path elements ending in `()` use Call for navigation
- If last element ends in `(_)`, uses CallWith to invoke setter method
- Otherwise uses resolver's Set to assign value at final path element
- Value cache is NOT updated (will update on next Get or DetectChanges)
- Struct field setting requires pointer to struct
- Slice index must be within bounds
- Map keys can be set freely
- Access property is independent of path semantics (both can restrict Set)
