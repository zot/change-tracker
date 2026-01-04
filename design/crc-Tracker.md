# Tracker
**Source Spec:** main.md, api.md

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

### Does
- NewTracker(): creates new tracker instance with self as resolver
- CreateVariable(value, parentID, path, props): creates variable with path+query parsing, assigns ID, caches value, calls ToValueJSON (which auto-registers objects), adds to rootIDs if root, adds ID to parent's ChildIDs if child
- GetVariable(id): retrieves variable by ID
- DestroyVariable(id): removes variable, unregisters object, removes from change tracking, removes from rootIDs if root, removes ID from parent's ChildIDs if child
- DetectChanges(): performs depth-first tree traversal from root variables, skips inactive variables and their descendants, compares current values to cached ValueJSON, marks value as changed, calls sortChanges, clears internal change records, returns []Change sorted by priority
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
