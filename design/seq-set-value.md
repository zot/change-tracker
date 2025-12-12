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
  |                    |                    | Set(val,          |
  |                    |                    | lastElem,         |
  |                    |                    | newValue)         |
  |                    |----------------------------------->|
  |                    |                        nil/error   |
  |                    |<-----------------------------------|
  |                    |                    |                   |
  |       nil/error    |                    |                   |
  |<-------------------|                    |                   |
  |                    |                    |                   |
```

## Notes
- Cannot set root variables directly (they hold external values)
- Navigates to parent of target using all but last path element
- Uses resolver's Set to assign value at final path element
- Value cache is NOT updated (will update on next Get or DetectChanges)
- Struct field setting requires pointer to struct
- Slice index must be within bounds
- Map keys can be set freely
