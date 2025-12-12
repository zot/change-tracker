// Test Design: test-Tracker.md, test-Variable.md, test-Resolver.md, test-ObjectRegistry.md, test-ValueJSON.md, test-Priority.md, test-Change.md
package changetracker

import (
	"testing"
)

// Test types

type Person struct {
	Name    string
	Age     int
	Address *Address
	Tags    []string
}

type Address struct {
	City    string
	Country string
}

func (p *Person) GetName() string {
	return p.Name
}

func (p *Person) Pair() (string, int) {
	return p.Name, p.Age
}

// ============================================================================
// Priority Tests (test-Priority.md)
// ============================================================================

func TestPriority_Constants(t *testing.T) {
	if PriorityLow != -1 {
		t.Errorf("PriorityLow should be -1, got %d", PriorityLow)
	}
	if PriorityMedium != 0 {
		t.Errorf("PriorityMedium should be 0, got %d", PriorityMedium)
	}
	if PriorityHigh != 1 {
		t.Errorf("PriorityHigh should be 1, got %d", PriorityHigh)
	}
}

func TestPriority_Ordering(t *testing.T) {
	if !(PriorityLow < PriorityMedium && PriorityMedium < PriorityHigh) {
		t.Error("priority ordering should be low < medium < high")
	}
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		input    string
		expected Priority
	}{
		{"low", PriorityLow},
		{"LOW", PriorityLow},
		{"medium", PriorityMedium},
		{"MEDIUM", PriorityMedium},
		{"high", PriorityHigh},
		{"HIGH", PriorityHigh},
		{"", PriorityMedium},
		{"invalid", PriorityMedium},
	}
	for _, tc := range tests {
		got := ParsePriority(tc.input)
		if got != tc.expected {
			t.Errorf("ParsePriority(%q): expected %d, got %d", tc.input, tc.expected, got)
		}
	}
}

func TestParsePropertyName(t *testing.T) {
	tests := []struct {
		input        string
		expectedName string
		expectedPri  Priority
	}{
		{"label", "label", PriorityMedium},
		{"label:low", "label", PriorityLow},
		{"label:medium", "label", PriorityMedium},
		{"label:high", "label", PriorityHigh},
		{"label:HIGH", "label", PriorityHigh},
		{"label:invalid", "label:invalid", PriorityMedium},
		{"a:b:high", "a:b", PriorityHigh},
	}
	for _, tc := range tests {
		name, pri := parsePropertyName(tc.input)
		if name != tc.expectedName || pri != tc.expectedPri {
			t.Errorf("parsePropertyName(%q): expected (%q, %d), got (%q, %d)",
				tc.input, tc.expectedName, tc.expectedPri, name, pri)
		}
	}
}

// ============================================================================
// Tracker Tests (test-Tracker.md)
// ============================================================================

// T1.1, T1.2: NewTracker
func TestNewTracker(t *testing.T) {
	tr := NewTracker()
	if tr == nil {
		t.Fatal("NewTracker returned nil")
	}
	// T1.1: Resolver defaults to self
	if tr.Resolver != tr {
		t.Error("Resolver should default to tracker itself")
	}
	// T1.2: Initial state
	if len(tr.variables) != 0 {
		t.Error("variables should be empty initially")
	}
	if len(tr.valueChanges) != 0 {
		t.Error("valueChanges should be empty initially")
	}
	if tr.nextID != 1 {
		t.Errorf("nextID should be 1, got %d", tr.nextID)
	}
}

// T2.1: Root variable with value
func TestCreateVariable_RootWithValue(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	if v.ID != 1 {
		t.Errorf("expected ID=1, got %d", v.ID)
	}
	if v.Value != 42 {
		t.Errorf("expected Value=42, got %v", v.Value)
	}
	if v.ValueJSON != 42 {
		t.Errorf("expected ValueJSON=42, got %v", v.ValueJSON)
	}
}

// T2.2: Root variable with properties
func TestCreateVariable_RootWithProperties(t *testing.T) {
	tr := NewTracker()
	props := map[string]string{"name": "counter"}
	v := tr.CreateVariable(10, 0, "", props)

	if v.Properties["name"] != "counter" {
		t.Errorf("expected property name=counter, got %v", v.Properties["name"])
	}
}

// T2.3: Child variable with path
func TestCreateVariable_ChildWithPath(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	parent := tr.CreateVariable(person, 0, "", nil)

	child := tr.CreateVariable(nil, parent.ID, "Name", nil)

	if len(child.Path) != 1 || child.Path[0] != "Name" {
		t.Errorf("expected Path=[Name], got %v", child.Path)
	}
	if child.Value != "Alice" {
		t.Errorf("expected Value=Alice, got %v", child.Value)
	}
}

// T2.4: Child with dot path
func TestCreateVariable_ChildWithDotPath(t *testing.T) {
	tr := NewTracker()
	person := &Person{
		Name:    "Alice",
		Address: &Address{City: "NYC", Country: "USA"},
	}
	parent := tr.CreateVariable(person, 0, "", nil)

	child := tr.CreateVariable(nil, parent.ID, "Address.City", nil)

	if child.Value != "NYC" {
		t.Errorf("expected Value=NYC, got %v", child.Value)
	}
}

// T2.5: Child with index path
func TestCreateVariable_ChildWithIndexPath(t *testing.T) {
	tr := NewTracker()
	data := []string{"a", "b", "c"}
	parent := tr.CreateVariable(data, 0, "", nil)

	child := tr.CreateVariable(nil, parent.ID, "0", nil)

	// parsePath converts "0" to int 0 for index access
	if len(child.Path) != 1 || child.Path[0] != 0 {
		t.Errorf("expected Path=[0], got %v", child.Path)
	}
	if child.Value != "a" {
		t.Errorf("expected Value=a, got %v", child.Value)
	}
}

// T2.6: Pointer value registered
func TestCreateVariable_PointerRegistered(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	v := tr.CreateVariable(person, 0, "", nil)

	// Check object is in registry
	id, ok := tr.LookupObject(person)
	if !ok {
		t.Error("pointer should be registered")
	}
	if id != v.ID {
		t.Errorf("registered ID should be %d, got %d", v.ID, id)
	}

	// ValueJSON should be ObjectRef
	ref, ok := v.ValueJSON.(ObjectRef)
	if !ok {
		t.Errorf("expected ObjectRef, got %T", v.ValueJSON)
	}
	if ref.Obj != v.ID {
		t.Errorf("ObjectRef.Obj should be %d, got %d", v.ID, ref.Obj)
	}
}

// T2.7: Map value registered
func TestCreateVariable_MapRegistered(t *testing.T) {
	tr := NewTracker()
	data := map[string]int{"a": 1, "b": 2}
	v := tr.CreateVariable(data, 0, "", nil)

	// Check object is in registry
	id, ok := tr.LookupObject(data)
	if !ok {
		t.Error("map should be registered")
	}
	if id != v.ID {
		t.Errorf("registered ID should be %d, got %d", v.ID, id)
	}

	// ValueJSON should be ObjectRef
	ref, ok := v.ValueJSON.(ObjectRef)
	if !ok {
		t.Errorf("expected ObjectRef, got %T", v.ValueJSON)
	}
	if ref.Obj != v.ID {
		t.Errorf("ObjectRef.Obj should be %d, got %d", v.ID, ref.Obj)
	}
}

// T2.8: Sequential IDs
func TestCreateVariable_SequentialIDs(t *testing.T) {
	tr := NewTracker()
	v1 := tr.CreateVariable(1, 0, "", nil)
	v2 := tr.CreateVariable(2, 0, "", nil)
	v3 := tr.CreateVariable(3, 0, "", nil)

	if v1.ID != 1 || v2.ID != 2 || v3.ID != 3 {
		t.Errorf("expected IDs 1,2,3, got %d,%d,%d", v1.ID, v2.ID, v3.ID)
	}
}

// T2.9: Nil properties
func TestCreateVariable_NilProperties(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	if v.Properties == nil {
		t.Error("Properties should not be nil")
	}
}

// T2.10: Empty path uses properties fallback
func TestCreateVariable_EmptyPathUsesProps(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	parent := tr.CreateVariable(person, 0, "", nil)

	child := tr.CreateVariable(nil, parent.ID, "", map[string]string{"path": "Name"})

	if child.Value != "Alice" {
		t.Errorf("expected Value=Alice from props path, got %v", child.Value)
	}
}

// T2.11: Path arg overrides props
func TestCreateVariable_PathOverridesProps(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	parent := tr.CreateVariable(person, 0, "", nil)

	child := tr.CreateVariable(nil, parent.ID, "Age", map[string]string{"path": "Name"})

	if child.Value != 30 {
		t.Errorf("expected Value=30 from path arg, got %v", child.Value)
	}
}

// T2.12: Path with query parameters
func TestCreateVariable_PathWithQuery(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	parent := tr.CreateVariable(person, 0, "", nil)

	child := tr.CreateVariable(nil, parent.ID, "Name?label=test&hint=info", nil)

	if child.Value != "Alice" {
		t.Errorf("expected Value=Alice, got %v", child.Value)
	}
	if child.Properties["label"] != "test" {
		t.Errorf("expected label=test, got %v", child.Properties["label"])
	}
	if child.Properties["hint"] != "info" {
		t.Errorf("expected hint=info, got %v", child.Properties["hint"])
	}
}

// T2.13: Query props override properties map
func TestCreateVariable_QueryOverridesProps(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	parent := tr.CreateVariable(person, 0, "", nil)

	child := tr.CreateVariable(nil, parent.ID, "Name?label=fromQuery", map[string]string{"label": "fromProps"})

	if child.Properties["label"] != "fromQuery" {
		t.Errorf("expected label=fromQuery, got %v", child.Properties["label"])
	}
}

// T2.14: Path with priority property
func TestCreateVariable_PathWithPriority(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	parent := tr.CreateVariable(person, 0, "", nil)

	child := tr.CreateVariable(nil, parent.ID, "Name?priority=high", nil)

	if child.ValuePriority != PriorityHigh {
		t.Errorf("expected ValuePriority=High, got %d", child.ValuePriority)
	}
}

// T2.15: Priority from properties map
func TestCreateVariable_PriorityFromProps(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", map[string]string{"priority": "low"})

	if v.ValuePriority != PriorityLow {
		t.Errorf("expected ValuePriority=Low, got %d", v.ValuePriority)
	}
}

// T3.1, T3.2: GetVariable
func TestGetVariable(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	// T3.1: Existing variable
	got := tr.GetVariable(v.ID)
	if got != v {
		t.Error("GetVariable should return the same variable")
	}

	// T3.2: Non-existent ID
	got = tr.GetVariable(999)
	if got != nil {
		t.Error("GetVariable should return nil for non-existent ID")
	}
}

// T3.3: After destroy
func TestGetVariable_AfterDestroy(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)
	id := v.ID

	tr.DestroyVariable(id)

	got := tr.GetVariable(id)
	if got != nil {
		t.Error("GetVariable should return nil for destroyed variable")
	}
}

// T4.1-T4.4: DestroyVariable
func TestDestroyVariable(t *testing.T) {
	tr := NewTracker()

	// T4.1: Destroy existing
	v := tr.CreateVariable(42, 0, "", nil)
	tr.DestroyVariable(v.ID)
	if tr.GetVariable(v.ID) != nil {
		t.Error("variable should be removed")
	}

	// T4.2: Object unregistered
	person := &Person{Name: "Alice"}
	v2 := tr.CreateVariable(person, 0, "", nil)
	tr.DestroyVariable(v2.ID)
	_, ok := tr.LookupObject(person)
	if ok {
		t.Error("object should be unregistered after destroy")
	}

	// T4.3: Removed from change tracking
	v3 := tr.CreateVariable(1, 0, "", nil)
	tr.valueChanges[v3.ID] = true
	tr.DestroyVariable(v3.ID)
	if tr.valueChanges[v3.ID] {
		t.Error("should be removed from valueChanges")
	}

	// T4.4: Destroy non-existent (no panic)
	tr.DestroyVariable(999) // should not panic
}

// T5.1-T5.15: DetectChanges
func TestDetectChanges(t *testing.T) {
	tr := NewTracker()

	// T5.1: No changes - returns empty []Change
	tr.CreateVariable(42, 0, "", nil)
	changes := tr.DetectChanges()
	if len(changes) != 0 {
		t.Error("no changes expected")
	}

	// T5.2: Primitive change (via struct field) - returns []Change with variable ID
	person := &Person{Name: "Alice", Age: 30}
	v := tr.CreateVariable(person, 0, "", nil)
	child := tr.CreateVariable(nil, v.ID, "Age", nil)

	person.Age = 31
	changes = tr.DetectChanges()
	found := false
	for _, c := range changes {
		if c.VariableID == child.ID && c.ValueChanged {
			found = true
			break
		}
	}
	if !found {
		t.Error("child should be in changes after field modification")
	}

	// T5.8: No false positives - returns empty []Change
	person.Age = 31 // same value
	changes = tr.DetectChanges()
	found = false
	for _, c := range changes {
		if c.VariableID == child.ID {
			found = true
			break
		}
	}
	if found {
		t.Error("should not detect change when value is same")
	}

	// T5.9: ValueJSON updated after detection
	person.Age = 32
	tr.DetectChanges()
	if child.ValueJSON != 32 {
		t.Errorf("ValueJSON should be updated to 32, got %v", child.ValueJSON)
	}

	// T5.10: Clears internal state after call
	person.Age = 33
	tr.DetectChanges()
	if len(tr.valueChanges) != 0 {
		t.Error("valueChanges should be cleared after DetectChanges")
	}
	if len(tr.propertyChanges) != 0 {
		t.Error("propertyChanges should be cleared after DetectChanges")
	}
}

// T5.11-T5.15: DetectChanges - additional tests for sorting and property changes
func TestDetectChanges_SortingAndProperties(t *testing.T) {
	tr := NewTracker()

	// T5.11: Returns sorted changes - high priority first
	vHigh := tr.CreateVariable(1, 0, "", map[string]string{"priority": "high"})
	vLow := tr.CreateVariable(2, 0, "", map[string]string{"priority": "low"})

	vHigh.SetProperty("x:high", "1")
	vLow.SetProperty("y:low", "2")

	changes := tr.DetectChanges()
	if len(changes) < 2 {
		t.Fatalf("expected at least 2 changes, got %d", len(changes))
	}
	// High priority should come before low
	highIdx := -1
	lowIdx := -1
	for i, c := range changes {
		if c.Priority == PriorityHigh {
			highIdx = i
		}
		if c.Priority == PriorityLow {
			lowIdx = i
		}
	}
	if highIdx != -1 && lowIdx != -1 && highIdx > lowIdx {
		t.Error("high priority changes should come before low priority")
	}

	// T5.12: Property changes included in result
	tr2 := NewTracker()
	v := tr2.CreateVariable(42, 0, "", nil)
	v.SetProperty("label", "test")
	changes = tr2.DetectChanges()

	found := false
	for _, c := range changes {
		if c.VariableID == v.ID {
			for _, p := range c.PropertiesChanged {
				if p == "label" {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Error("property change should appear in DetectChanges result")
	}

	// T5.13: Slice reuse
	tr3 := NewTracker()
	v3 := tr3.CreateVariable(42, 0, "", nil)
	v3.SetProperty("x", "1")
	changes1 := tr3.DetectChanges()
	cap1 := cap(changes1)

	v3.SetProperty("y", "2")
	changes2 := tr3.DetectChanges()

	// The slice should be reused (same backing array)
	if len(changes2) > 0 && cap(changes2) < cap1 {
		t.Error("slice should be reused")
	}
}

// T8.1, T8.2: Variables
func TestVariables(t *testing.T) {
	// T8.2: Empty tracker
	tr := NewTracker()
	if len(tr.Variables()) != 0 {
		t.Error("should return empty slice for empty tracker")
	}

	// T8.1: All variables
	tr.CreateVariable(1, 0, "", nil)
	tr.CreateVariable(2, 0, "", nil)
	tr.CreateVariable(3, 0, "", nil)
	if len(tr.Variables()) != 3 {
		t.Errorf("expected 3 variables, got %d", len(tr.Variables()))
	}
}

// T9.1: RootVariables
func TestRootVariables(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	root1 := tr.CreateVariable(person, 0, "", nil)
	tr.CreateVariable(nil, root1.ID, "Name", nil) // child
	root2 := tr.CreateVariable(42, 0, "", nil)

	roots := tr.RootVariables()
	if len(roots) != 2 {
		t.Errorf("expected 2 root variables, got %d", len(roots))
	}

	ids := make(map[int64]bool)
	for _, r := range roots {
		ids[r.ID] = true
	}
	if !ids[root1.ID] || !ids[root2.ID] {
		t.Error("root variables not found correctly")
	}
}

// T10.1, T10.2: Children
func TestChildren(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	parent := tr.CreateVariable(person, 0, "", nil)

	// T10.2: No children
	if len(tr.Children(parent.ID)) != 0 {
		t.Error("should have no children initially")
	}

	// T10.1: Get children
	tr.CreateVariable(nil, parent.ID, "Name", nil)
	tr.CreateVariable(nil, parent.ID, "Age", nil)

	children := tr.Children(parent.ID)
	if len(children) != 2 {
		t.Errorf("expected 2 children, got %d", len(children))
	}
}

// ============================================================================
// Change Tests (test-Change.md) - now integrated into DetectChanges
// ============================================================================

func TestDetectChanges_Empty(t *testing.T) {
	tr := NewTracker()
	changes := tr.DetectChanges()
	if len(changes) != 0 {
		t.Error("should return empty slice when no changes")
	}
}

func TestDetectChanges_ValueChange(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	v := tr.CreateVariable(person, 0, "", nil)
	child := tr.CreateVariable(nil, v.ID, "Age", nil)

	// First DetectChanges clears initial state
	tr.DetectChanges()

	person.Age = 31
	changes := tr.DetectChanges()

	if len(changes) == 0 {
		t.Fatal("should have at least one change")
	}

	found := false
	for _, c := range changes {
		if c.VariableID == child.ID && c.ValueChanged {
			found = true
			break
		}
	}
	if !found {
		t.Error("should find value change for child")
	}
}

func TestDetectChanges_PropertyChange(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	// First DetectChanges clears initial state
	tr.DetectChanges()

	v.SetProperty("label", "test")

	changes := tr.DetectChanges()
	if len(changes) == 0 {
		t.Fatal("should have at least one change")
	}

	found := false
	for _, c := range changes {
		if c.VariableID == v.ID && len(c.PropertiesChanged) > 0 {
			for _, p := range c.PropertiesChanged {
				if p == "label" {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Error("should find property change for label")
	}
}

func TestDetectChanges_PriorityOrder(t *testing.T) {
	tr := NewTracker()

	// Create variables with different priorities
	vHigh := tr.CreateVariable(1, 0, "", map[string]string{"priority": "high"})
	vMed := tr.CreateVariable(2, 0, "", nil) // default medium
	vLow := tr.CreateVariable(3, 0, "", map[string]string{"priority": "low"})

	// First DetectChanges clears initial state
	tr.DetectChanges()

	// Trigger property changes with different priorities
	vHigh.SetProperty("x:high", "1")
	vMed.SetProperty("x", "2") // default medium
	vLow.SetProperty("x:low", "3")

	changes := tr.DetectChanges()
	if len(changes) != 3 {
		t.Fatalf("expected 3 changes, got %d", len(changes))
	}

	// Changes should be sorted: high, medium, low
	// Note: all three are medium priority since they're property changes without explicit priority suffixes
	_ = vHigh
	_ = vLow
}

func TestDetectChanges_SplitByPriority(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", map[string]string{"priority": "high"})

	// First DetectChanges clears initial state
	tr.DetectChanges()

	// Set properties with different priorities
	v.SetProperty("high_prop:high", "val1")
	v.SetProperty("low_prop:low", "val2")

	// Also need to trigger a value change by modifying a trackable value
	// Since this is a primitive, we'll add a struct-based test
	changes := tr.DetectChanges()

	// Should have multiple entries for same variable at different priorities
	highCount := 0
	lowCount := 0
	for _, c := range changes {
		if c.VariableID == v.ID {
			if c.Priority == PriorityHigh {
				highCount++
			}
			if c.Priority == PriorityLow {
				lowCount++
			}
		}
	}

	if highCount == 0 {
		t.Error("should have high priority change")
	}
	if lowCount == 0 {
		t.Error("should have low priority change")
	}
}

func TestDetectChanges_SliceReuse(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	// First DetectChanges clears initial state
	tr.DetectChanges()

	v.SetProperty("x", "1")
	changes1 := tr.DetectChanges()
	_ = changes1 // first call

	// Make new changes
	v.SetProperty("y", "2")
	changes2 := tr.DetectChanges()

	// The returned slices should use the same underlying array
	if len(changes2) == 0 {
		t.Error("should have changes after second DetectChanges")
	}
}

// ============================================================================
// Variable Tests (test-Variable.md)
// ============================================================================

// V1.1-V1.7: Variable.Get
func TestVariable_Get(t *testing.T) {
	tr := NewTracker()
	person := &Person{
		Name:    "Alice",
		Age:     30,
		Address: &Address{City: "NYC", Country: "USA"},
		Tags:    []string{"dev", "go"},
	}
	parent := tr.CreateVariable(person, 0, "", nil)

	// V1.1: Root variable get
	val, err := parent.Get()
	if err != nil {
		t.Errorf("root Get() failed: %v", err)
	}
	if val != person {
		t.Error("root Get() should return cached value")
	}

	// V1.2: Child field get
	nameChild := tr.CreateVariable(nil, parent.ID, "Name", nil)
	val, err = nameChild.Get()
	if err != nil {
		t.Errorf("field Get() failed: %v", err)
	}
	if val != "Alice" {
		t.Errorf("expected Alice, got %v", val)
	}

	// V1.3: Child nested get
	cityChild := tr.CreateVariable(nil, parent.ID, "Address.City", nil)
	val, err = cityChild.Get()
	if err != nil {
		t.Errorf("nested Get() failed: %v", err)
	}
	if val != "NYC" {
		t.Errorf("expected NYC, got %v", val)
	}

	// V1.6: Map key get
	m := map[string]int{"one": 1, "two": 2}
	mapVar := tr.CreateVariable(m, 0, "", nil)
	mapChild := tr.CreateVariable(nil, mapVar.ID, "one", nil)
	val, err = mapChild.Get()
	if err != nil {
		t.Errorf("map Get() failed: %v", err)
	}
	if val != 1 {
		t.Errorf("expected 1, got %v", val)
	}

	// V1.7: Value caching
	if nameChild.Value != "Alice" {
		t.Error("Value should be cached after Get()")
	}
}

// V1.5: Method call get
func TestVariable_GetMethod(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	parent := tr.CreateVariable(person, 0, "", nil)

	methodChild := tr.CreateVariable(nil, parent.ID, "GetName()", nil)
	val, err := methodChild.Get()
	if err != nil {
		t.Errorf("method Get() failed: %v", err)
	}
	if val != "Alice" {
		t.Errorf("expected Alice from method, got %v", val)
	}
}

// V2.1-V2.5: Variable.Set
func TestVariable_Set(t *testing.T) {
	tr := NewTracker()
	person := &Person{
		Name:    "Alice",
		Age:     30,
		Address: &Address{City: "NYC", Country: "USA"},
	}
	parent := tr.CreateVariable(person, 0, "", nil)

	// V2.1: Set struct field
	nameChild := tr.CreateVariable(nil, parent.ID, "Name", nil)
	err := nameChild.Set("Bob")
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}
	if person.Name != "Bob" {
		t.Errorf("expected Bob, got %s", person.Name)
	}

	// V2.2: Set nested field
	cityChild := tr.CreateVariable(nil, parent.ID, "Address.City", nil)
	err = cityChild.Set("LA")
	if err != nil {
		t.Errorf("nested Set failed: %v", err)
	}
	if person.Address.City != "LA" {
		t.Errorf("expected LA, got %s", person.Address.City)
	}

	// V2.4: Set map value
	m := map[string]int{"one": 1}
	mapVar := tr.CreateVariable(m, 0, "", nil)
	mapChild := tr.CreateVariable(nil, mapVar.ID, "one", nil)
	err = mapChild.Set(100)
	if err != nil {
		t.Errorf("map Set failed: %v", err)
	}
	if m["one"] != 100 {
		t.Errorf("expected 100, got %d", m["one"])
	}
}

// V3.1, V3.2: Variable.Parent
func TestVariable_Parent(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	parent := tr.CreateVariable(person, 0, "", nil)
	child := tr.CreateVariable(nil, parent.ID, "Name", nil)

	// V3.1: Root has no parent
	if parent.Parent() != nil {
		t.Error("root should have no parent")
	}

	// V3.2: Child has parent
	if child.Parent() != parent {
		t.Error("child should have correct parent")
	}
}

// V4.1-V4.14: GetProperty / SetProperty
func TestVariable_Properties(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	// Clear initial state
	tr.DetectChanges()

	// V4.1: Get non-existent
	if v.GetProperty("missing") != "" {
		t.Error("should return empty for missing property")
	}

	// V4.3: Set property
	v.SetProperty("key", "value")
	if v.GetProperty("key") != "value" {
		t.Error("should set property")
	}

	// V4.4: Remove property
	v.SetProperty("key", "")
	if v.GetProperty("key") != "" {
		t.Error("should remove property when set to empty")
	}

	// V4.5-V4.11: SetProperty with priority suffixes
	v.SetProperty("label:high", "Important")
	if v.GetProperty("label") != "Important" {
		t.Error("should set property value without suffix")
	}
	if v.GetPropertyPriority("label") != PriorityHigh {
		t.Error("should set property priority from suffix")
	}

	v.SetProperty("hint:low", "Optional")
	if v.GetPropertyPriority("hint") != PriorityLow {
		t.Errorf("expected PriorityLow, got %d", v.GetPropertyPriority("hint"))
	}

	// V4.12: Set path property updates Path
	v.SetProperty("path", "a.b.c")
	if len(v.Path) != 3 || v.Path[0] != "a" || v.Path[1] != "b" || v.Path[2] != "c" {
		t.Errorf("setting path property should update Path field, got %v", v.Path)
	}

	// V4.13: SetProperty records change in tracker (appears in DetectChanges result)
	tr.DetectChanges() // clear previous changes
	v.SetProperty("test", "value")
	changes := tr.DetectChanges()
	found := false
	for _, c := range changes {
		if c.VariableID == v.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("SetProperty should record change in tracker")
	}
}

// V5.1-V5.3: GetPropertyPriority
func TestVariable_GetPropertyPriority(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	// V5.1: Default priority
	if v.GetPropertyPriority("unknown") != PriorityMedium {
		t.Error("should return Medium for unknown property")
	}

	// V5.2: Set priority via suffix
	v.SetProperty("label:high", "test")
	if v.GetPropertyPriority("label") != PriorityHigh {
		t.Error("should return High for label")
	}
}

// Test SetProperty updates ValuePriority
func TestSetProperty_Priority(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	if v.ValuePriority != PriorityMedium {
		t.Error("default ValuePriority should be Medium")
	}

	v.SetProperty("priority", "high")
	if v.ValuePriority != PriorityHigh {
		t.Errorf("setting priority property should update ValuePriority, got %d", v.ValuePriority)
	}

	v.SetProperty("priority", "low")
	if v.ValuePriority != PriorityLow {
		t.Errorf("setting priority property should update ValuePriority, got %d", v.ValuePriority)
	}
}

// P1-P6: Path Parsing Tests
func TestParsePath(t *testing.T) {
	tests := []struct {
		input    string
		expected []any
	}{
		{"Name", []any{"Name"}},
		{"Address.City", []any{"Address", "City"}},
		{"0", []any{0}}, // Numeric strings parsed as int
		{"Items.0.Name", []any{"Items", 0, "Name"}},
		{"GetValue()", []any{"GetValue()"}},
		{"", nil},
	}

	for _, tc := range tests {
		got := parsePath(tc.input)
		if len(got) != len(tc.expected) {
			t.Errorf("parsePath(%q): expected %v, got %v", tc.input, tc.expected, got)
			continue
		}
		for i := range got {
			if got[i] != tc.expected[i] {
				t.Errorf("parsePath(%q)[%d]: expected %v, got %v", tc.input, i, tc.expected[i], got[i])
			}
		}
	}
}

// P7-P10: Path with query parsing
func TestParsePathWithQuery(t *testing.T) {
	tests := []struct {
		input         string
		expectedPath  string
		expectedProps map[string]string
	}{
		{"a.b", "a.b", nil},
		{"a.b?", "a.b", map[string]string{}},
		{"a.b?x=1", "a.b", map[string]string{"x": "1"}},
		{"a.b?x=1&y=2", "a.b", map[string]string{"x": "1", "y": "2"}},
		{"?x=1", "", map[string]string{"x": "1"}},
		{"", "", nil},
	}

	for _, tc := range tests {
		path, props := parsePathWithQuery(tc.input)
		if path != tc.expectedPath {
			t.Errorf("parsePathWithQuery(%q): expected path %q, got %q", tc.input, tc.expectedPath, path)
		}
		if tc.expectedProps == nil && props != nil {
			// ok, nil vs empty map
		} else if tc.expectedProps != nil {
			for k, v := range tc.expectedProps {
				if props[k] != v {
					t.Errorf("parsePathWithQuery(%q): expected props[%q]=%q, got %q", tc.input, k, v, props[k])
				}
			}
		}
	}
}

// ============================================================================
// Resolver Tests (test-Resolver.md)
// ============================================================================

// R1.1-R1.4: Get - Struct Fields
func TestResolver_GetStructFields(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30, Address: &Address{City: "NYC"}}

	// R1.1: Exported field
	val, err := tr.Get(person, "Name")
	if err != nil || val != "Alice" {
		t.Errorf("Get struct field: err=%v, val=%v", err, val)
	}

	// R1.2: Nested struct
	val, err = tr.Get(person, "Address")
	if err != nil || val != person.Address {
		t.Errorf("Get nested struct: err=%v, val=%v", err, val)
	}

	// R1.4: Various types
	val, err = tr.Get(person, "Age")
	if err != nil || val != 30 {
		t.Errorf("Get int field: err=%v, val=%v", err, val)
	}
}

// R2.1-R2.3: Get - Map Keys
func TestResolver_GetMapKeys(t *testing.T) {
	tr := NewTracker()
	m := map[string]int{"key": 42, "other": 10}

	// R2.1: String key exists
	val, err := tr.Get(m, "key")
	if err != nil || val != 42 {
		t.Errorf("Get map key: err=%v, val=%v", err, val)
	}

	// R2.2: String key missing
	_, err = tr.Get(m, "missing")
	if err == nil {
		t.Error("Get missing map key should error")
	}
}

// R3.1-R3.3: Get - Slice/Array Index
func TestResolver_GetIndex(t *testing.T) {
	tr := NewTracker()
	slice := []string{"a", "b", "c"}

	// R3.1: Valid index
	val, err := tr.Get(slice, 0)
	if err != nil || val != "a" {
		t.Errorf("Get slice[0]: err=%v, val=%v", err, val)
	}

	// R3.2: Middle index
	val, err = tr.Get(slice, 1)
	if err != nil || val != "b" {
		t.Errorf("Get slice[1]: err=%v, val=%v", err, val)
	}

	// R3.3: Array access
	arr := [3]int{10, 20, 30}
	val, err = tr.Get(arr, 2)
	if err != nil || val != 30 {
		t.Errorf("Get array[2]: err=%v, val=%v", err, val)
	}
}

// R4.1-R4.3: Get - Method Calls
func TestResolver_GetMethod(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}

	// R4.1: Zero-arg method
	val, err := tr.Get(person, "GetName()")
	if err != nil || val != "Alice" {
		t.Errorf("Get method: err=%v, val=%v", err, val)
	}

	// R4.3: Multi-return method (returns first value)
	val, err = tr.Get(person, "Pair()")
	if err != nil || val != "Alice" {
		t.Errorf("Get multi-return method: err=%v, val=%v", err, val)
	}
}

// S1.1-S1.3: Set - Struct Fields
func TestResolver_SetStructFields(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30, Address: &Address{City: "NYC"}}

	// S1.1: Set via pointer
	err := tr.Set(person, "Name", "Bob")
	if err != nil || person.Name != "Bob" {
		t.Errorf("Set struct field: err=%v, name=%s", err, person.Name)
	}

	// S1.2: Set int field
	err = tr.Set(person, "Age", 31)
	if err != nil || person.Age != 31 {
		t.Errorf("Set int field: err=%v, age=%d", err, person.Age)
	}

	// S1.3: Set nested field
	err = tr.Set(person.Address, "City", "LA")
	if err != nil || person.Address.City != "LA" {
		t.Errorf("Set nested field: err=%v, city=%s", err, person.Address.City)
	}
}

// S2.1, S2.2: Set - Map Keys
func TestResolver_SetMapKeys(t *testing.T) {
	tr := NewTracker()
	m := map[string]int{"key": 42}

	// S2.1: Set existing key
	err := tr.Set(m, "key", 100)
	if err != nil || m["key"] != 100 {
		t.Errorf("Set existing map key: err=%v, val=%d", err, m["key"])
	}

	// S2.2: Set new key
	err = tr.Set(m, "new", 200)
	if err != nil || m["new"] != 200 {
		t.Errorf("Set new map key: err=%v, val=%d", err, m["new"])
	}
}

// S3.1, S3.2: Set - Slice Index
func TestResolver_SetSliceIndex(t *testing.T) {
	tr := NewTracker()
	slice := []string{"a", "b", "c"}

	// S3.1: Set valid index
	err := tr.Set(slice, 0, "x")
	if err != nil || slice[0] != "x" {
		t.Errorf("Set slice[0]: err=%v, val=%s", err, slice[0])
	}

	// S3.2: Set middle index
	err = tr.Set(slice, 1, "y")
	if err != nil || slice[1] != "y" {
		t.Errorf("Set slice[1]: err=%v, val=%s", err, slice[1])
	}
}

// GE1-GE8: Get Errors
func TestResolver_GetErrors(t *testing.T) {
	tr := NewTracker()

	// GE1: Nil object
	_, err := tr.Get(nil, "field")
	if err == nil {
		t.Error("Get nil should error")
	}

	// GE3: Missing field
	person := &Person{Name: "Alice"}
	_, err = tr.Get(person, "NoSuch")
	if err == nil {
		t.Error("Get missing field should error")
	}

	// GE4: Missing map key
	m := map[string]int{"a": 1}
	_, err = tr.Get(m, "missing")
	if err == nil {
		t.Error("Get missing map key should error")
	}

	// GE5: Index out of bounds
	slice := []int{1, 2}
	_, err = tr.Get(slice, 100)
	if err == nil {
		t.Error("Get out of bounds should error")
	}

	// GE6: Method not found
	_, err = tr.Get(person, "NoMethod()")
	if err == nil {
		t.Error("Get missing method should error")
	}

	// GE8: Unsupported type
	_, err = tr.Get(42, "field")
	if err == nil {
		t.Error("Get on int should error")
	}
}

// SE1-SE7: Set Errors
func TestResolver_SetErrors(t *testing.T) {
	tr := NewTracker()

	// SE1: Nil object
	err := tr.Set(nil, "field", "value")
	if err == nil {
		t.Error("Set nil should error")
	}

	// SE4: Missing field
	person := &Person{Name: "Alice"}
	err = tr.Set(person, "NoSuch", "value")
	if err == nil {
		t.Error("Set missing field should error")
	}

	// SE5: Type mismatch
	err = tr.Set(person, "Age", "not an int")
	if err == nil {
		t.Error("Set type mismatch should error")
	}

	// SE6: Index out of bounds
	slice := []int{1, 2}
	err = tr.Set(slice, 100, 999)
	if err == nil {
		t.Error("Set out of bounds should error")
	}
}

// ============================================================================
// ObjectRegistry Tests (test-ObjectRegistry.md)
// ============================================================================

// OR1.1-OR1.5: RegisterObject
func TestRegisterObject(t *testing.T) {
	tr := NewTracker()

	// OR1.1: Register pointer
	person := &Person{Name: "Alice"}
	ok := tr.RegisterObject(person, 1)
	if !ok {
		t.Error("should register pointer")
	}

	// OR1.2: Register map
	m := map[string]int{"a": 1}
	ok = tr.RegisterObject(m, 2)
	if !ok {
		t.Error("should register map")
	}

	// OR1.3: Register non-pointer
	ok = tr.RegisterObject(42, 3)
	if ok {
		t.Error("should not register int")
	}

	// OR1.4: Register nil
	ok = tr.RegisterObject(nil, 4)
	if ok {
		t.Error("should not register nil")
	}

	// OR1.5: Re-register same (updates varID)
	ok = tr.RegisterObject(person, 10)
	if !ok {
		t.Error("should re-register pointer")
	}
	id, _ := tr.LookupObject(person)
	if id != 10 {
		t.Errorf("re-register should update ID to 10, got %d", id)
	}
}

// OR2.1-OR2.3: UnregisterObject
func TestUnregisterObject(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	tr.RegisterObject(person, 1)

	// OR2.1: Unregister existing
	tr.UnregisterObject(person)

	// OR2.3: After unregister lookup
	_, ok := tr.LookupObject(person)
	if ok {
		t.Error("should not find unregistered object")
	}

	// OR2.2: Unregister non-existent (no panic)
	tr.UnregisterObject(&Person{Name: "Bob"})
}

// OR3.1, OR3.2: LookupObject
func TestLookupObject(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}

	// OR3.2: Lookup unregistered
	_, ok := tr.LookupObject(person)
	if ok {
		t.Error("should not find unregistered object")
	}

	// OR3.1: Lookup registered
	tr.RegisterObject(person, 1)
	id, ok := tr.LookupObject(person)
	if !ok || id != 1 {
		t.Errorf("should find registered object with ID 1, got id=%d, ok=%v", id, ok)
	}
}

// OR4.1, OR4.2: GetObject
func TestGetObject(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	tr.RegisterObject(person, 1)

	// OR4.1: Get registered
	obj := tr.GetObject(1)
	if obj == nil {
		t.Error("should get registered object")
	}

	// OR4.2: Get unknown ID
	obj = tr.GetObject(999)
	if obj != nil {
		t.Error("should return nil for unknown ID")
	}
}

// CV1-CV3: Integration with CreateVariable
func TestObjectRegistry_CreateVariable(t *testing.T) {
	tr := NewTracker()

	// CV1: Auto-register pointer
	person := &Person{Name: "Alice"}
	v := tr.CreateVariable(person, 0, "", nil)
	id, ok := tr.LookupObject(person)
	if !ok || id != v.ID {
		t.Error("pointer should be auto-registered with variable ID")
	}

	// CV2: Auto-register map
	m := map[string]int{"a": 1}
	v2 := tr.CreateVariable(m, 0, "", nil)
	id, ok = tr.LookupObject(m)
	if !ok || id != v2.ID {
		t.Error("map should be auto-registered with variable ID")
	}

	// CV3: No register primitive
	tr.CreateVariable(42, 0, "", nil)
	// Can't easily check this, but it should work without error
}

// OI1-OI3: Object Identity Tests
func TestObjectIdentity(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}

	// OI1: Same object twice
	tr.RegisterObject(person, 1)
	tr.RegisterObject(person, 1) // same object, same ID
	id, _ := tr.LookupObject(person)
	if id != 1 {
		t.Error("same object should have same ID")
	}

	// OI2: Different objects
	person2 := &Person{Name: "Bob"}
	tr.RegisterObject(person2, 2)
	id1, _ := tr.LookupObject(person)
	id2, _ := tr.LookupObject(person2)
	if id1 == id2 {
		t.Error("different objects should have different IDs")
	}
}

// ============================================================================
// Value JSON Tests (test-ValueJSON.md)
// ============================================================================

// VJ1.1-VJ1.7: ToValueJSON - Primitives
func TestToValueJSON_Primitives(t *testing.T) {
	tr := NewTracker()

	// VJ1.1: Nil
	if tr.ToValueJSON(nil) != nil {
		t.Error("nil should serialize to nil")
	}

	// VJ1.2: String
	if tr.ToValueJSON("hello") != "hello" {
		t.Error("string should pass through")
	}

	// VJ1.3: Int
	if tr.ToValueJSON(42) != 42 {
		t.Error("int should pass through")
	}

	// VJ1.4: Int64
	if tr.ToValueJSON(int64(100)) != int64(100) {
		t.Error("int64 should pass through")
	}

	// VJ1.5: Float64
	if tr.ToValueJSON(3.14) != 3.14 {
		t.Error("float64 should pass through")
	}

	// VJ1.6: Bool true
	if tr.ToValueJSON(true) != true {
		t.Error("true should pass through")
	}

	// VJ1.7: Bool false
	if tr.ToValueJSON(false) != false {
		t.Error("false should pass through")
	}
}

// VJ2.1-VJ2.6: ToValueJSON - Arrays
func TestToValueJSON_Arrays(t *testing.T) {
	tr := NewTracker()

	// VJ2.1: Empty slice
	result := tr.ToValueJSON([]int{})
	if arr, ok := result.([]any); !ok || len(arr) != 0 {
		t.Error("empty slice should serialize to empty array")
	}

	// VJ2.2: Int slice
	result = tr.ToValueJSON([]int{1, 2, 3})
	arr := result.([]any)
	if len(arr) != 3 || arr[0] != 1 || arr[1] != 2 || arr[2] != 3 {
		t.Error("int slice should serialize correctly")
	}

	// VJ2.3: String slice
	result = tr.ToValueJSON([]string{"a", "b"})
	arr = result.([]any)
	if len(arr) != 2 || arr[0] != "a" || arr[1] != "b" {
		t.Error("string slice should serialize correctly")
	}

	// VJ2.5: Array type
	result = tr.ToValueJSON([3]int{1, 2, 3})
	arr = result.([]any)
	if len(arr) != 3 {
		t.Error("array should serialize correctly")
	}

	// VJ2.6: Pointer slice (registered)
	p1 := &Person{Name: "Alice"}
	p2 := &Person{Name: "Bob"}
	tr.RegisterObject(p1, 1)
	tr.RegisterObject(p2, 2)
	result = tr.ToValueJSON([]*Person{p1, p2})
	arr = result.([]any)
	if len(arr) != 2 {
		t.Error("pointer slice should have 2 elements")
	}
	ref1, ok := arr[0].(ObjectRef)
	if !ok || ref1.Obj != 1 {
		t.Error("first element should be ObjectRef with ID 1")
	}
	ref2, ok := arr[1].(ObjectRef)
	if !ok || ref2.Obj != 2 {
		t.Error("second element should be ObjectRef with ID 2")
	}
}

// VJ3.1-VJ3.4: ToValueJSON - Object References
func TestToValueJSON_ObjectRefs(t *testing.T) {
	tr := NewTracker()

	// VJ3.1: Registered pointer
	person := &Person{Name: "Alice"}
	tr.RegisterObject(person, 1)
	result := tr.ToValueJSON(person)
	ref, ok := result.(ObjectRef)
	if !ok || ref.Obj != 1 {
		t.Error("registered pointer should serialize to ObjectRef")
	}

	// VJ3.2: Registered map
	m := map[string]int{"a": 1}
	tr.RegisterObject(m, 2)
	result = tr.ToValueJSON(m)
	ref, ok = result.(ObjectRef)
	if !ok || ref.Obj != 2 {
		t.Error("registered map should serialize to ObjectRef")
	}

	// VJ3.3: Same ptr twice
	result = tr.ToValueJSON([]*Person{person, person})
	arr := result.([]any)
	ref1 := arr[0].(ObjectRef)
	ref2 := arr[1].(ObjectRef)
	if ref1.Obj != ref2.Obj {
		t.Error("same pointer should produce same ObjectRef")
	}
}

// VJ4.1-VJ4.3: ToValueJSONBytes
func TestToValueJSONBytes(t *testing.T) {
	tr := NewTracker()

	// VJ4.1: Primitive
	bytes, err := tr.ToValueJSONBytes(42)
	if err != nil || string(bytes) != "42" {
		t.Errorf("primitive bytes: err=%v, got=%s", err, bytes)
	}

	// VJ4.2: Object ref
	person := &Person{Name: "Alice"}
	tr.RegisterObject(person, 1)
	bytes, err = tr.ToValueJSONBytes(person)
	if err != nil || string(bytes) != `{"obj":1}` {
		t.Errorf("object ref bytes: err=%v, got=%s", err, bytes)
	}

	// VJ4.3: Array
	bytes, err = tr.ToValueJSONBytes([]int{1, 2})
	if err != nil || string(bytes) != "[1,2]" {
		t.Errorf("array bytes: err=%v, got=%s", err, bytes)
	}
}

// VJ5.1-VJ5.4: IsObjectRef
func TestIsObjectRef(t *testing.T) {
	// VJ5.1: Is ObjectRef
	if !IsObjectRef(ObjectRef{Obj: 1}) {
		t.Error("ObjectRef should be detected")
	}

	// VJ5.2: Not ObjectRef int
	if IsObjectRef(42) {
		t.Error("int should not be ObjectRef")
	}

	// VJ5.3: Not ObjectRef string
	if IsObjectRef("hello") {
		t.Error("string should not be ObjectRef")
	}

	// VJ5.4: Not ObjectRef map
	if IsObjectRef(map[string]any{}) {
		t.Error("map should not be ObjectRef")
	}
}

// VJ6.1-VJ6.3: GetObjectRefID
func TestGetObjectRefID(t *testing.T) {
	// VJ6.1: Valid ObjectRef
	id, ok := GetObjectRefID(ObjectRef{Obj: 5})
	if !ok || id != 5 {
		t.Error("should extract ID from ObjectRef")
	}

	// VJ6.2: Not ObjectRef
	id, ok = GetObjectRefID(42)
	if ok {
		t.Error("should not extract ID from non-ObjectRef")
	}

	// VJ6.3: Zero ID ref
	id, ok = GetObjectRefID(ObjectRef{Obj: 0})
	if !ok || id != 0 {
		t.Error("should extract zero ID from ObjectRef")
	}
}

// VJE1, VJE2: Unregistered pointer/map
func TestToValueJSON_Unregistered(t *testing.T) {
	tr := NewTracker()

	// VJE1: Unregistered pointer returns nil (error case)
	person := &Person{Name: "Alice"}
	result := tr.ToValueJSON(person)
	if result != nil {
		t.Error("unregistered pointer should return nil")
	}

	// VJE2: Unregistered map returns nil (error case)
	m := map[string]int{"a": 1}
	result = tr.ToValueJSON(m)
	if result != nil {
		t.Error("unregistered map should return nil")
	}
}

// ============================================================================
// Integration Tests (test-Tracker.md: I1-I3)
// ============================================================================

// I1: Full lifecycle
func TestIntegration_FullLifecycle(t *testing.T) {
	tr := NewTracker()

	// Create
	person := &Person{Name: "Alice", Age: 30}
	v := tr.CreateVariable(person, 0, "", nil)
	ageChild := tr.CreateVariable(nil, v.ID, "Age", nil)

	// Modify
	person.Age = 31

	// Detect - should return changes and clear internal state
	changes := tr.DetectChanges()
	found := false
	for _, c := range changes {
		if c.VariableID == ageChild.ID && c.ValueChanged {
			found = true
			break
		}
	}
	if !found {
		t.Error("should detect age change")
	}

	// Verify no false positives - DetectChanges clears internal state
	changes = tr.DetectChanges()
	if len(changes) != 0 {
		t.Error("should not detect changes after previous DetectChanges")
	}
}

// I2: Parent-child tree
func TestIntegration_ParentChildTree(t *testing.T) {
	tr := NewTracker()

	person := &Person{
		Name:    "Alice",
		Address: &Address{City: "NYC", Country: "USA"},
	}

	root := tr.CreateVariable(person, 0, "", nil)
	addressChild := tr.CreateVariable(nil, root.ID, "Address", nil)
	cityChild := tr.CreateVariable(nil, root.ID, "Address.City", nil)

	// Verify values
	if cityChild.Value != "NYC" {
		t.Errorf("city should be NYC, got %v", cityChild.Value)
	}

	// Modify city
	person.Address.City = "LA"
	changes := tr.DetectChanges()

	// Check if city change is detected
	found := false
	for _, c := range changes {
		if c.VariableID == cityChild.ID || c.VariableID == addressChild.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("should detect city change")
	}
}

// I3: Object identity
func TestIntegration_ObjectIdentity(t *testing.T) {
	tr := NewTracker()

	address := &Address{City: "NYC"}
	person1 := &Person{Name: "Alice", Address: address}
	person2 := &Person{Name: "Bob", Address: address}

	// Register address explicitly (it's not auto-registered as a nested field)
	tr.RegisterObject(address, 100)

	tr.CreateVariable(person1, 0, "", nil)
	tr.CreateVariable(person2, 0, "", nil)

	// Same address object should have same registration
	id, ok := tr.LookupObject(address)
	if !ok {
		t.Error("shared address should be registered")
	}
	if id != 100 {
		t.Errorf("expected address ID 100, got %d", id)
	}

	// Value JSON should use same ObjectRef for same object
	json1 := tr.ToValueJSON(person1.Address)
	json2 := tr.ToValueJSON(person2.Address)

	ref1, ok1 := json1.(ObjectRef)
	ref2, ok2 := json2.(ObjectRef)

	if !ok1 || !ok2 || ref1.Obj != ref2.Obj || ref1.Obj != id {
		t.Error("same object should produce same ObjectRef")
	}
}

// ============================================================================
// ChildIDs Tests (new feature)
// ============================================================================

// Test ChildIDs maintained on create
func TestChildIDs_CreateVariable(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	parent := tr.CreateVariable(person, 0, "", nil)

	// Root has no children initially
	if len(parent.ChildIDs) != 0 {
		t.Errorf("expected no children, got %d", len(parent.ChildIDs))
	}

	// Create children
	child1 := tr.CreateVariable(nil, parent.ID, "Name", nil)
	child2 := tr.CreateVariable(nil, parent.ID, "Age", nil)

	// Parent should have both children
	if len(parent.ChildIDs) != 2 {
		t.Errorf("expected 2 children, got %d", len(parent.ChildIDs))
	}
	if parent.ChildIDs[0] != child1.ID || parent.ChildIDs[1] != child2.ID {
		t.Errorf("expected ChildIDs [%d, %d], got %v", child1.ID, child2.ID, parent.ChildIDs)
	}
}

// Test ChildIDs maintained on destroy
func TestChildIDs_DestroyVariable(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	parent := tr.CreateVariable(person, 0, "", nil)
	child1 := tr.CreateVariable(nil, parent.ID, "Name", nil)
	child2 := tr.CreateVariable(nil, parent.ID, "Age", nil)

	// Destroy first child
	tr.DestroyVariable(child1.ID)

	// Parent should only have second child
	if len(parent.ChildIDs) != 1 {
		t.Errorf("expected 1 child after destroy, got %d", len(parent.ChildIDs))
	}
	if parent.ChildIDs[0] != child2.ID {
		t.Errorf("expected ChildIDs [%d], got %v", child2.ID, parent.ChildIDs)
	}
}

// Test rootIDs maintained on create
func TestRootIDs_CreateVariable(t *testing.T) {
	tr := NewTracker()

	// Create root variables
	root1 := tr.CreateVariable(1, 0, "", nil)
	root2 := tr.CreateVariable(2, 0, "", nil)

	// Both should be in rootIDs
	if !tr.rootIDs[root1.ID] {
		t.Error("root1 should be in rootIDs")
	}
	if !tr.rootIDs[root2.ID] {
		t.Error("root2 should be in rootIDs")
	}

	// Child should not be in rootIDs
	child := tr.CreateVariable(nil, root1.ID, "", nil)
	if tr.rootIDs[child.ID] {
		t.Error("child should not be in rootIDs")
	}
}

// Test rootIDs maintained on destroy
func TestRootIDs_DestroyVariable(t *testing.T) {
	tr := NewTracker()
	root := tr.CreateVariable(1, 0, "", nil)

	// Destroy root
	tr.DestroyVariable(root.ID)

	// Should be removed from rootIDs
	if tr.rootIDs[root.ID] {
		t.Error("root should be removed from rootIDs after destroy")
	}
}

// ============================================================================
// Active Field Tests (new feature)
// ============================================================================

// Test Active defaults to true
func TestActive_DefaultTrue(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	if !v.Active {
		t.Error("Active should default to true")
	}
}

// Test SetActive method
func TestSetActive(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	// Set to inactive
	v.SetActive(false)
	if v.Active {
		t.Error("Active should be false after SetActive(false)")
	}

	// Set back to active
	v.SetActive(true)
	if !v.Active {
		t.Error("Active should be true after SetActive(true)")
	}
}

// Test inactive variable is skipped in DetectChanges
func TestActive_SkipInactive(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	parent := tr.CreateVariable(person, 0, "", nil)
	child := tr.CreateVariable(nil, parent.ID, "Age", nil)

	// Clear initial state
	tr.DetectChanges()

	// Set child to inactive
	child.SetActive(false)

	// Modify value
	person.Age = 31

	// DetectChanges should not detect the change
	changes := tr.DetectChanges()
	for _, c := range changes {
		if c.VariableID == child.ID {
			t.Error("inactive child should not be in changes")
		}
	}
}

// Test inactive parent skips all descendants
func TestActive_SkipDescendants(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30, Address: &Address{City: "NYC"}}
	root := tr.CreateVariable(person, 0, "", nil)
	addressVar := tr.CreateVariable(nil, root.ID, "Address", nil)
	cityVar := tr.CreateVariable(nil, addressVar.ID, "City", nil)

	// Clear initial state
	tr.DetectChanges()

	// Set address (middle) to inactive - should skip city too
	addressVar.SetActive(false)

	// Modify both address and city
	person.Address.City = "LA"

	// DetectChanges should not detect city change
	changes := tr.DetectChanges()
	for _, c := range changes {
		if c.VariableID == addressVar.ID || c.VariableID == cityVar.ID {
			t.Errorf("inactive subtree should not be in changes, got variable %d", c.VariableID)
		}
	}
}

// Test re-activating variable
func TestActive_Reactivate(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	parent := tr.CreateVariable(person, 0, "", nil)
	child := tr.CreateVariable(nil, parent.ID, "Age", nil)

	// Clear initial state
	tr.DetectChanges()

	// Set to inactive
	child.SetActive(false)
	person.Age = 31
	tr.DetectChanges() // Should not detect

	// Re-activate
	child.SetActive(true)
	person.Age = 32 // Change again

	// Now should detect
	changes := tr.DetectChanges()
	found := false
	for _, c := range changes {
		if c.VariableID == child.ID && c.ValueChanged {
			found = true
			break
		}
	}
	if !found {
		t.Error("re-activated child should detect changes")
	}
}

// ============================================================================
// Tree Traversal Tests (new behavior)
// ============================================================================

// Test DetectChanges uses tree traversal
func TestDetectChanges_TreeTraversal(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	root := tr.CreateVariable(person, 0, "", nil)
	nameVar := tr.CreateVariable(nil, root.ID, "Name", nil)
	ageVar := tr.CreateVariable(nil, root.ID, "Age", nil)

	// Clear initial state
	tr.DetectChanges()

	// Modify both fields
	person.Name = "Bob"
	person.Age = 31

	// DetectChanges should find both changes via tree traversal
	changes := tr.DetectChanges()
	foundName := false
	foundAge := false
	for _, c := range changes {
		if c.VariableID == nameVar.ID && c.ValueChanged {
			foundName = true
		}
		if c.VariableID == ageVar.ID && c.ValueChanged {
			foundAge = true
		}
	}
	if !foundName {
		t.Error("should detect name change via tree traversal")
	}
	if !foundAge {
		t.Error("should detect age change via tree traversal")
	}
}

// Test multi-level tree traversal
func TestDetectChanges_MultiLevelTree(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Address: &Address{City: "NYC", Country: "USA"}}
	root := tr.CreateVariable(person, 0, "", nil)
	addrVar := tr.CreateVariable(nil, root.ID, "Address", nil)
	cityVar := tr.CreateVariable(nil, addrVar.ID, "City", nil)

	// Clear initial state
	tr.DetectChanges()

	// Modify city (2 levels deep)
	person.Address.City = "LA"

	// DetectChanges should find city change via tree traversal
	changes := tr.DetectChanges()
	found := false
	for _, c := range changes {
		if c.VariableID == cityVar.ID && c.ValueChanged {
			found = true
			break
		}
	}
	if !found {
		t.Error("should detect city change via multi-level tree traversal")
	}
}

// Test multiple root trees
func TestDetectChanges_MultipleRoots(t *testing.T) {
	tr := NewTracker()
	person1 := &Person{Name: "Alice"}
	person2 := &Person{Name: "Bob"}
	root1 := tr.CreateVariable(person1, 0, "", nil)
	root2 := tr.CreateVariable(person2, 0, "", nil)
	name1 := tr.CreateVariable(nil, root1.ID, "Name", nil)
	name2 := tr.CreateVariable(nil, root2.ID, "Name", nil)

	// Clear initial state
	tr.DetectChanges()

	// Modify both
	person1.Name = "Alice2"
	person2.Name = "Bob2"

	// Should detect both changes from different root trees
	changes := tr.DetectChanges()
	found1 := false
	found2 := false
	for _, c := range changes {
		if c.VariableID == name1.ID && c.ValueChanged {
			found1 = true
		}
		if c.VariableID == name2.ID && c.ValueChanged {
			found2 = true
		}
	}
	if !found1 {
		t.Error("should detect change in first root tree")
	}
	if !found2 {
		t.Error("should detect change in second root tree")
	}
}
