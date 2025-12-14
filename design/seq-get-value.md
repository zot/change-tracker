# Sequence: Get Value
**Source Spec:** api.md, resolver.md

## Participants
- Client: caller getting a value
- Variable: the variable being queried
- Tracker: provides resolver and parent lookup
- Resolver: navigates path elements

## Sequence

```
Client              Variable            Tracker             Resolver
  |                    |                    |                   |
  |  Get()             |                    |                   |
  |------------------->|                    |                   |
  |                    |                    |                   |
  |                    |    [if Access == "w" or "action"]      |
  |                    | return error       |                   |
  |                    | (not readable)     |                   |
  |                    |-------.            |                   |
  |                    |<------'            |                   |
  |                    |                    |                   |
  |                    |    [if path ends with "(_)"]           |
  |                    | return error       |                   |
  |                    | (write-only path)  |                   |
  |                    |-------.            |                   |
  |                    |<------'            |                   |
  |                    |                    |                   |
  |                    |    [if root: ParentID==0]              |
  |                    | return Value       |                   |
  |                    |-------.            |                   |
  |                    |<------'            |                   |
  |                    |                    |                   |
  |                    |    [if child: ParentID!=0]             |
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
  |                    |  [for each elem in Path]               |
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
  |                    |                    |                   |
  |                    |  [end for each]    |                   |
  |                    |                    |                   |
  |                    | Value = val        |                   |
  |                    |--------.           |                   |
  |                    |<-------' (cache)   |                   |
  |                    |                    |                   |
  |       value, nil   |                    |                   |
  |<-------------------|                    |                   |
  |                    |                    |                   |
```

## Notes
- Access check is first: `access: "w"` (write-only) or `access: "action"` returns error immediately
- Root variables return their cached Value directly
- Child variables navigate from parent's cached Value
- Path ending in `(_)` is write-only; Get returns error
- Path elements ending in `()` use Call for zero-arg method invocation
- Other path elements use Get for field/key/index access
- The result is cached in Variable.Value for child navigation
- Errors propagate if any path element resolution fails
- Path elements can be strings (field, key, method) or ints (index)
- Access property is independent of path semantics (both can restrict Get)
