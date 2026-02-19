# Sequence: Detect Changes
**Source Spec:** main.md, api.md

## Participants
- Client: caller triggering change detection
- Tracker: orchestrates change detection via priority-ordered graph traversal
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
  |                    |  [clear all Variable.Error to nil]     |              |
  |                    |                    |                   |              |
  |                    |  === COLLECTION PHASE ===              |              |
  |                    |  [for each rootID in rootIDs]          |              |
  |                    |                    |                   |              |
  |                    | collectActive(id,  |                   |              |
  |                    |   &hi, &med, &lo)  |                   |              |
  |                    |--------.           |                   |              |
  |                    |        |           |                   |              |
  |                    |        |    [if v.Active == false]     |              |
  |                    |        |    skip variable and children |              |
  |                    |        |    return                     |              |
  |                    |        |           |                   |              |
  |                    |        |    [if v.IsReadable()]        |              |
  |                    |        |    append to hi/med/lo        |              |
  |                    |        |    by v.ValuePriority         |              |
  |                    |        |           |                   |              |
  |                    |        |  [for each childID in v.ChildIDs]            |
  |                    |        | collectActive(childID,        |              |
  |                    |        |   &hi, &med, &lo)             |              |
  |                    |        |    (recursive)                |              |
  |                    |        |--------.  |                   |              |
  |                    |        |<-------'  |                   |              |
  |                    |        |           |                   |              |
  |                    |<-------'           |                   |              |
  |                    |                    |                   |              |
  |                    |  === CHECK PHASE (priority order) ===  |              |
  |                    |  [init checked set]                    |              |
  |                    |  [for each group: hi, med, lo]         |              |
  |                    |  [for each id in group]                |              |
  |                    |                    |                   |              |
  |                    | checkWithAncestors |                   |              |
  |                    | (id, checked)      |                   |              |
  |                    |--------.           |                   |              |
  |                    |        |           |                   |              |
  |                    |        |    [if checked[id]: skip]     |              |
  |                    |        |           |                   |              |
  |                    |        |    [if parent readable        |              |
  |                    |        |     && !checked[parentID]]    |              |
  |                    |        | checkWithAncestors            |              |
  |                    |        | (parentID, checked)           |              |
  |                    |        |--------.  |                   |              |
  |                    |        |<-------'  |                   |              |
  |                    |        |           |                   |              |
  |                    |        | checked[id] = true            |              |
  |                    |        |           |                   |              |
  |                    | checkSingle(id)    |                   |              |
  |                    |--------.           |                   |              |
  |                    |        |           |                   |              |
  |                    |        | GetValue()|                   |              |
  |                    |        |---------->|                   |              |
  |                    |        |           |                   |              |
  |                    |        |           |    [if child var] |              |
  |                    |        |           | parent.NavValue() |              |
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
  |                    |        |           | v.ChangeCount++   |              |
  |                    |        |---------->|                   |              |
  |                    |        |           |                   |              |
  |                    |        |           | ValueJSON =       |              |
  |                    |        |           | currentJSON       |              |
  |                    |        |---------->|                   |              |
  |                    |        |           |                   |              |
  |                    |        |           | Value =           |              |
  |                    |        |           | currentValue      |              |
  |                    |        |---------->|                   |              |
  |                    |        |           |                   |              |
  |                    |<-------'           |                   |              |
  |                    |                    |                   |              |
  |                    |  [end for each group]                  |              |
  |                    |                    |                   |              |
  |                    |                    |                   |              |
  |                    |  [if changed]      |                   |              |
  |                    |  t.ChangeCount++   |                   |              |
  |                    |--------.           |                   |              |
  |                    |<-------'           |                   |              |
  |                    |                    |                   |              |
  |     bool           |                    |                   |              |
  |<-------------------|                    |                   |              |
  |                    |                    |                   |              |
  |  GetChanges()      |                    |                   |              |
  |------------------->|                    |                   |              |
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
- **Two-phase approach**: Collection phase walks the tree to gather active variables; check phase processes them in priority order
- Collection phase: Tree traversal from roots respects Active flag (inactive parent prunes entire subtree)
- Collection phase: Non-readable variables (w, action) are skipped but their children are still collected
- Check phase: All high-priority variables checked first, then medium, then low
- **Parent-before-child guarantee**: Before checking any variable, checkWithAncestors walks up the parent chain and checks any unchecked readable ancestors first. This ensures NavigationValue() is fresh. A `checked` set prevents double-processing. Lower-priority parents are pulled forward when needed.
- Comparison uses Value JSON representation (deep equality)
- Both Value and ValueJSON are updated after comparison
- Root variables use their cached Value directly (no path navigation)
- Child variables navigate from parent's cached Value using path
- DetectChanges only marks value changes (not property changes) and returns bool
- Property changes are recorded immediately when SetProperty() is called
- GetChanges() calls sortChanges() to build sorted []Change, then clears internal records
- Returned slice is valid until the next call to GetChanges()

### Access vs Active Behavior
| Condition | Variable Collected | Variable Checked | Children Collected |
|-----------|-------------------|------------------|-------------------|
| Active=true, Access=rw | Yes | Yes | Yes |
| Active=true, Access=r | Yes | Yes | Yes |
| Active=true, Access=w | No | No | Yes |
| Active=true, Access=action | No | No | Yes |
| Active=false | No | No | No |
