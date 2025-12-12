# Tracker
**Source Spec:** main.md, api.md

## Responsibilities

### Knows
- variables: map[int64]*Variable - all tracked variables indexed by ID
- nextID: int64 - next variable ID to assign (starts at 1)
- valueChanges: map[int64]bool - set of variable IDs with value changes
- propertyChanges: map[int64][]string - map of variable IDs to changed property names
- sortedChanges: []Change - reusable slice for sortChanges output (flat array, not pointers)
- objectRegistry: map[uintptr]weakEntry - weak map from object pointers to variable IDs
- Resolver: Resolver - pluggable resolver for path navigation (defaults to self)

### Does
- NewTracker(): creates new tracker instance with self as resolver
- CreateVariable(value, parentID, path, props): creates variable with path+query parsing, assigns ID, registers objects, caches value
- GetVariable(id): retrieves variable by ID
- DestroyVariable(id): removes variable, unregisters object, removes from change tracking
- DetectChanges(): compares current values to cached ValueJSON, marks value as changed, calls sortChanges, clears internal change records, returns []Change sorted by priority
- sortChanges() (internal): returns []Change sorted by priority (high -> medium -> low), reuses sortedChanges slice
- recordPropertyChange(varID, propName): records a property change (called by Variable.SetProperty)
- Variables(): returns all variables
- RootVariables(): returns variables with no parent
- Children(parentID): returns child variables of a parent
- RegisterObject(obj, varID): manually registers pointer/map in object registry
- UnregisterObject(obj): removes object from registry
- LookupObject(obj): finds variable ID for registered object
- GetObject(varID): retrieves object by variable ID (may return nil if collected)
- ToValueJSON(value): serializes value to Value JSON form
- ToValueJSONBytes(value): serializes value to JSON bytes
- Get(obj, pathElement): resolver implementation using reflection
- Set(obj, pathElement, value): resolver implementation using reflection

## Collaborators
- Variable: creates and manages variables
- Resolver: uses resolver for path navigation (often itself)
- ObjectRef: produces object references during serialization
- Change: produces Change objects in sortChanges
- Priority: uses Priority for sorting changes

## Sequences
- seq-create-variable.md: variable creation and registration
- seq-detect-changes.md: change detection workflow (includes sorting and clearing)
- seq-get-value.md: getting values via path resolution
- seq-set-value.md: setting values via path resolution
- seq-to-value-json.md: value serialization
- seq-set-property.md: property change recording
