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
  |                    |                    | Get(val, elem)    |
  |                    |----------------------------------->|
  |                    |                                    |
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
- Root variables return their cached Value directly
- Child variables navigate from parent's cached Value
- Each path element is resolved via tracker's Resolver
- The result is cached in Variable.Value for child navigation
- Errors propagate if any path element resolution fails
- Path elements can be strings (field, key, method) or ints (index)
