# Sequence: Create Variable
**Source Spec:** api.md

## Participants
- Client: caller creating a variable
- Tracker: manages variables and object registry
- Variable: the created variable instance
- Parent: parent variable (if child)
- Registry: object registry (internal to Tracker)

## Sequence

### CreateVariable (auto-assigned ID)
```
Client              Tracker             Variable
  |                    |                    |
  |  CreateVariable    |                    |
  |  (value,parentID,  |                    |
  |   path,props)      |                    |
  |------------------->|                    |
  |                    |                    |
  |                    | id = nextID        |
  |                    | nextID++           |
  |                    |--------.           |
  |                    |<-------'           |
  |                    |                    |
  |                    | CreateVariableWithId(id, ...)
  |                    |--------.           |
  |                    |        | (see below)
  |                    |<-------' *Variable |
  |                    |                    |
  |       *Variable    |                    |
  |<-------------------|                    |
  |                    |                    |
```

### CreateVariableWithId (caller-specified ID)
```
Client              Tracker             Variable            Parent          Registry
  |                    |                    |                   |               |
  |CreateVariableWithId|                    |                   |               |
  |  (id,value,parentID|                    |                   |               |
  |   path,props)      |                    |                   |               |
  |------------------->|                    |                   |               |
  |                    |                    |                   |               |
  |                    |  [if id in use]    |                   |               |
  |       nil          |  return nil        |                   |               |
  |<-------------------|                    |                   |               |
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
  |                    |    [if Access != "action"]             |               |
  |                    | ToValueJSON        |                   |               |
  |                    | (Value)            |                   |               |
  |                    |--------.           |                   |               |
  |                    |        | (auto-registers              |               |
  |                    |        |  pointers/maps)              |               |
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
- CreateVariable increments nextID, then delegates to CreateVariableWithId
- CreateVariableWithId returns nil if the ID is already in use
- CreateVariableWithId does NOT modify nextID - callers vending their own IDs manage collision avoidance
- ID assignment is sequential starting from 1 (for auto-assigned IDs)
- Active field is set to true by default
- For root variables (parentID == 0): adds variable ID to rootIDs set
- For child variables (parentID != 0): adds variable ID to parent's ChildIDs slice
- Path argument: if empty, uses props["path"]; if non-empty, overrides props["path"]
- Path can include URL-style query: "a.b?width=1&height=2"
- Query properties in path override properties map
- Path parsing splits dot-separated string into elements
- Integer strings become int path elements, others remain strings
- The "priority" property sets ValuePriority (low/medium/high)
- Objects (pointers/maps) are automatically registered via ToValueJSON (see seq-to-value-json.md)
- ValueJSON is cached for later change detection
- If props is nil, an empty map is initialized
- For "action" access: initial value computation is skipped to avoid premature action invocation
- For "action" access: ToValueJSON is also skipped since there is no initial value

### Access/Path Validation
CreateVariable validates that the access mode is compatible with the path:
- `access: "r"` or `access: "rw"` with path ending in `(_)` returns error (cannot read from setter)
- `access: "w"` with path ending in `()` returns error (use `rw`, `r`, or `action` for zero-arg methods)
- Paths ending in `()` are allowed with `rw`, `r`, or `action` access (supports variadic method calls)
- With `rw` access and `()` path: Get() calls method with no args, Set() calls method with args
