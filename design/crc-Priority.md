# Priority
**Source Spec:** main.md, api.md
**Requirements:** R21, R22, R23, R24, R71

## Responsibilities

### Knows
- PriorityLow: -1 - low priority level
- PriorityMedium: 0 - medium priority level (default)
- PriorityHigh: 1 - high priority level

### Does
- ParsePriority(s string): converts a string ("low", "medium", "high") to a Priority value; defaults to PriorityMedium for unrecognized strings

## Collaborators
- Variable: uses Priority for ValuePriority and PropertyPriorities

## Sequences
- seq-create-variable.md: priority set from "priority" property
- seq-set-property.md: priority extracted from property name suffix
