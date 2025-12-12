# Sequence: Destroy Variable
**Source Spec:** api.md

## Participants
- Client: caller destroying a variable
- Tracker: manages variables and object registry
- Variable: the variable being destroyed
- Parent: parent variable (if child)
- Registry: object registry (internal to Tracker)

## Sequence

```
Client              Tracker             Variable            Parent          Registry
  |                    |                    |                   |               |
  |  DestroyVariable   |                    |                   |               |
  |  (id)              |                    |                   |               |
  |------------------->|                    |                   |               |
  |                    |                    |                   |               |
  |                    | GetVariable(id)    |                   |               |
  |                    |--------.           |                   |               |
  |                    |<-------' v         |                   |               |
  |                    |                    |                   |               |
  |                    |    [if v == nil: return]               |               |
  |                    |                    |                   |               |
  |                    |    [if root: v.ParentID == 0]          |               |
  |                    | delete             |                   |               |
  |                    | rootIDs[id]        |                   |               |
  |                    |--------.           |                   |               |
  |                    |<-------'           |                   |               |
  |                    |                    |                   |               |
  |                    |    [if child: v.ParentID != 0]         |               |
  |                    | GetVariable        |                   |               |
  |                    | (v.ParentID)       |                   |               |
  |                    |--------.           |                   |               |
  |                    |<-------' parent    |                   |               |
  |                    |                    |                   |               |
  |                    |                    |                   | ChildIDs -=   |
  |                    |                    |                   | id            |
  |                    |---------------------------------->     |               |
  |                    |                    |                   |               |
  |                    |    [if Value is pointer/map]           |               |
  |                    | unregister         |                   |               |
  |                    | (v.Value)          |                   |               |
  |                    |-------------------------------------------------->     |
  |                    |                                                        |
  |                    |<--------------------------------------------------'    |
  |                    |                    |                   |               |
  |                    | delete             |                   |               |
  |                    | valueChanges[id]   |                   |               |
  |                    |--------.           |                   |               |
  |                    |<-------'           |                   |               |
  |                    |                    |                   |               |
  |                    | delete             |                   |               |
  |                    | propertyChanges[id]|                   |               |
  |                    |--------.           |                   |               |
  |                    |<-------'           |                   |               |
  |                    |                    |                   |               |
  |                    | delete             |                   |               |
  |                    | variables[id]      |                   |               |
  |                    |--------.           |                   |               |
  |                    |<-------'           |                   |               |
  |                    |                    |                   |               |
  |<-------------------|                    |                   |               |
  |                    |                    |                   |               |
```

## Notes
- If variable ID doesn't exist, the method returns early (no-op)
- For root variables (ParentID == 0): removes variable ID from rootIDs set
- For child variables (ParentID != 0): removes variable ID from parent's ChildIDs slice
- Unregisters the object from the object registry (if it was a pointer/map)
- Removes the variable from the change tracking sets (valueChanges, propertyChanges)
- Removes the variable from the variables map
- Does NOT automatically destroy child variables - caller is responsible for that
