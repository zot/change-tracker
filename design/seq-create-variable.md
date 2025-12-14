# Sequence: Create Variable
**Source Spec:** api.md

## Participants
- Client: caller creating a variable
- Tracker: manages variables and object registry
- Variable: the created variable instance
- Parent: parent variable (if child)
- Registry: object registry (internal to Tracker)

## Sequence

```
Client              Tracker             Variable            Parent          Registry
  |                    |                    |                   |               |
  |  CreateVariable    |                    |                   |               |
  |  (value,parentID,  |                    |                   |               |
  |   path,props)      |                    |                   |               |
  |------------------->|                    |                   |               |
  |                    |                    |                   |               |
  |                    | nextID++           |                   |               |
  |                    |--------.           |                   |               |
  |                    |        |           |                   |               |
  |                    |<-------'           |                   |               |
  |                    |                    |                   |               |
  |                    | new Variable       |                   |               |
  |                    | (Active = true)    |                   |               |
  |                    |------------------->|                   |               |
  |                    |                    |                   |               |
  |                    |    [if path empty: use props["path"]]  |               |
  |                    |    [if path has "?": parse query]      |               |
  |                    | parsePathAndQuery  |                   |               |
  |                    | (path or props)    |                   |               |
  |                    |--------.           |                   |               |
  |                    |        | pathStr,  |                   |               |
  |                    |        | queryProps|                   |               |
  |                    |<-------'           |                   |               |
  |                    |                    |                   |               |
  |                    |    [merge: props, then queryProps]     |               |
  |                    | mergeProperties    |                   |               |
  |                    |--------.           |                   |               |
  |                    |<-------'           |                   |               |
  |                    |                    |                   |               |
  |                    |                    | parsePath(pathStr)|               |
  |                    |                    |-------.           |               |
  |                    |                    |       |           |               |
  |                    |                    |<------'           |               |
  |                    |                    |                   |               |
  |                    |    [validate access/path combination]  |               |
  |                    |    [if Access in (r,rw) and path ends in (_)]          |
  |                    |    return error: cannot read from setter               |
  |                    |    [if Access in (w,rw) and path ends in ()]           |
  |                    |    return error: use action for zero-arg methods       |
  |                    |                    |                   |               |
  |                    |    [if has "priority" property]        |               |
  |                    |                    | ValuePriority =   |               |
  |                    |                    | parsePriority()   |               |
  |                    |                    |-------.           |               |
  |                    |                    |<------'           |               |
  |                    |                    |                   |               |
  |                    |    [if root: parentID==0]              |               |
  |                    |                    | Value = value     |               |
  |                    |                    |-------.           |               |
  |                    |                    |<------'           |               |
  |                    |                    |                   |               |
  |                    | rootIDs[ID] = true |                   |               |
  |                    |--------.           |                   |               |
  |                    |<-------'           |                   |               |
  |                    |                    |                   |               |
  |                    |    [if child: parentID!=0]             |               |
  |                    | GetVariable        |                   |               |
  |                    | (parentID)         |                   |               |
  |                    |--------.           |                   |               |
  |                    |<-------' parent    |                   |               |
  |                    |                    |                   |               |
  |                    |    [if Access != "action"]             |               |
  |                    |                    | Get() via path    |               |
  |                    |                    |-------.           |               |
  |                    |                    |<------' computed  |               |
  |                    |                    |                   |               |
  |                    |    [if Access == "action"]             |               |
  |                    |    skip initial value computation      |               |
  |                    |    (avoid premature action invocation) |               |
  |                    |                    |                   |               |
  |                    |                    |                   | ChildIDs +=   |
  |                    |                    |                   | newVar.ID     |
  |                    |---------------------------------->     |               |
  |                    |                    |                   |               |
  |                    |    [if Value is pointer/map]           |               |
  |                    | register           |                   |               |
  |                    | (Value, ID)        |                   |               |
  |                    |-------------------------------------------------->     |
  |                    |                                                        |
  |                    |<--------------------------------------------------'    |
  |                    |                    |                   |               |
  |                    |    [if Access != "action"]             |               |
  |                    | ToValueJSON        |                   |               |
  |                    | (Value)            |                   |               |
  |                    |--------.           |                   |               |
  |                    |<-------' json      |                   |               |
  |                    |                    |                   |               |
  |                    |                    | ValueJSON = json  |               |
  |                    |------------------->|                   |               |
  |                    |                    |                   |               |
  |                    |    [if Access == "action"]             |               |
  |                    |    skip ToValueJSON                    |               |
  |                    |    (no initial value to serialize)     |               |
  |                    |                    |                   |               |
  |                    | store in map       |                   |               |
  |                    |--------.           |                   |               |
  |                    |<-------'           |                   |               |
  |                    |                    |                   |               |
  |       *Variable    |                    |                   |               |
  |<-------------------|                    |                   |               |
  |                    |                    |                   |               |
```

## Notes
- ID assignment is sequential starting from 1
- Active field is set to true by default
- For root variables (parentID == 0): adds variable ID to rootIDs set
- For child variables (parentID != 0): adds variable ID to parent's ChildIDs slice
- Path argument: if empty, uses props["path"]; if non-empty, overrides props["path"]
- Path can include URL-style query: "a.b?width=1&height=2"
- Query properties in path override properties map
- Path parsing splits dot-separated string into elements
- Integer strings become int path elements, others remain strings
- The "priority" property sets ValuePriority (low/medium/high)
- Objects (pointers/maps) are automatically registered
- ValueJSON is cached for later change detection
- If props is nil, an empty map is initialized
- For "action" access: initial value computation is skipped to avoid premature action invocation
- For "action" access: ToValueJSON is also skipped since there is no initial value

### Access/Path Validation
CreateVariable validates that the access mode is compatible with the path:
- `access: "r"` or `access: "rw"` with path ending in `(_)` returns error (cannot read from setter)
- `access: "w"` or `access: "rw"` with path ending in `()` returns error (use `action` for zero-arg methods)
- `rw` is a union of `r` and `w`, so it inherits restrictions from both
- Only `action` access allows both `()` and `(_)` path endings without restriction
