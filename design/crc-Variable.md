# Variable
**Source Spec:** main.md, api.md, resolver.md

## Responsibilities

### Knows
- ID: int64 - unique identifier assigned by tracker
- ParentID: int64 - ID of parent variable (0 = root)
- ChildIDs: []int64 - IDs of child variables (maintained automatically by tracker)
- Active: bool - whether this variable and its children are checked for changes (default: true)
- Access: string - access mode: "r" (read-only), "w" (write-only), "rw" (read-write, default), "action" (action trigger)
- Properties: map[string]string - metadata (path stored separately)
- PropertyPriorities: map[string]Priority - priority for each property
- Path: []any - parsed path elements
- Value: any - cached value for child navigation
- ValueJSON: any - cached Value JSON for change detection
- ValuePriority: Priority - priority of the value (set via "priority" property)
- tracker: *Tracker - reference to owning tracker (for resolver access)

### Does
- Get(): checks access (error if "w" or "action"), navigates from parent's cached value using path, returns current value
- Set(value): checks access (error if "r"), navigates to target location and sets value; for write-only or action variables with `()` paths, calls the method for side effects
- Parent(): returns parent variable or nil
- SetActive(active bool): sets whether the variable and its children participate in change detection
- GetAccess(): returns access mode ("r", "w", "rw", or "action")
- IsReadable(): returns true if access allows reading ("r" or "rw")
- IsWritable(): returns true if access allows writing ("w", "rw", or "action")
- GetProperty(name): returns property value or empty string
- SetProperty(name, value): sets or removes property, handles priority suffixes, records change in tracker
  - Handles priority suffixes (:low, :medium, :high)
  - Setting "priority" property updates ValuePriority
  - Setting "path" property re-parses and updates Path field
  - Setting "access" property updates Access field (validates: r, w, rw, action)
  - Records property change in tracker for DetectChanges
- GetPropertyPriority(name): returns priority for a property (default: PriorityMedium)

## Collaborators
- Tracker: uses tracker's resolver for path navigation, references parent variables
- Resolver: indirectly via tracker for Get/Set operations
- Priority: uses Priority type for ValuePriority and PropertyPriorities

## Sequences
- seq-get-value.md: getting variable's value (access check: error if write-only or action)
- seq-set-value.md: setting variable's value (access check: error if read-only)
- seq-set-property.md: setting property with priority handling
- seq-detect-changes.md: participates in change detection via tree traversal (skips write-only and action variables)
- seq-create-variable.md: ChildIDs maintained when created
- seq-destroy-variable.md: ChildIDs maintained when destroyed

## Notes

### Access Property
The `access` property controls read/write permissions and initialization behavior independent of path semantics:

| Value | Get | Set | Scanned for Changes | Initial Value Computed |
|-------|-----|-----|---------------------|------------------------|
| `rw` (default) | OK | OK | Yes | Yes |
| `r` | OK | Error | Yes | Yes |
| `w` | Error | OK | No | Yes |
| `action` | Error | OK | No | No |

Access checks occur before path-based checks. Write-only (`access: "w"`) and action (`access: "action"`) variables are excluded from change detection scans because their values cannot be read.

The key difference between `w` and `action`:
- **Write-only (`w`)**: Initial value IS computed during CreateVariable. Appropriate for variables like `Password` where you want to set values but not read them back.
- **Action (`action`)**: Initial value is NOT computed during CreateVariable. Essential for action-triggering paths like `AddContact(_)` where navigating the path would invoke the action prematurely.

For write-only or action variables with `()` paths (zero-arg methods), Set() calls the method for its side effects. This allows triggering actions without reading the return value.

### Path Restrictions by Access Mode

CreateVariable validates access/path combinations:

| Access   | Valid Path Endings     | Invalid Path Endings |
|----------|------------------------|---------------------|
| `rw`     | fields, indices        | `()`, `(_)`         |
| `r`      | fields, indices, `()`  | `(_)`               |
| `w`      | fields, indices, `(_)` | `()`                |
| `action` | `()`, `(_)`            | (none)              |

- `rw` is a union of `r` and `w`, so it inherits restrictions from both: no `()` (from `w`) and no `(_)` (from `r`)
- Paths ending in `(_)` require `access: "w"` or `access: "action"` (not `r` or `rw`)
- Paths ending in `()` require `access: "r"` or `access: "action"` (not `w` or `rw`)

Validation errors at CreateVariable:
- `access: "r"` or `access: "rw"` with path ending in `(_)` -> error (cannot read from setter)
- `access: "w"` or `access: "rw"` with path ending in `()` -> error (use `action` for zero-arg methods)
