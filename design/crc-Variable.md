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
- WrapperValue: any - optional wrapper object for child navigation (created via Resolver.CreateWrapper when "wrapper" property is set)
- WrapperJSON: any - serialized WrapperValue (ToValueJSON)
- tracker: *Tracker - reference to owning tracker (for resolver access)

### Does
- Get(): checks access (error if "w" or "action"), navigates from parent's NavigationValue using path, returns current value
- Set(value): checks access (error if "r"), navigates from parent's NavigationValue to target location and sets value; for write-only or action variables with `()` paths, calls the method for side effects
- Parent(): returns parent variable or nil
- SetActive(active bool): sets whether the variable and its children participate in change detection
- NavigationValue(): returns WrapperValue if present, otherwise Value (used by child variables for path navigation)
- GetAccess(): returns access mode ("r", "w", "rw", or "action")
- IsReadable(): returns true if access allows reading ("r" or "rw")
- IsWritable(): returns true if access allows writing ("w", "rw", or "action")
- GetProperty(name): returns property value or empty string
- SetProperty(name, value): sets or removes property, handles priority suffixes, records change in tracker
  - Handles priority suffixes (:low, :medium, :high)
  - Setting "priority" property updates ValuePriority
  - Setting "path" property re-parses and updates Path field
  - Setting "access" property updates Access field (validates: r, w, rw, action)
  - Setting "wrapper" property triggers wrapper update (creates or destroys wrapper)
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

| Access   | Valid Path Endings        | Invalid Path Endings |
|----------|---------------------------|---------------------|
| `rw`     | fields, indices, `()`     | `(_)`               |
| `r`      | fields, indices, `()`     | `(_)`               |
| `w`      | fields, indices, `(_)`    | `()`                |
| `action` | `()`, `(_)`               | (none)              |

- Paths ending in `(_)` require `access: "w"` or `access: "action"` (not `r` or `rw`)
- Paths ending in `()` are allowed with `rw`, `r`, or `action` access (supports variadic method calls)
- With `rw` access and `()` path: Get() calls method with no args, Set() calls method with args

Validation errors at CreateVariable:
- `access: "r"` or `access: "rw"` with path ending in `(_)` -> error (cannot read from setter)
- `access: "w"` with path ending in `()` -> error (use `rw`, `r`, or `action` for zero-arg methods)

### Wrapper Support

A variable can have an optional wrapper that stands in for its value when child variables navigate paths. This allows the resolver to provide a different interface to children than the underlying value.

**Wrapper Lifecycle:**
1. When "wrapper" property is set AND ValueJSON is non-nil:
   - Resolver.CreateWrapper(variable) is called
   - If it returns non-nil, the wrapper is registered and stored in WrapperValue
   - WrapperJSON stores the serialized form of the wrapper
2. When ValueJSON changes (in DetectChanges or when wrapper property changes):
   - CreateWrapper(v) is called
   - If the returned wrapper is the **same pointer** as v.WrapperValue:
     - No action taken (wrapper preserved with its state)
   - If the returned wrapper is **different** (including nil):
     - Old wrapper is unregistered from the object registry
     - New wrapper is registered (if non-nil)
     - WrapperValue is updated
     - WrapperJSON is recomputed
3. When "wrapper" property is cleared or variable is destroyed:
   - Wrapper is unregistered and cleared

This design allows `CreateWrapper` to:
- Return the same wrapper object (modified in place) to preserve persistent state
- Return a new wrapper when replacement is needed
- Return nil to remove the wrapper

**Child Navigation:**
Child variables use NavigationValue() to get the starting point for path resolution:
- If WrapperValue is non-nil, children navigate from WrapperValue
- Otherwise, children navigate from Value

**Usage:**
```go
// Custom resolver that creates wrappers
type myResolver struct {
    *Tracker
}

func (r *myResolver) CreateWrapper(v *Variable) any {
    // Return a wrapper object that exposes a different interface
    return &MyWrapper{original: v.Value}
}

tr := NewTracker()
tr.Resolver = &myResolver{tr}
parent := tr.CreateVariable(myData, 0, "?wrapper=true", nil)
// Children now navigate through MyWrapper instead of myData
```

**Preserving Wrapper State:**
```go
func (r *myResolver) CreateWrapper(v *Variable) any {
    // Reuse existing wrapper to preserve state
    if w, ok := v.WrapperValue.(*StatefulWrapper); ok {
        w.data = v.Value.(*MyData)  // Update data reference
        return w                     // Same pointer preserves cache
    }
    // Create new wrapper
    return &StatefulWrapper{
        cache: make(map[string]any),
        data:  v.Value.(*MyData),
    }
}
```
