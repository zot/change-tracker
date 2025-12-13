# Variable
**Source Spec:** main.md, api.md, resolver.md

## Responsibilities

### Knows
- ID: int64 - unique identifier assigned by tracker
- ParentID: int64 - ID of parent variable (0 = root)
- ChildIDs: []int64 - IDs of child variables (maintained automatically by tracker)
- Active: bool - whether this variable and its children are checked for changes (default: true)
- Access: string - access mode: "r" (read-only), "w" (write-only), "rw" (read-write, default)
- Properties: map[string]string - metadata (path stored separately)
- PropertyPriorities: map[string]Priority - priority for each property
- Path: []any - parsed path elements
- Value: any - cached value for child navigation
- ValueJSON: any - cached Value JSON for change detection
- ValuePriority: Priority - priority of the value (set via "priority" property)
- tracker: *Tracker - reference to owning tracker (for resolver access)

### Does
- Get(): checks access (error if "w"), navigates from parent's cached value using path, returns current value
- Set(value): checks access (error if "r"), navigates to target location and sets value; for write-only variables with `()` paths, calls the method for side effects
- Parent(): returns parent variable or nil
- SetActive(active bool): sets whether the variable and its children participate in change detection
- GetAccess(): returns access mode ("r", "w", or "rw")
- IsReadable(): returns true if access allows reading ("r" or "rw")
- IsWritable(): returns true if access allows writing ("w" or "rw")
- GetProperty(name): returns property value or empty string
- SetProperty(name, value): sets or removes property, handles priority suffixes, records change in tracker
  - Handles priority suffixes (:low, :medium, :high)
  - Setting "priority" property updates ValuePriority
  - Setting "path" property re-parses and updates Path field
  - Setting "access" property updates Access field (validates: r, w, rw)
  - Records property change in tracker for DetectChanges
- GetPropertyPriority(name): returns priority for a property (default: PriorityMedium)

## Collaborators
- Tracker: uses tracker's resolver for path navigation, references parent variables
- Resolver: indirectly via tracker for Get/Set operations
- Priority: uses Priority type for ValuePriority and PropertyPriorities

## Sequences
- seq-get-value.md: getting variable's value (access check: error if write-only)
- seq-set-value.md: setting variable's value (access check: error if read-only)
- seq-set-property.md: setting property with priority handling
- seq-detect-changes.md: participates in change detection via tree traversal (skips write-only variables)
- seq-create-variable.md: ChildIDs maintained when created
- seq-destroy-variable.md: ChildIDs maintained when destroyed

## Notes

### Access Property
The `access` property controls read/write permissions independent of path semantics:

| Value | Get | Set | Scanned for Changes |
|-------|-----|-----|---------------------|
| `rw` (default) | OK | OK | Yes |
| `r` | OK | Error | Yes |
| `w` | Error | OK | No |

Access checks occur before path-based checks. A write-only variable (`access: "w"`) is excluded from change detection scans because its value cannot be read.

For write-only variables with `()` paths (zero-arg methods), Set() calls the method for its side effects. This allows triggering actions without reading the return value.
