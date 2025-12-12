# Sequence: Detect Changes
**Source Spec:** main.md, api.md

## Participants
- Client: caller triggering change detection
- Tracker: orchestrates change detection
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
  |                    |  [for each variable in tracker]        |              |
  |                    |                    |                   |              |
  |                    | Get()              |                   |              |
  |                    |------------------->|                   |              |
  |                    |                    |                   |              |
  |                    |                    |    [if child var] |              |
  |                    |                    | parent.Value      |              |
  |                    |                    |-------.           |              |
  |                    |                    |<------'           |              |
  |                    |                    |                   |              |
  |                    |                    |  [for each path element]         |
  |                    |                    | Get(val, elem)    |              |
  |                    |                    |------------------>|              |
  |                    |                    |      value        |              |
  |                    |                    |<------------------|              |
  |                    |                    |                   |              |
  |                    |      currentValue  |                   |              |
  |                    |<-------------------|                   |              |
  |                    |                    |                   |              |
  |                    | ToValueJSON        |                   |              |
  |                    | (currentValue)     |                   |              |
  |                    |--------.           |                   |              |
  |                    |<-------' currentJSON                   |              |
  |                    |                    |                   |              |
  |                    | compare            |                   |              |
  |                    | currentJSON vs     |                   |              |
  |                    | Variable.ValueJSON |                   |              |
  |                    |--------.           |                   |              |
  |                    |<-------'           |                   |              |
  |                    |                    |                   |              |
  |                    |    [if different]  |                   |              |
  |                    | valueChanges[ID]   |                   |              |
  |                    | = true             |                   |              |
  |                    |--------.           |                   |              |
  |                    |<-------'           |                   |              |
  |                    |                    |                   |              |
  |                    |                    | ValueJSON =       |              |
  |                    |                    | currentJSON       |              |
  |                    |------------------->|                   |              |
  |                    |                    |                   |              |
  |                    |                    | Value =           |              |
  |                    |                    | currentValue      |              |
  |                    |------------------->|                   |              |
  |                    |                    |                   |              |
  |                    |  [end for each]    |                   |              |
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
- All variables are checked, not just root variables
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
