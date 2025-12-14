# Sequence: Detect Changes
**Source Spec:** main.md, api.md

## Participants
- Client: caller triggering change detection
- Tracker: orchestrates change detection via tree traversal
- Variable: each tracked variable
- Resolver: navigates to current values
- Change: change record data structure

## Sequence

```
Client              Tracker             Variable            Resolver        Change
  |                    |                    |                   |              |
  |  DetectChanges()   |                    |                   |              |
  |------------------->|                    |                   |              |
  |                    |                    |                   |              |
  |                    |  [for each rootID in rootIDs]          |              |
  |                    |                    |                   |              |
  |                    | checkVariable(id)  |                   |              |
  |                    | (recursive DFS)    |                   |              |
  |                    |--------.           |                   |              |
  |                    |        |           |                   |              |
  |                    |        |    [if v.Active == false]     |              |
  |                    |        |    skip variable and children |              |
  |                    |        |    return                     |              |
  |                    |        |           |                   |              |
  |                    |        |    [if v.Access == "w" or "action"]          |
  |                    |        |    skip variable (not readable)              |
  |                    |        |    continue to children       |              |
  |                    |        |           |                   |              |
  |                    |        |    [if v.Active == true && v.Access is "r" or "rw"]
  |                    |        |           |                   |              |
  |                    |        | Get()     |                   |              |
  |                    |        |---------->|                   |              |
  |                    |        |           |                   |              |
  |                    |        |           |    [if child var] |              |
  |                    |        |           | parent.Value      |              |
  |                    |        |           |-------.           |              |
  |                    |        |           |<------'           |              |
  |                    |        |           |                   |              |
  |                    |        |           |  [for each path element]         |
  |                    |        |           | Get(val, elem)    |              |
  |                    |        |           |------------------>|              |
  |                    |        |           |      value        |              |
  |                    |        |           |<------------------|              |
  |                    |        |           |                   |              |
  |                    |        | currentValue                  |              |
  |                    |        |<----------|                   |              |
  |                    |        |           |                   |              |
  |                    |        | ToValueJSON                   |              |
  |                    |        | (currentValue)                |              |
  |                    |        |--------.  |                   |              |
  |                    |        |<------'   |                   |              |
  |                    |        | currentJSON                   |              |
  |                    |        |           |                   |              |
  |                    |        | compare   |                   |              |
  |                    |        | currentJSON vs                |              |
  |                    |        | v.ValueJSON                   |              |
  |                    |        |--------.  |                   |              |
  |                    |        |<-------'  |                   |              |
  |                    |        |           |                   |              |
  |                    |        |    [if different]             |              |
  |                    |        | valueChanges[ID]              |              |
  |                    |        | = true    |                   |              |
  |                    |        |--------.  |                   |              |
  |                    |        |<-------'  |                   |              |
  |                    |        |           |                   |              |
  |                    |        |           | ValueJSON =       |              |
  |                    |        |           | currentJSON       |              |
  |                    |        |---------->|                   |              |
  |                    |        |           |                   |              |
  |                    |        |           | Value =           |              |
  |                    |        |           | currentValue      |              |
  |                    |        |---------->|                   |              |
  |                    |        |           |                   |              |
  |                    |        |  [for each childID in v.ChildIDs]            |
  |                    |        | checkVariable(childID)        |              |
  |                    |        | (recursive)                   |              |
  |                    |        |--------.  |                   |              |
  |                    |        |<-------'  |                   |              |
  |                    |        |           |                   |              |
  |                    |<-------'           |                   |              |
  |                    |                    |                   |              |
  |                    |  [end for each root]                   |              |
  |                    |                    |                   |              |
  |                    | sortChanges()      |                   |              |
  |                    | (internal)         |                   |              |
  |                    |--------.           |                   |              |
  |                    |        |  [build Change entries by priority]          |
  |                    |        |------------------------------------------>   |
  |                    |<-------' []Change  |                   |              |
  |                    |                    |                   |              |
  |                    | clear              |                   |              |
  |                    | valueChanges       |                   |              |
  |                    | propertyChanges    |                   |              |
  |                    |--------.           |                   |              |
  |                    |<-------'           |                   |              |
  |                    |                    |                   |              |
  |     []Change       |                    |                   |              |
  |<-------------------|                    |                   |              |
  |                    |                    |                   |              |
```

## Notes
- Tree traversal: DetectChanges iterates over root variables and performs depth-first traversal
- Active check: If a variable's Active field is false, it and all its descendants are skipped
- Access check: If a variable's Access is "w" (write-only) or "action", the variable is skipped but children are still processed
- Root variables are tracked in rootIDs set for efficient iteration
- Child variables are found via parent's ChildIDs slice
- Comparison uses Value JSON representation (deep equality)
- Both Value and ValueJSON are updated after comparison
- Root variables use their cached Value directly (no path navigation)
- Child variables navigate from parent's cached Value using path
- DetectChanges only marks value changes (not property changes)
- Property changes are recorded immediately when SetProperty() is called
- After detection, sortChanges() is called internally to build sorted []Change
- Internal change records (valueChanges, propertyChanges) are cleared after sorting
- The sorted changes slice is preserved and returned
- Returned slice is valid until the next call to DetectChanges()

### Access vs Active Behavior
| Condition | Variable Scanned | Children Scanned |
|-----------|------------------|------------------|
| Active=true, Access=rw | Yes | Yes |
| Active=true, Access=r | Yes | Yes |
| Active=true, Access=w | No | Yes |
| Active=true, Access=action | No | Yes |
| Active=false | No | No |
