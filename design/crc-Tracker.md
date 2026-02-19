# Tracker
**Source Spec:** main.md, api.md
**Requirements:** R1, R2, R3, R4, R5, R6, R35, R36, R37, R38, R39, R40, R41, R51, R52, R53, R54, R55, R56, R57, R58, R70, R71, R72, R73, R74, R80, R81, R82, R85, R86, R87

## Responsibilities

### Knows
- variables: map[int64]*Variable - all tracked variables indexed by ID
- nextID: int64 - next variable ID to assign (starts at 1)
- rootIDs: map[int64]bool - set of root variable IDs (variables with ParentID == 0) for efficient tree traversal
- valueChanges: map[int64]bool - set of variable IDs with value changes
- propertyChanges: map[int64][]string - map of variable IDs to changed property names
- sortedChanges: []Change - reusable slice for sortChanges output (flat array, not pointers)
- objectRegistry: map[uintptr]weakEntry - weak map from object pointers to variable IDs
- Resolver: Resolver - pluggable resolver for path navigation (defaults to self)
- ChangeCount: int64 - incremented each time DetectChanges() finds any changes
- DiagLevel: int - diagnostic level (0 = disabled); controls which Diag() calls are collected
- computingVar: *Variable - variable currently having its value computed (nil when idle)

### Does
- NewTracker(): creates new tracker instance with self as resolver
- CreateVariableWithId(id, value, parentID, path, props): creates variable with caller-specified ID; returns nil if ID in use; does NOT touch nextID
- CreateVariable(value, parentID, path, props): increments nextID, then delegates to CreateVariableWithId
- GetVariable(id): retrieves variable by ID
- DestroyVariable(id): removes variable, unregisters object, removes from change tracking, removes from rootIDs if root, removes ID from parent's ChildIDs if child
- DetectChanges(): collects active variables via tree traversal (respecting Active flag), groups by priority, checks in priority order with parent-before-child guarantee via checked set, returns bool
- collectActiveVariables(id, &high, &med, &low) (internal): recursive tree walk that collects readable variables into priority buckets; skips inactive subtrees; skips non-readable but still collects their children
- checkWithAncestors(id, checked) (internal): ensures readable ancestors are checked before this variable; uses checked set to prevent double-processing; delegates to checkSingleVariable
- checkSingleVariable(id) (internal): checks one variable for changes without recursion; gets value, converts to ValueJSON, compares, updates cache if changed
- sortChanges() (internal): returns []Change sorted by priority (high -> medium -> low), reuses sortedChanges slice
- recordPropertyChange(varID, propName): records a property change (called by Variable.SetProperty)
- Variables(): returns all variables
- RootVariables(): returns variables with no parent (uses rootIDs set)
- Children(parentID): returns child variables of a parent (uses parent's ChildIDs)
- UnregisterObject(obj): removes object from registry
- LookupObject(obj): finds ID for registered object
- GetObject(id): retrieves object by ID (may return nil if collected)
- ToValueJSON(value): serializes value to Value JSON form; auto-registers unregistered pointers/maps (this is the ONLY way objects get registered)
- ToValueJSONBytes(value): serializes value to JSON bytes
- Get(obj, pathElement): resolver implementation using reflection
- Set(obj, pathElement, value): resolver implementation using reflection
- Call(obj, methodName): resolver implementation - invokes zero-arg method via reflection
- CallWith(obj, methodName, value): resolver implementation - invokes one-arg void method via reflection
- Diag(level, format, args...): if DiagLevel >= level and computingVar != nil, appends formatted string to computingVar.Diags

## Collaborators
- Variable: creates and manages variables
- Resolver: uses resolver for path navigation (often itself)
- ObjectRef: produces object references during serialization
- Change: produces Change objects in sortChanges
- Priority: uses Priority for sorting changes

## Sequences
- seq-create-variable.md: variable creation, registration, parent ChildIDs update, rootIDs update
- seq-destroy-variable.md: variable destruction, unregistration, parent ChildIDs update, rootIDs update
- seq-detect-changes.md: change detection workflow with tree traversal (includes sorting and clearing)
- seq-get-value.md: getting values via path resolution
- seq-set-value.md: setting values via path resolution
- seq-to-value-json.md: value serialization
- seq-set-property.md: property change recording
