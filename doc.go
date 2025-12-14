// Package changetracker provides variable management with automatic change detection.
//
// The package provides:
//   - A change tracker that manages variables and detects changes
//   - Variables that hold values and track parent-child relationships
//   - Object registry with weak references for consistent object identity
//   - Value JSON serialization with object references
//   - Change detection via value comparison
//   - Pluggable value resolution for navigating into objects
//
// # Basic Usage
//
// Create a tracker and register variables:
//
//	tracker := changetracker.NewTracker()
//	data := &MyData{Count: 42}
//	root := tracker.CreateVariable(data, 0, "", nil)
//
//	// Create a child variable for a field
//	countVar := tracker.CreateVariable(nil, root.ID, "Count", nil)
//
//	// Modify value externally
//	data.Count = 100
//
//	// Detect changes
//	changes := tracker.DetectChanges()
//
// # Path Navigation
//
// Variables use dot-separated paths to navigate into values:
//
//	// Navigate to nested fields
//	cityVar := tracker.CreateVariable(nil, root.ID, "Address.City", nil)
//
//	// Use method calls in paths (requires appropriate access mode)
//	nameVar := tracker.CreateVariable(nil, root.ID, "GetName()?access=r", nil)
//
//	// Use setter methods (requires access "w" or "action")
//	setterVar := tracker.CreateVariable(nil, root.ID, "SetValue(_)?access=w", nil)
//
// # Access Modes
//
// Variables support four access modes that control read/write permissions:
//
//	| Mode   | Get | Set | Change Detection | Initial Value |
//	|--------|-----|-----|------------------|---------------|
//	| rw     | OK  | OK  | Yes              | Computed      |
//	| r      | OK  | Err | Yes              | Computed      |
//	| w      | Err | OK  | No               | Computed      |
//	| action | Err | OK  | No               | Skipped       |
//
// Path restrictions apply based on access mode:
//   - Paths ending in () require access "r" or "action"
//   - Paths ending in (_) require access "w" or "action"
//
// The "action" mode is designed for variables that trigger side effects,
// where computing the initial value would invoke the action prematurely.
//
// # Priority System
//
// Values and properties can have priority levels (Low, Medium, High).
// Changes are returned sorted by priority (high first):
//
//	// Set priority via path query
//	v := tracker.CreateVariable(nil, root.ID, "Count?priority=high", nil)
//
//	// Set property with priority suffix
//	v.SetProperty("label:high", "Important")
//
// # Object Registry
//
// The tracker maintains a weak map from Go objects (pointers/maps) to variable IDs.
// This enables consistent object identity in Value JSON serialization:
//
//	alice := &Person{Name: "Alice"}
//	tracker.CreateVariable(alice, 0, "", nil)  // ID 1
//
//	// Serialize to Value JSON - registered objects become {"obj": id}
//	json := tracker.ToValueJSON(alice)  // {"obj": 1}
//
// # Change Detection
//
// DetectChanges performs a depth-first traversal from root variables,
// comparing current values to cached Value JSON:
//
//	changes := tracker.DetectChanges()
//	for _, change := range changes {
//	    fmt.Printf("Variable %d changed: value=%v props=%v\n",
//	        change.VariableID, change.ValueChanged, change.PropertiesChanged)
//	}
//
// Active/inactive variables control which subtrees participate in detection.
// Non-readable variables (access "w" or "action") are skipped.
package changetracker
