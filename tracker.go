// Package changetracker provides variable management with automatic change detection.
// CRC: crc-Tracker.md, crc-Variable.md, crc-Resolver.md, crc-ObjectRef.md, crc-ObjectRegistry.md, crc-Change.md, crc-Priority.md
// Spec: main.md, api.md
package changetracker

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"weak"
)

// Priority represents priority level for values and properties.
// CRC: crc-Priority.md
// Spec: api.md
type Priority int

const (
	PriorityLow    Priority = -1
	PriorityMedium Priority = 0 // default
	PriorityHigh   Priority = 1
)

// ParsePriority converts a string to a Priority.
func ParsePriority(s string) Priority {
	switch strings.ToLower(s) {
	case "low":
		return PriorityLow
	case "high":
		return PriorityHigh
	default:
		return PriorityMedium
	}
}

// Resolver is the interface for navigating into values.
// CRC: crc-Resolver.md
// Spec: resolver.md
type Resolver interface {
	// Get retrieves a value at the given path element within obj.
	// pathElement can be:
	//   - string: field name, map key, or method name (with "()" suffix for methods)
	//   - int: slice/array index (0-based)
	Get(obj any, pathElement any) (any, error)

	// Set assigns a value at the given path element within obj.
	Set(obj any, pathElement any, value any) error
}

// ObjectRef represents an object reference in Value JSON form.
// CRC: crc-ObjectRef.md
// Spec: value-json.md
type ObjectRef struct {
	Obj int64 `json:"obj"`
}

// IsObjectRef checks if a value is an ObjectRef.
// CRC: crc-ObjectRef.md
func IsObjectRef(value any) bool {
	_, ok := value.(ObjectRef)
	return ok
}

// GetObjectRefID extracts the ID from an ObjectRef.
// CRC: crc-ObjectRef.md
func GetObjectRefID(value any) (int64, bool) {
	ref, ok := value.(ObjectRef)
	if !ok {
		return 0, false
	}
	return ref.Obj, true
}

// Change represents a change to a variable.
// CRC: crc-Change.md
// Spec: api.md
type Change struct {
	VariableID        int64
	Priority          Priority
	ValueChanged      bool
	PropertiesChanged []string
}

// weakEntry holds a weak reference to an object and its variable ID.
type weakEntry struct {
	ptr   weak.Pointer[any]
	varID int64
}

// propertyChange tracks which properties changed for a variable.
type propertyChange struct {
	properties map[string]bool // set of changed property names
}

// Tracker is the central change tracker.
// CRC: crc-Tracker.md
// Spec: main.md, api.md
type Tracker struct {
	Resolver Resolver // defaults to the tracker itself

	variables map[int64]*Variable
	nextID    int64
	rootIDs   map[int64]bool // set of root variable IDs for efficient tree traversal

	// Change tracking
	valueChanges    map[int64]bool            // variables with value changes
	propertyChanges map[int64]*propertyChange // variables with property changes

	// Sorted changes (reused slice)
	sortedChanges []Change

	// Object registry: maps object pointer to weak entry
	// CRC: crc-ObjectRegistry.md
	ptrToEntry map[uintptr]weakEntry
	idToPtr    map[int64]uintptr
}

// NewTracker creates a new change tracker.
// Sequence: seq-create-variable.md
func NewTracker() *Tracker {
	t := &Tracker{
		variables:       make(map[int64]*Variable),
		nextID:          1,
		rootIDs:         make(map[int64]bool),
		valueChanges:    make(map[int64]bool),
		propertyChanges: make(map[int64]*propertyChange),
		sortedChanges:   make([]Change, 0, 16),
		ptrToEntry:      make(map[uintptr]weakEntry),
		idToPtr:         make(map[int64]uintptr),
	}
	t.Resolver = t // default resolver is the tracker itself
	return t
}

// Variable is a tracked variable.
// CRC: crc-Variable.md
// Spec: main.md, api.md
type Variable struct {
	ID                 int64
	ParentID           int64
	ChildIDs           []int64  // IDs of child variables (maintained automatically)
	Active             bool     // whether this variable and its children are checked for changes
	Properties         map[string]string
	PropertyPriorities map[string]Priority
	Path               []any    // parsed path elements
	Value              any      // cached value for child navigation
	ValueJSON          any      // cached Value JSON for change detection
	ValuePriority      Priority // priority of the value

	tracker *Tracker
}

// CreateVariable creates a new variable in the tracker.
// Sequence: seq-create-variable.md
func (t *Tracker) CreateVariable(value any, parentID int64, path string, properties map[string]string) *Variable {
	if properties == nil {
		properties = make(map[string]string)
	}

	v := &Variable{
		ID:                 t.nextID,
		ParentID:           parentID,
		ChildIDs:           nil, // initialized as nil, will be allocated on first child
		Active:             true,
		Properties:         make(map[string]string),
		PropertyPriorities: make(map[string]Priority),
		ValuePriority:      PriorityMedium,
		tracker:            t,
	}
	t.nextID++

	// Copy properties from the properties map
	for k, val := range properties {
		v.Properties[k] = val
	}

	// Parse path with optional query parameters
	pathPart, queryProps := parsePathWithQuery(path)

	// If path is empty, use path from properties
	if pathPart == "" {
		pathPart = properties["path"]
	}

	// Query properties override properties map
	for k, val := range queryProps {
		v.Properties[k] = val
	}

	// Store path in properties and parse it
	if pathPart != "" {
		v.Properties["path"] = pathPart
		v.Path = parsePath(pathPart)
	}

	// Set ValuePriority from priority property
	if priorityStr, ok := v.Properties["priority"]; ok {
		v.ValuePriority = ParsePriority(priorityStr)
	}

	// Cache value and manage tree structure
	if parentID == 0 {
		// Root variable: use provided value and add to rootIDs
		v.Value = value
		t.rootIDs[v.ID] = true
	} else {
		// Child variable: compute value from parent and add to parent's ChildIDs
		v.Value, _ = v.Get()
		if parent := t.variables[parentID]; parent != nil {
			parent.ChildIDs = append(parent.ChildIDs, v.ID)
		}
	}

	// Register object if pointer or map
	t.registerIfNeeded(v.Value, v.ID)

	// Cache Value JSON for change detection
	v.ValueJSON = t.ToValueJSON(v.Value)

	t.variables[v.ID] = v
	return v
}

// parsePathWithQuery splits a path into the path portion and query parameters.
// Example: "a.b?width=1&height=2" -> ("a.b", {"width": "1", "height": "2"})
func parsePathWithQuery(path string) (string, map[string]string) {
	if path == "" {
		return "", nil
	}

	idx := strings.Index(path, "?")
	if idx == -1 {
		return path, nil
	}

	pathPart := path[:idx]
	queryPart := path[idx+1:]

	props := make(map[string]string)
	if queryPart == "" {
		return pathPart, props
	}

	pairs := strings.Split(queryPart, "&")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		eqIdx := strings.Index(pair, "=")
		if eqIdx == -1 {
			props[pair] = ""
		} else {
			props[pair[:eqIdx]] = pair[eqIdx+1:]
		}
	}

	return pathPart, props
}

// parsePath splits a dot-separated path into path elements.
// Numeric strings are converted to integers for slice/array access.
func parsePath(path string) []any {
	if path == "" {
		return nil
	}
	parts := strings.Split(path, ".")
	result := make([]any, len(parts))
	for i, part := range parts {
		// Try to parse as integer for index access
		if idx, err := parseInt(part); err == nil {
			result[i] = idx
		} else {
			result[i] = part
		}
	}
	return result
}

// parseInt parses a string as a non-negative integer.
func parseInt(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	// Check for valid integer: digits only, no leading zeros (except "0")
	if s[0] == '0' && len(s) > 1 {
		return 0, fmt.Errorf("leading zero")
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not an integer")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// registerIfNeeded registers an object in the registry if it's a pointer or map.
func (t *Tracker) registerIfNeeded(value any, varID int64) {
	if value == nil {
		return
	}
	rv := reflect.ValueOf(value)
	kind := rv.Kind()
	if kind == reflect.Ptr || kind == reflect.Map {
		t.RegisterObject(value, varID)
	}
}

// GetVariable retrieves a variable by ID.
// CRC: crc-Tracker.md
func (t *Tracker) GetVariable(id int64) *Variable {
	return t.variables[id]
}

// DestroyVariable removes a variable from the tracker.
// CRC: crc-Tracker.md
// Sequence: seq-destroy-variable.md
func (t *Tracker) DestroyVariable(id int64) {
	v, ok := t.variables[id]
	if !ok {
		return
	}

	// Remove from rootIDs if root variable
	if v.ParentID == 0 {
		delete(t.rootIDs, id)
	} else {
		// Remove from parent's ChildIDs if child variable
		if parent := t.variables[v.ParentID]; parent != nil {
			for i, childID := range parent.ChildIDs {
				if childID == id {
					parent.ChildIDs = append(parent.ChildIDs[:i], parent.ChildIDs[i+1:]...)
					break
				}
			}
		}
	}

	// Unregister object if it was registered
	if v.Value != nil {
		t.UnregisterObject(v.Value)
	}

	// Remove from change tracking
	delete(t.valueChanges, id)
	delete(t.propertyChanges, id)

	// Remove from variables
	delete(t.variables, id)
}

// DetectChanges compares current values to cached ValueJSON using tree traversal,
// sorts changes by priority, clears internal change records, and returns the sorted changes.
// CRC: crc-Tracker.md
// Sequence: seq-detect-changes.md
func (t *Tracker) DetectChanges() []Change {
	// Perform depth-first traversal starting from root variables
	for rootID := range t.rootIDs {
		t.checkVariable(rootID)
	}

	// Sort changes by priority
	result := t.sortChanges()

	// Clear internal change records (but preserve the sorted changes slice)
	t.valueChanges = make(map[int64]bool)
	t.propertyChanges = make(map[int64]*propertyChange)

	return result
}

// checkVariable recursively checks a variable and its children for changes.
// If the variable is inactive, it and all its descendants are skipped.
func (t *Tracker) checkVariable(id int64) {
	v := t.variables[id]
	if v == nil {
		return
	}

	// If inactive, skip this variable and all its descendants
	if !v.Active {
		return
	}

	// Get current value
	currentValue, err := v.Get()
	if err == nil {
		// Convert to Value JSON
		currentJSON := t.ToValueJSON(currentValue)

		// Compare with cached ValueJSON
		if !jsonEqual(v.ValueJSON, currentJSON) {
			t.valueChanges[v.ID] = true
		}

		// Update cached values
		v.Value = currentValue
		v.ValueJSON = currentJSON

		// Re-register if value changed to a new pointer/map
		t.registerIfNeeded(currentValue, v.ID)
	}

	// Recursively check all children
	for _, childID := range v.ChildIDs {
		t.checkVariable(childID)
	}
}

// jsonEqual compares two Value JSON values for equality.
func jsonEqual(a, b any) bool {
	// Use JSON serialization for comparison
	aBytes, err1 := json.Marshal(a)
	bBytes, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return reflect.DeepEqual(a, b)
	}
	return string(aBytes) == string(bBytes)
}

// sortChanges returns changes sorted by priority (high -> medium -> low).
// This is an internal method called by DetectChanges.
// CRC: crc-Tracker.md
// Sequence: seq-detect-changes.md
func (t *Tracker) sortChanges() []Change {
	// Reset the reusable slice
	t.sortedChanges = t.sortedChanges[:0]

	// Collect changes by priority
	highChanges := make([]Change, 0)
	mediumChanges := make([]Change, 0)
	lowChanges := make([]Change, 0)

	// Build combined set of changed variable IDs from valueChanges and propertyChanges
	changedIDs := make(map[int64]bool)
	for id := range t.valueChanges {
		changedIDs[id] = true
	}
	for id := range t.propertyChanges {
		changedIDs[id] = true
	}

	// Process all changed variables
	for id := range changedIDs {
		v := t.variables[id]
		if v == nil {
			continue
		}

		valueChanged := t.valueChanges[id]
		propChange := t.propertyChanges[id]

		// Group properties by priority
		highProps := make([]string, 0)
		mediumProps := make([]string, 0)
		lowProps := make([]string, 0)

		if propChange != nil {
			for propName := range propChange.properties {
				priority := v.PropertyPriorities[propName]
				switch priority {
				case PriorityHigh:
					highProps = append(highProps, propName)
				case PriorityLow:
					lowProps = append(lowProps, propName)
				default:
					mediumProps = append(mediumProps, propName)
				}
			}
		}

		// Add value change at its priority level
		if valueChanged {
			switch v.ValuePriority {
			case PriorityHigh:
				// Combine with high-priority properties or create new
				if len(highProps) > 0 {
					highChanges = append(highChanges, Change{
						VariableID:        id,
						Priority:          PriorityHigh,
						ValueChanged:      true,
						PropertiesChanged: highProps,
					})
					highProps = nil // consumed
				} else {
					highChanges = append(highChanges, Change{
						VariableID:   id,
						Priority:     PriorityHigh,
						ValueChanged: true,
					})
				}
			case PriorityLow:
				if len(lowProps) > 0 {
					lowChanges = append(lowChanges, Change{
						VariableID:        id,
						Priority:          PriorityLow,
						ValueChanged:      true,
						PropertiesChanged: lowProps,
					})
					lowProps = nil
				} else {
					lowChanges = append(lowChanges, Change{
						VariableID:   id,
						Priority:     PriorityLow,
						ValueChanged: true,
					})
				}
			default: // Medium
				if len(mediumProps) > 0 {
					mediumChanges = append(mediumChanges, Change{
						VariableID:        id,
						Priority:          PriorityMedium,
						ValueChanged:      true,
						PropertiesChanged: mediumProps,
					})
					mediumProps = nil
				} else {
					mediumChanges = append(mediumChanges, Change{
						VariableID:   id,
						Priority:     PriorityMedium,
						ValueChanged: true,
					})
				}
			}
		}

		// Add remaining property-only changes at their priority levels
		if len(highProps) > 0 {
			highChanges = append(highChanges, Change{
				VariableID:        id,
				Priority:          PriorityHigh,
				PropertiesChanged: highProps,
			})
		}
		if len(mediumProps) > 0 {
			mediumChanges = append(mediumChanges, Change{
				VariableID:        id,
				Priority:          PriorityMedium,
				PropertiesChanged: mediumProps,
			})
		}
		if len(lowProps) > 0 {
			lowChanges = append(lowChanges, Change{
				VariableID:        id,
				Priority:          PriorityLow,
				PropertiesChanged: lowProps,
			})
		}
	}

	// Concatenate in priority order: high, medium, low
	t.sortedChanges = append(t.sortedChanges, highChanges...)
	t.sortedChanges = append(t.sortedChanges, mediumChanges...)
	t.sortedChanges = append(t.sortedChanges, lowChanges...)

	return t.sortedChanges
}

// recordPropertyChange records that a property changed for a variable.
// CRC: crc-Tracker.md
// Sequence: seq-set-property.md
func (t *Tracker) recordPropertyChange(varID int64, propName string) {
	pc := t.propertyChanges[varID]
	if pc == nil {
		pc = &propertyChange{properties: make(map[string]bool)}
		t.propertyChanges[varID] = pc
	}
	pc.properties[propName] = true
}

// Variables returns all variables in the tracker.
// CRC: crc-Tracker.md
func (t *Tracker) Variables() []*Variable {
	result := make([]*Variable, 0, len(t.variables))
	for _, v := range t.variables {
		result = append(result, v)
	}
	return result
}

// RootVariables returns variables with no parent (parentID == 0).
// CRC: crc-Tracker.md
func (t *Tracker) RootVariables() []*Variable {
	result := make([]*Variable, 0, len(t.rootIDs))
	for id := range t.rootIDs {
		if v := t.variables[id]; v != nil {
			result = append(result, v)
		}
	}
	return result
}

// Children returns child variables of a given parent.
// CRC: crc-Tracker.md
func (t *Tracker) Children(parentID int64) []*Variable {
	parent := t.variables[parentID]
	if parent == nil {
		return nil
	}
	result := make([]*Variable, 0, len(parent.ChildIDs))
	for _, childID := range parent.ChildIDs {
		if v := t.variables[childID]; v != nil {
			result = append(result, v)
		}
	}
	return result
}

// RegisterObject manually registers an object with a variable ID.
// CRC: crc-Tracker.md, crc-ObjectRegistry.md
// Sequence: seq-create-variable.md
func (t *Tracker) RegisterObject(obj any, varID int64) bool {
	if obj == nil {
		return false
	}
	rv := reflect.ValueOf(obj)
	kind := rv.Kind()
	if kind != reflect.Ptr && kind != reflect.Map {
		return false
	}

	ptr := rv.Pointer()

	// Create weak reference
	entry := weakEntry{
		ptr:   weak.Make(&obj),
		varID: varID,
	}

	t.ptrToEntry[ptr] = entry
	t.idToPtr[varID] = ptr
	return true
}

// UnregisterObject removes an object from the registry.
// CRC: crc-Tracker.md, crc-ObjectRegistry.md
func (t *Tracker) UnregisterObject(obj any) {
	if obj == nil {
		return
	}
	rv := reflect.ValueOf(obj)
	kind := rv.Kind()
	if kind != reflect.Ptr && kind != reflect.Map {
		return
	}

	ptr := rv.Pointer()
	if entry, ok := t.ptrToEntry[ptr]; ok {
		delete(t.idToPtr, entry.varID)
		delete(t.ptrToEntry, ptr)
	}
}

// LookupObject finds the variable ID for a registered object.
// CRC: crc-Tracker.md, crc-ObjectRegistry.md
// Sequence: seq-to-value-json.md
func (t *Tracker) LookupObject(obj any) (int64, bool) {
	if obj == nil {
		return 0, false
	}
	rv := reflect.ValueOf(obj)
	kind := rv.Kind()
	if kind != reflect.Ptr && kind != reflect.Map {
		return 0, false
	}

	ptr := rv.Pointer()
	entry, ok := t.ptrToEntry[ptr]
	if !ok {
		return 0, false
	}

	// Check if object is still alive
	if entry.ptr.Value() == nil {
		// Object was collected, clean up
		delete(t.idToPtr, entry.varID)
		delete(t.ptrToEntry, ptr)
		return 0, false
	}

	return entry.varID, true
}

// GetObject retrieves an object by its variable ID.
// CRC: crc-Tracker.md, crc-ObjectRegistry.md
func (t *Tracker) GetObject(varID int64) any {
	ptr, ok := t.idToPtr[varID]
	if !ok {
		return nil
	}

	entry, ok := t.ptrToEntry[ptr]
	if !ok {
		return nil
	}

	obj := entry.ptr.Value()
	if obj == nil {
		// Object was collected, clean up
		delete(t.idToPtr, varID)
		delete(t.ptrToEntry, ptr)
		return nil
	}

	return *obj
}

// ToValueJSON serializes a value to Value JSON form.
// Sequence: seq-to-value-json.md
func (t *Tracker) ToValueJSON(value any) any {
	if value == nil {
		return nil
	}

	rv := reflect.ValueOf(value)

	// Handle pointers and maps - must be registered
	switch rv.Kind() {
	case reflect.Ptr, reflect.Map:
		if id, ok := t.LookupObject(value); ok {
			return ObjectRef{Obj: id}
		}
		// Unregistered pointer/map - this is an error condition per spec
		// but we'll return nil to avoid panic
		return nil

	case reflect.Slice, reflect.Array:
		result := make([]any, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			result[i] = t.ToValueJSON(rv.Index(i).Interface())
		}
		return result

	default:
		// Primitives pass through
		return value
	}
}

// ToValueJSONBytes serializes a value to Value JSON as a byte slice.
// CRC: crc-Tracker.md
func (t *Tracker) ToValueJSONBytes(value any) ([]byte, error) {
	valueJSON := t.ToValueJSON(value)
	return json.Marshal(valueJSON)
}

// Get implements the Resolver interface using reflection.
// Sequence: seq-get-value.md
func (t *Tracker) Get(obj any, pathElement any) (any, error) {
	if obj == nil {
		return nil, fmt.Errorf("cannot navigate nil value")
	}

	rv := reflect.ValueOf(obj)

	// Dereference pointers
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, fmt.Errorf("cannot navigate nil pointer")
		}
		rv = rv.Elem()
	}

	switch pe := pathElement.(type) {
	case string:
		return t.getByString(rv, pe)
	case int:
		return t.getByIndex(rv, pe)
	default:
		return nil, fmt.Errorf("unsupported path element type: %T", pathElement)
	}
}

func (t *Tracker) getByString(rv reflect.Value, name string) (any, error) {
	// Check for method call
	if strings.HasSuffix(name, "()") {
		methodName := strings.TrimSuffix(name, "()")
		return t.callMethod(rv, methodName)
	}

	switch rv.Kind() {
	case reflect.Struct:
		field := rv.FieldByName(name)
		if !field.IsValid() {
			return nil, fmt.Errorf("field %q not found", name)
		}
		if !field.CanInterface() {
			return nil, fmt.Errorf("field %q is unexported", name)
		}
		return field.Interface(), nil

	case reflect.Map:
		key := reflect.ValueOf(name)
		if !key.Type().AssignableTo(rv.Type().Key()) {
			return nil, fmt.Errorf("key type mismatch")
		}
		val := rv.MapIndex(key)
		if !val.IsValid() {
			return nil, fmt.Errorf("key %q not found", name)
		}
		return val.Interface(), nil

	default:
		return nil, fmt.Errorf("cannot get property %q from %s", name, rv.Kind())
	}
}

func (t *Tracker) getByIndex(rv reflect.Value, index int) (any, error) {
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		if index < 0 || index >= rv.Len() {
			return nil, fmt.Errorf("index %d out of bounds (len=%d)", index, rv.Len())
		}
		return rv.Index(index).Interface(), nil

	default:
		return nil, fmt.Errorf("cannot index %s", rv.Kind())
	}
}

func (t *Tracker) callMethod(rv reflect.Value, methodName string) (any, error) {
	// Try on the value first
	method := rv.MethodByName(methodName)
	if !method.IsValid() {
		// Try on pointer to value
		if rv.CanAddr() {
			method = rv.Addr().MethodByName(methodName)
		}
	}

	if !method.IsValid() {
		return nil, fmt.Errorf("method %q not found", methodName)
	}

	mt := method.Type()
	if mt.NumIn() != 0 {
		return nil, fmt.Errorf("method %q requires arguments", methodName)
	}
	if mt.NumOut() == 0 {
		return nil, fmt.Errorf("method %q returns no values", methodName)
	}

	results := method.Call(nil)
	return results[0].Interface(), nil
}

// Set implements the Resolver interface using reflection.
// Sequence: seq-set-value.md
func (t *Tracker) Set(obj any, pathElement any, value any) error {
	if obj == nil {
		return fmt.Errorf("cannot set on nil value")
	}

	rv := reflect.ValueOf(obj)

	// For struct fields, we need a pointer
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	switch pe := pathElement.(type) {
	case string:
		return t.setByString(rv, pe, value)
	case int:
		return t.setByIndex(rv, pe, value)
	default:
		return fmt.Errorf("unsupported path element type: %T", pathElement)
	}
}

func (t *Tracker) setByString(rv reflect.Value, name string, value any) error {
	switch rv.Kind() {
	case reflect.Struct:
		field := rv.FieldByName(name)
		if !field.IsValid() {
			return fmt.Errorf("field %q not found", name)
		}
		if !field.CanSet() {
			return fmt.Errorf("field %q is not settable", name)
		}
		val := reflect.ValueOf(value)
		if !val.Type().AssignableTo(field.Type()) {
			return fmt.Errorf("type mismatch: cannot assign %s to %s", val.Type(), field.Type())
		}
		field.Set(val)
		return nil

	case reflect.Map:
		key := reflect.ValueOf(name)
		if !key.Type().AssignableTo(rv.Type().Key()) {
			return fmt.Errorf("key type mismatch")
		}
		val := reflect.ValueOf(value)
		if !val.Type().AssignableTo(rv.Type().Elem()) {
			return fmt.Errorf("value type mismatch")
		}
		rv.SetMapIndex(key, val)
		return nil

	default:
		return fmt.Errorf("cannot set property %q on %s", name, rv.Kind())
	}
}

func (t *Tracker) setByIndex(rv reflect.Value, index int, value any) error {
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		if index < 0 || index >= rv.Len() {
			return fmt.Errorf("index %d out of bounds (len=%d)", index, rv.Len())
		}
		elem := rv.Index(index)
		if !elem.CanSet() {
			return fmt.Errorf("element at index %d is not settable", index)
		}
		val := reflect.ValueOf(value)
		if !val.Type().AssignableTo(elem.Type()) {
			return fmt.Errorf("type mismatch: cannot assign %s to %s", val.Type(), elem.Type())
		}
		elem.Set(val)
		return nil

	default:
		return fmt.Errorf("cannot index %s", rv.Kind())
	}
}

// Variable methods

// Get gets the variable's value by navigating from the parent's value using the path.
// Sequence: seq-get-value.md
func (v *Variable) Get() (any, error) {
	// Root variable returns cached value
	if v.ParentID == 0 {
		return v.Value, nil
	}

	// Get parent's value
	parent := v.tracker.GetVariable(v.ParentID)
	if parent == nil {
		return nil, fmt.Errorf("parent variable %d not found", v.ParentID)
	}

	current := parent.Value

	// Apply each path element
	for _, elem := range v.Path {
		val, err := v.tracker.Resolver.Get(current, elem)
		if err != nil {
			return nil, err
		}
		current = val
	}

	return current, nil
}

// Set sets the variable's value by navigating from the parent's value using the path.
// Sequence: seq-set-value.md
func (v *Variable) Set(value any) error {
	if len(v.Path) == 0 {
		// Root or no-path variable: update Value directly
		v.Value = value
		return nil
	}

	// Get parent's value
	parent := v.tracker.GetVariable(v.ParentID)
	if parent == nil {
		return fmt.Errorf("parent variable %d not found", v.ParentID)
	}

	// Navigate to the parent of the target
	current := parent.Value
	for i := 0; i < len(v.Path)-1; i++ {
		val, err := v.tracker.Resolver.Get(current, v.Path[i])
		if err != nil {
			return err
		}
		current = val
	}

	// Set the value at the last path element
	return v.tracker.Resolver.Set(current, v.Path[len(v.Path)-1], value)
}

// Parent returns the parent variable, or nil if this is a root variable.
// CRC: crc-Variable.md
func (v *Variable) Parent() *Variable {
	if v.ParentID == 0 {
		return nil
	}
	return v.tracker.GetVariable(v.ParentID)
}

// SetActive sets whether the variable and its children should be checked for changes.
// CRC: crc-Variable.md
func (v *Variable) SetActive(active bool) {
	v.Active = active
}

// GetProperty returns a property value, or empty string if not set.
// CRC: crc-Variable.md
func (v *Variable) GetProperty(name string) string {
	return v.Properties[name]
}

// GetPropertyPriority returns the priority for a property.
// CRC: crc-Variable.md
func (v *Variable) GetPropertyPriority(name string) Priority {
	if p, ok := v.PropertyPriorities[name]; ok {
		return p
	}
	return PriorityMedium
}

// SetProperty sets a property. Empty value removes the property.
// Sequence: seq-set-property.md
func (v *Variable) SetProperty(name, value string) {
	// Parse priority suffix from name
	baseName, priority := parsePropertyName(name)

	if value == "" {
		delete(v.Properties, baseName)
		delete(v.PropertyPriorities, baseName)
	} else {
		v.Properties[baseName] = value
		v.PropertyPriorities[baseName] = priority
	}

	// Record property change in tracker
	v.tracker.recordPropertyChange(v.ID, baseName)

	// Handle special properties
	switch baseName {
	case "path":
		v.Path = parsePath(value)
	case "priority":
		v.ValuePriority = ParsePriority(value)
	}
}

// parsePropertyName extracts the base name and priority from a property name.
// Example: "label:high" -> ("label", PriorityHigh)
func parsePropertyName(name string) (string, Priority) {
	if idx := strings.LastIndex(name, ":"); idx != -1 {
		suffix := name[idx+1:]
		switch strings.ToLower(suffix) {
		case "low":
			return name[:idx], PriorityLow
		case "high":
			return name[:idx], PriorityHigh
		case "medium":
			return name[:idx], PriorityMedium
		}
	}
	return name, PriorityMedium
}
