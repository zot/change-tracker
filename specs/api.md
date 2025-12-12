# Change Tracker API

Detailed API documentation for `github.com/zot/change-tracker`.

## Package

```go
package changetracker
```

## Types

### Priority

Priority level for values and properties.

```go
type Priority int

const (
    PriorityLow    Priority = -1
    PriorityMedium Priority = 0  // default
    PriorityHigh   Priority = 1
)
```

### Tracker

The change tracker.

```go
type Tracker struct {
    Resolver Resolver  // defaults to the tracker itself
    // Internal fields for variable storage, ID generation, changed set, object registry, root variable IDs
}
```

### Variable

A tracked variable.

```go
type Variable struct {
    ID                 int64
    ParentID           int64
    ChildIDs           []int64   // IDs of child variables (maintained automatically)
    Active             bool      // whether this variable and its children are checked for changes (default: true)
    Properties         map[string]string
    PropertyPriorities map[string]Priority
    Path               []any     // parsed path elements
    Value              any       // cached value for child navigation
    ValueJSON          any       // cached Value JSON for change detection
    ValuePriority      Priority  // priority of the value (from "priority" property)
}
```

The path specifies how to navigate from the parent's value to this variable's value. It is parsed into the `Path` field.

The ChildIDs field is maintained automatically by the tracker: when a variable is created with a parent, its ID is added to the parent's ChildIDs; when destroyed, it is removed.

The Active field controls change detection: when true (the default), the variable and its children are checked for changes. When false, the variable and all its descendants are skipped during `DetectChanges()`.

### Resolver

Interface for navigating into values.

```go
type Resolver interface {
    // Get retrieves a value at the given path element within obj.
    // pathElement can be:
    //   - string: field name, map key, or method name (with "()" suffix for methods)
    //   - int: slice/array index (0-based)
    Get(obj any, pathElement any) (any, error)

    // Set assigns a value at the given path element within obj.
    // pathElement types same as Get.
    // Returns error if the path element doesn't exist or isn't settable.
    Set(obj any, pathElement any, value any) error
}
```

### ObjectRef

Represents an object reference in Value JSON form.

```go
type ObjectRef struct {
    Obj int64 `json:"obj"`
}
```

### Change

Represents a change to a variable. Used in sorted change lists.

```go
type Change struct {
    VariableID        int64
    Priority          Priority
    ValueChanged      bool
    PropertiesChanged []string  // names of changed properties at this priority level
}
```

A single variable may produce multiple Change entries if its value and properties have different priorities. For example, a variable with a high-priority value change and a low-priority property change would appear twice in the sorted changes slice.

## Tracker Methods

### NewTracker

Creates a new change tracker.

```go
func NewTracker() *Tracker
```

### CreateVariable

Creates a new variable in the tracker.

```go
func (t *Tracker) CreateVariable(value any, parentID int64, path string, properties map[string]string) *Variable
```

**Parameters:**
- `value` - The initial value (used for root variables; ignored for child variables which derive value from path)
- `parentID` - ID of the parent variable (0 = no parent, making this a root variable)
- `path` - Path string with optional URL-style query parameters for properties (see below)
- `properties` - Optional metadata map (can be nil)

**Path Parameter:**
- If `path` is empty, uses `properties["path"]` as the path
- If `path` is non-empty, it overrides any `path` in properties
- Path can include URL-style query syntax: `"a.b?width=1&height=2"`
- Properties in the path query override those in the properties map
- The `priority` property (if present) sets the variable's `ValuePriority`

**Returns:** The created variable with an assigned ID.

**Behavior:**
1. Assigns a unique ID to the variable (incrementing from 1)
2. Sets `Active` to true (default)
3. For root variables (parentID == 0): adds the variable ID to the root variable set
4. For child variables: adds the new variable's ID to the parent's `ChildIDs`
5. Merges properties: starts with `properties` map, overlays properties from path query
6. Parses the path portion into path elements
7. Sets `ValuePriority` from the `priority` property (if present)
8. For root variables: caches the provided value
9. For child variables: calls `Get()` to compute and cache the value from parent
10. If the cached value is a pointer or map, registers it in the object registry
11. Converts cached value to Value JSON and stores for change detection
12. If `properties` is nil, initializes an empty map
13. Stores the variable in the tracker

### GetVariable

Retrieves a variable by ID.

```go
func (t *Tracker) GetVariable(id int64) *Variable
```

**Returns:** The variable, or nil if not found.

### DestroyVariable

Removes a variable from the tracker.

```go
func (t *Tracker) DestroyVariable(id int64)
```

**Behavior:**
- Removes the variable from the tracker
- For root variables: removes the variable ID from the root variable set
- For child variables: removes the variable ID from the parent's `ChildIDs`
- Removes the variable from the changed set if present
- Unregisters the object from the object registry (if it was a pointer)
- Does not automatically destroy child variables (caller's responsibility)

### DetectChanges

Compares current Value JSON to stored Value JSON using tree traversal, updates the changed set, and returns sorted changes.

```go
func (t *Tracker) DetectChanges() []Change
```

**Returns:** A slice of Change objects sorted by priority (high → medium → low).

**Behavior:**
1. For each root variable ID in the root variable set:
   - Perform a depth-first traversal starting from the root variable
   - For each variable visited:
     - If the variable is inactive (`Active == false`), skip it and do not visit its children
     - If the variable is active:
       - Get the current value and convert to Value JSON
       - Compare to the stored Value JSON
       - If different, mark the variable's value as changed
       - Update the stored Value JSON to the current Value JSON
       - Recursively visit all child variables
2. Sort all changes (value and property) by priority
3. Clear the internal change records but preserve the sorted changes slice
4. Return the sorted changes

**Notes:**
- Property changes are recorded immediately when `SetProperty()` is called, not during `DetectChanges()`. The sorting step collects both value changes detected in step 1 and any property changes recorded since the last `DetectChanges()`.
- A variable may appear multiple times in the result if it has changes at different priority levels (e.g., high-priority value change and low-priority property change).
- Reuses an internal slice to minimize allocations. The returned slice is valid until the next call to `DetectChanges()`.

### Variables

Returns all variables in the tracker.

```go
func (t *Tracker) Variables() []*Variable
```

**Returns:** A slice of all variables (order not guaranteed).

### RootVariables

Returns variables with no parent (parentID == 0).

```go
func (t *Tracker) RootVariables() []*Variable
```

### Children

Returns child variables of a given parent.

```go
func (t *Tracker) Children(parentID int64) []*Variable
```

## Object Registry Methods

### RegisterObject

Manually registers an object with a variable ID.

```go
func (t *Tracker) RegisterObject(obj any, varID int64) bool
```

**Parameters:**
- `obj` - The object to register (must be a pointer)
- `varID` - The variable ID to associate with this object

**Returns:** `true` if registered, `false` if obj is not a pointer.

**Note:** Objects are automatically registered when `CreateVariable` is called with a pointer value. This method is for manual registration when needed.

### UnregisterObject

Removes an object from the registry.

```go
func (t *Tracker) UnregisterObject(obj any)
```

### LookupObject

Finds the variable ID for a registered object.

```go
func (t *Tracker) LookupObject(obj any) (int64, bool)
```

**Returns:** The variable ID and `true` if found, or `0` and `false` if not registered.

### GetObject

Retrieves an object by its variable ID.

```go
func (t *Tracker) GetObject(varID int64) any
```

**Returns:** The object, or nil if not found or if the weak reference has been collected.

## Value JSON Methods

### ToValueJSON

Serializes a value to Value JSON form.

```go
func (t *Tracker) ToValueJSON(value any) any
```

**Returns:** The value in Value JSON form:
- Primitives (string, number, bool, nil) pass through unchanged
- Registered objects (pointers, maps) become `ObjectRef{Obj: id}`
- Slices/arrays become slices with elements in Value JSON form
- Unregistered pointers/maps cause an error

### ToValueJSONBytes

Serializes a value to Value JSON as a byte slice.

```go
func (t *Tracker) ToValueJSONBytes(value any) ([]byte, error)
```

**Returns:** JSON-encoded bytes of the Value JSON form.

### IsObjectRef

Checks if a value is an object reference.

```go
func IsObjectRef(value any) bool
```

### GetObjectRefID

Extracts the ID from an object reference.

```go
func GetObjectRefID(value any) (int64, bool)
```

**Returns:** The object ID and `true` if value is an ObjectRef, otherwise `0` and `false`.

## Tracker as Resolver

The Tracker type implements the Resolver interface, providing reflection-based value navigation.

```go
func (t *Tracker) Get(obj any, pathElement any) (any, error)
func (t *Tracker) Set(obj any, pathElement any, value any) error
```

### Get Behavior

**String path elements:**
- Struct field: Looks up by field name (exported fields only)
- Map key: Looks up by string key
- Method: If pathElement ends with "()", calls the zero-argument method

**Integer path elements:**
- Slice/array: Returns element at index (0-based)

**Errors:**
- Returns error if obj is nil
- Returns error if path element not found
- Returns error if path element type is unsupported
- Returns error if method requires arguments or returns no values

### Set Behavior

**String path elements:**
- Struct field: Sets the field (must be settable - pointer to struct required)
- Map key: Sets the map entry

**Integer path elements:**
- Slice: Sets element at index (must be within bounds)

**Errors:**
- Returns error if obj is nil or not a pointer (for struct fields)
- Returns error if field/key doesn't exist
- Returns error if field isn't settable
- Returns error if value type doesn't match

## Variable Methods

### Get

Gets the variable's value by navigating from the parent's value using the path.

```go
func (v *Variable) Get() (any, error)
```

**Behavior:**
1. Get the parent's cached value (or nil for root variables)
2. Apply each path element using the tracker's resolver
3. Cache the result for child navigation
4. Return the value

**Returns:** The value, or error if navigation fails.

### Set

Sets the variable's value by navigating from the parent's value using the path.

```go
func (v *Variable) Set(value any) error
```

**Parameters:**
- `value` - The value to set

**Behavior:**
1. Get the parent's cached value
2. Navigate to the parent of the target using all but the last path element
3. Use the resolver to set the value at the last path element

**Returns:** Error if navigation or setting fails.

### Parent

Returns the parent variable, or nil if this is a root variable.

```go
func (v *Variable) Parent() *Variable
```

### SetActive

Sets whether the variable and its children should be checked for changes.

```go
func (v *Variable) SetActive(active bool)
```

**Parameters:**
- `active` - When true (default), the variable and its children participate in change detection. When false, the variable and all its descendants are skipped during `DetectChanges()`.

**Note:** This change takes effect on the next `DetectChanges()` call. Setting a variable to inactive effectively "prunes" that entire subtree from change detection.

### GetProperty

```go
func (v *Variable) GetProperty(name string) string
```

Returns the property value, or empty string if not set.

### SetProperty

```go
func (v *Variable) SetProperty(name, value string)
```

Sets a property. Empty value removes the property.

**Priority Suffixes:**
- Property names can include a priority suffix: `:low`, `:medium`, `:high`
- Example: `SetProperty("label:high", "Important")` sets `Properties["label"]` with `PropertyPriorities["label"] = PriorityHigh`
- Without a suffix, the property defaults to `PriorityMedium`

**Special Properties:**
- Setting `priority` (values: `"low"`, `"medium"`, `"high"`) updates `ValuePriority`
- Setting `path` re-parses the path and updates the `Path` field

**Change Tracking:**
- Records the property change in the tracker (property name added to changed properties)
- The change appears in the result of the next `DetectChanges()` call at the property's priority level

### GetPropertyPriority

```go
func (v *Variable) GetPropertyPriority(name string) Priority
```

Returns the priority for a property, or `PriorityMedium` if not explicitly set.

## Comparison Strategy

Change detection compares Value JSON representations. Each variable stores its last known Value JSON, and `DetectChanges()` compares the current Value JSON to the stored one.

This means:
- Primitives compare by value
- Arrays compare element by element (in Value JSON form)
- Registered objects compare by their object reference `{"obj": ID}`
- Two references to the same registered object are always equal

## Weak Reference Behavior

The object registry uses Go 1.24+ weak references (`weak.Pointer`):

- Registered objects don't prevent garbage collection
- When an object is collected, its registry entry is automatically cleaned up
- `LookupObject` may return `false` for objects that were collected
- `GetObject` may return `nil` for collected objects
- The variable remains in the tracker even if its object is collected

This allows long-running applications to register many objects without memory leaks, as objects are naturally cleaned up when no longer referenced by application code.
