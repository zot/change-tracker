# Variable
**Source Spec:** main.md, api.md

## Responsibilities

### Knows
- ID: int64 - unique identifier assigned by tracker
- ParentID: int64 - ID of parent variable (0 = root)
- ChildIDs: []int64 - IDs of child variables (maintained automatically by tracker)
- Active: bool - whether this variable and its children are checked for changes (default: true)
- Properties: map[string]string - metadata (path stored separately)
- PropertyPriorities: map[string]Priority - priority for each property
- Path: []any - parsed path elements
- Value: any - cached value for child navigation
- ValueJSON: any - cached Value JSON for change detection
- ValuePriority: Priority - priority of the value (set via "priority" property)
- tracker: *Tracker - reference to owning tracker (for resolver access)

### Does
- Get(): navigates from parent's cached value using path, returns current value
- Set(value): navigates to target location and sets value
- Parent(): returns parent variable or nil
- SetActive(active bool): sets whether the variable and its children participate in change detection
- GetProperty(name): returns property value or empty string
- SetProperty(name, value): sets or removes property, handles priority suffixes, records change in tracker
  - Handles priority suffixes (:low, :medium, :high)
  - Setting "priority" property updates ValuePriority
  - Setting "path" property re-parses and updates Path field
  - Records property change in tracker for DetectChanges
- GetPropertyPriority(name): returns priority for a property (default: PriorityMedium)

## Collaborators
- Tracker: uses tracker's resolver for path navigation, references parent variables
- Resolver: indirectly via tracker for Get/Set operations
- Priority: uses Priority type for ValuePriority and PropertyPriorities

## Sequences
- seq-get-value.md: getting variable's value
- seq-set-value.md: setting variable's value
- seq-set-property.md: setting property with priority handling
- seq-detect-changes.md: participates in change detection via tree traversal
- seq-create-variable.md: ChildIDs maintained when created
- seq-destroy-variable.md: ChildIDs maintained when destroyed
