# Change
**Source Spec:** main.md, api.md

## Responsibilities

### Knows
- VariableID: int64 - which variable changed
- Priority: Priority - priority level of this change entry
- ValueChanged: bool - whether the value changed
- PropertiesChanged: []string - names of properties that changed at this priority level

### Does
- (Value type with no methods - simple data container)

## Collaborators
- Tracker: created by internal sortChanges method (called by DetectChanges), stored in sortedChanges slice
- Variable: references variables that changed
- Priority: uses Priority type for change priority level

## Sequences
- seq-detect-changes.md: creation and sorting of Change objects as part of DetectChanges

## Notes
- A single variable may produce multiple Change entries if its value and properties have different priorities
- For example: high-priority value change + low-priority property change = 2 Change entries
- Used as return type from Tracker.DetectChanges()
- Stored as flat array (not pointers) in Tracker.sortedChanges for reuse
