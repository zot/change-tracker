# Sequence: Create Variable
**Source Spec:** api.md

## Participants
- Client: caller creating a variable
- Tracker: manages variables and object registry
- Variable: the created variable instance
- Registry: object registry (internal to Tracker)

## Sequence

```
Client              Tracker             Variable            Registry
  |                    |                    |                   |
  |  CreateVariable    |                    |                   |
  |  (value,parentID,  |                    |                   |
  |   path,props)      |                    |                   |
  |------------------->|                    |                   |
  |                    |                    |                   |
  |                    | nextID++           |                   |
  |                    |--------.           |                   |
  |                    |        |           |                   |
  |                    |<-------'           |                   |
  |                    |                    |                   |
  |                    | new Variable       |                   |
  |                    |------------------->|                   |
  |                    |                    |                   |
  |                    |    [if path empty: use props["path"]]  |
  |                    |    [if path has "?": parse query]      |
  |                    | parsePathAndQuery  |                   |
  |                    | (path or props)    |                   |
  |                    |--------.           |                   |
  |                    |        | pathStr,  |                   |
  |                    |        | queryProps|                   |
  |                    |<-------'           |                   |
  |                    |                    |                   |
  |                    |    [merge: props, then queryProps]     |
  |                    | mergeProperties    |                   |
  |                    |--------.           |                   |
  |                    |<-------'           |                   |
  |                    |                    |                   |
  |                    |                    | parsePath(pathStr)|
  |                    |                    |-------.           |
  |                    |                    |       |           |
  |                    |                    |<------'           |
  |                    |                    |                   |
  |                    |    [if has "priority" property]        |
  |                    |                    | ValuePriority =   |
  |                    |                    | parsePriority()   |
  |                    |                    |-------.           |
  |                    |                    |<------'           |
  |                    |                    |                   |
  |                    |    [if root: parentID==0]              |
  |                    |                    | Value = value     |
  |                    |                    |-------.           |
  |                    |                    |<------'           |
  |                    |                    |                   |
  |                    |    [if child: parentID!=0]             |
  |                    | GetVariable        |                   |
  |                    | (parentID)         |                   |
  |                    |--------.           |                   |
  |                    |<-------' parent    |                   |
  |                    |                    |                   |
  |                    |                    | Get() via path    |
  |                    |                    |-------.           |
  |                    |                    |<------' computed  |
  |                    |                    |                   |
  |                    |    [if Value is pointer/map]           |
  |                    | register           |                   |
  |                    | (Value, ID)        |                   |
  |                    |--------------------------------------->|
  |                    |                                        |
  |                    |<---------------------------------------'
  |                    |                    |                   |
  |                    | ToValueJSON        |                   |
  |                    | (Value)            |                   |
  |                    |--------.           |                   |
  |                    |<-------' json      |                   |
  |                    |                    |                   |
  |                    |                    | ValueJSON = json  |
  |                    |------------------->|                   |
  |                    |                    |                   |
  |                    | store in map       |                   |
  |                    |--------.           |                   |
  |                    |<-------'           |                   |
  |                    |                    |                   |
  |       *Variable    |                    |                   |
  |<-------------------|                    |                   |
  |                    |                    |                   |
```

## Notes
- ID assignment is sequential starting from 1
- Path argument: if empty, uses props["path"]; if non-empty, overrides props["path"]
- Path can include URL-style query: "a.b?width=1&height=2"
- Query properties in path override properties map
- Path parsing splits dot-separated string into elements
- Integer strings become int path elements, others remain strings
- The "priority" property sets ValuePriority (low/medium/high)
- Objects (pointers/maps) are automatically registered
- ValueJSON is cached for later change detection
- If props is nil, an empty map is initialized
