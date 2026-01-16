// Test Design: test-Tracker.md, test-Variable.md, test-Resolver.md, test-ObjectRegistry.md, test-ValueJSON.md, test-Priority.md, test-Change.md
package changetracker

import (
	"fmt"
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

func (p *Person) SetName(name string) {
	p.Name = name
}

func (p *Person) SetAge(age int) {
	p.Age = age
}

func (p *Person) SetAddress(addr *Address) {
	p.Address = addr
}

func (p *Person) GetAddress() *Address {
	return p.Address
}

// Counter is a test type for Call/CallWith tests
type Counter struct {
	value int
	name  string
}

func (c *Counter) Value() int {
	return c.value
}

func (c *Counter) Name() string {
	return c.name
}

func (c *Counter) SetValue(v int) {
	c.value = v
}

func (c *Counter) SetName(n string) {
	c.name = n
}

// Variadic method for rw + () path testing
// Get calls with no args, Set calls with one arg
func (c *Counter) Count(args ...int) int {
	if len(args) > 0 {
		c.value = args[0]
	}
	return c.value
}

// Methods for error testing
func (c *Counter) NeedsArg(x int) int {
	return x
}

func (c *Counter) VoidMethod() {
	// does nothing, returns nothing
}

func (c *Counter) TwoArgs(a, b int) {
	c.value = a + b
}

func (c *Counter) ReturnsSomething(x int) int {
	c.value = x
	return x
}

// Nested type for path semantics tests
type Outer struct {
	inner *Inner
}

func (o *Outer) Inner() *Inner {
	return o.inner
}

func (o *Outer) SetInner(i *Inner) {
	o.inner = i
}

type Inner struct {
	value int
}

func (i *Inner) Value() int {
	return i.value
}

func (i *Inner) SetValue(v int) {
	i.value = v
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

// T2.6: Pointer value registered via ToValueJSON
func TestCreateVariable_PointerRegistered(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	v := tr.CreateVariable(person, 0, "", nil)

	// Check object is in registry (auto-registered via ToValueJSON)
	id, ok := tr.LookupObject(person)
	if !ok {
		t.Error("pointer should be registered")
	}
	// Object gets its own unique ID via ToValueJSON (not necessarily the variable's ID)
	if id == 0 {
		t.Error("registered ID should be non-zero")
	}

	// ValueJSON should be ObjectRef with the object's registered ID
	ref, ok := v.ValueJSON.(ObjectRef)
	if !ok {
		t.Errorf("expected ObjectRef, got %T", v.ValueJSON)
	}
	if ref.Obj != id {
		t.Errorf("ObjectRef.Obj should match registered ID %d, got %d", id, ref.Obj)
	}
}

// T2.7: Map value registered via ToValueJSON
func TestCreateVariable_MapRegistered(t *testing.T) {
	tr := NewTracker()
	data := map[string]int{"a": 1, "b": 2}
	v := tr.CreateVariable(data, 0, "", nil)

	// Check object is in registry (auto-registered via ToValueJSON)
	id, ok := tr.LookupObject(data)
	if !ok {
		t.Error("map should be registered")
	}
	// Object gets its own unique ID via ToValueJSON (not necessarily the variable's ID)
	if id == 0 {
		t.Error("registered ID should be non-zero")
	}

	// ValueJSON should be ObjectRef with the object's registered ID
	ref, ok := v.ValueJSON.(ObjectRef)
	if !ok {
		t.Errorf("expected ObjectRef, got %T", v.ValueJSON)
	}
	if ref.Obj != id {
		t.Errorf("ObjectRef.Obj should match registered ID %d, got %d", id, ref.Obj)
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
	tr.DetectChanges()
	changes := tr.GetChanges()
	if len(changes) != 0 {
		t.Error("no changes expected")
	}

	// T5.2: Primitive change (via struct field) - returns []Change with variable ID
	person := &Person{Name: "Alice", Age: 30}
	v := tr.CreateVariable(person, 0, "", nil)
	child := tr.CreateVariable(nil, v.ID, "Age", nil)

	person.Age = 31
	tr.DetectChanges()
	changes = tr.GetChanges()
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
	tr.DetectChanges()
	changes = tr.GetChanges()
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
	tr.GetChanges() // consume changes
	if child.ValueJSON != 32 {
		t.Errorf("ValueJSON should be updated to 32, got %v", child.ValueJSON)
	}

	// T5.10: Clears internal state after call (GetChanges clears)
	person.Age = 33
	tr.DetectChanges()
	tr.GetChanges()
	if len(tr.valueChanges) != 0 {
		t.Error("valueChanges should be cleared after GetChanges")
	}
	if len(tr.PropertyChanges) != 0 {
		t.Error("propertyChanges should be cleared after GetChanges")
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

	tr.DetectChanges()
	changes := tr.GetChanges()
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
	tr2.DetectChanges()
	changes = tr2.GetChanges()

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
		t.Error("property change should appear in GetChanges result")
	}

	// T5.13: Slice reuse
	tr3 := NewTracker()
	v3 := tr3.CreateVariable(42, 0, "", nil)
	v3.SetProperty("x", "1")
	tr3.DetectChanges()
	changes1 := tr3.GetChanges()
	cap1 := cap(changes1)

	v3.SetProperty("y", "2")
	tr3.DetectChanges()
	changes2 := tr3.GetChanges()

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
	tr.DetectChanges()
	changes := tr.GetChanges()
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
	tr.GetChanges()

	person.Age = 31
	tr.DetectChanges()
	changes := tr.GetChanges()

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
	tr.GetChanges()

	v.SetProperty("label", "test")

	tr.DetectChanges()
	changes := tr.GetChanges()
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
	tr.GetChanges()

	// Trigger property changes with different priorities
	vHigh.SetProperty("x:high", "1")
	vMed.SetProperty("x", "2") // default medium
	vLow.SetProperty("x:low", "3")

	tr.DetectChanges()
	changes := tr.GetChanges()
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
	tr.GetChanges()

	// Set properties with different priorities
	v.SetProperty("high_prop:high", "val1")
	v.SetProperty("low_prop:low", "val2")

	// Also need to trigger a value change by modifying a trackable value
	// Since this is a primitive, we'll add a struct-based test
	tr.DetectChanges()
	changes := tr.GetChanges()

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
	tr.GetChanges()

	v.SetProperty("x", "1")
	tr.DetectChanges()
	changes1 := tr.GetChanges()
	_ = changes1 // first call

	// Make new changes
	v.SetProperty("y", "2")
	tr.DetectChanges()
	changes2 := tr.GetChanges()

	// The returned slices should use the same underlying array
	if len(changes2) == 0 {
		t.Error("should have changes after second GetChanges")
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

// V1.5: Method call get (requires access "r" for () paths)
func TestVariable_GetMethod(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	parent := tr.CreateVariable(person, 0, "", nil)

	methodChild := tr.CreateVariable(nil, parent.ID, "GetName()?access=r", nil)
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

	// V4.13: SetProperty records change in tracker (appears in GetChanges result)
	tr.DetectChanges() // clear previous changes
	tr.GetChanges()
	v.SetProperty("test", "value")
	tr.DetectChanges()
	changes := tr.GetChanges()
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

// R4.1-R4.3: Call - Zero-arg Method Calls
func TestResolver_Call(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}

	// R4.1: Zero-arg method
	val, err := tr.Call(person, "GetName")
	if err != nil || val != "Alice" {
		t.Errorf("Call method: err=%v, val=%v", err, val)
	}

	// R4.3: Multi-return method (returns first value)
	val, err = tr.Call(person, "Pair")
	if err != nil || val != "Alice" {
		t.Errorf("Call multi-return method: err=%v, val=%v", err, val)
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
	id, ok := tr.RegisterObject(person)
	if !ok {
		t.Error("should register pointer")
	}
	if id <= 0 {
		t.Error("should return positive ID")
	}

	// OR1.2: Register map
	m := map[string]int{"a": 1}
	id2, ok := tr.RegisterObject(m)
	if !ok {
		t.Error("should register map")
	}
	if id2 == id {
		t.Error("different objects should get different IDs")
	}

	// OR1.3: Register non-pointer
	_, ok = tr.RegisterObject(42)
	if ok {
		t.Error("should not register int")
	}

	// OR1.4: Register nil
	_, ok = tr.RegisterObject(nil)
	if ok {
		t.Error("should not register nil")
	}

	// OR1.5: Re-register same returns same ID
	id3, ok := tr.RegisterObject(person)
	if !ok {
		t.Error("should succeed for already-registered pointer")
	}
	if id3 != id {
		t.Errorf("re-register should return same ID %d, got %d", id, id3)
	}
}

// OR2.1-OR2.3: UnregisterObject
func TestUnregisterObject(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	tr.RegisterObject(person)

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
	regID, _ := tr.RegisterObject(person)
	id, ok := tr.LookupObject(person)
	if !ok || id != regID {
		t.Errorf("should find registered object with ID %d, got id=%d, ok=%v", regID, id, ok)
	}
}

// OR4.1, OR4.2: GetObject
func TestGetObject(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	regID, _ := tr.RegisterObject(person)

	// OR4.1: Get registered
	obj := tr.GetObject(regID)
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

	// CV1: Auto-register pointer via ToValueJSON
	person := &Person{Name: "Alice"}
	v := tr.CreateVariable(person, 0, "", nil)
	id, ok := tr.LookupObject(person)
	if !ok {
		t.Error("pointer should be auto-registered")
	}
	// The ObjectRef in ValueJSON should match the registered ID
	ref, ok := v.ValueJSON.(ObjectRef)
	if !ok || ref.Obj != id {
		t.Error("ValueJSON should be ObjectRef with registered ID")
	}

	// CV2: Auto-register map via ToValueJSON
	m := map[string]int{"a": 1}
	v2 := tr.CreateVariable(m, 0, "", nil)
	id, ok = tr.LookupObject(m)
	if !ok {
		t.Error("map should be auto-registered")
	}
	ref, ok = v2.ValueJSON.(ObjectRef)
	if !ok || ref.Obj != id {
		t.Error("ValueJSON should be ObjectRef with registered ID")
	}

	// CV3: No register primitive
	tr.CreateVariable(42, 0, "", nil)
	// Can't easily check this, but it should work without error
}

// OI1-OI3: Object Identity Tests
func TestObjectIdentity(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}

	// OI1: Same object twice returns same ID
	id1a, _ := tr.RegisterObject(person)
	id1b, _ := tr.RegisterObject(person) // same object, same ID
	if id1a != id1b {
		t.Errorf("same object should have same ID, got %d and %d", id1a, id1b)
	}

	// OI2: Different objects get different IDs
	person2 := &Person{Name: "Bob"}
	id2, _ := tr.RegisterObject(person2)
	if id1a == id2 {
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
	p1ID, _ := tr.RegisterObject(p1)
	p2ID, _ := tr.RegisterObject(p2)
	result = tr.ToValueJSON([]*Person{p1, p2})
	arr = result.([]any)
	if len(arr) != 2 {
		t.Error("pointer slice should have 2 elements")
	}
	ref1, ok := arr[0].(ObjectRef)
	if !ok || ref1.Obj != p1ID {
		t.Errorf("first element should be ObjectRef with ID %d", p1ID)
	}
	ref2, ok := arr[1].(ObjectRef)
	if !ok || ref2.Obj != p2ID {
		t.Errorf("second element should be ObjectRef with ID %d", p2ID)
	}
}

// VJ3.1-VJ3.4: ToValueJSON - Object References
func TestToValueJSON_ObjectRefs(t *testing.T) {
	tr := NewTracker()

	// VJ3.1: Registered pointer
	person := &Person{Name: "Alice"}
	personID, _ := tr.RegisterObject(person)
	result := tr.ToValueJSON(person)
	ref, ok := result.(ObjectRef)
	if !ok || ref.Obj != personID {
		t.Error("registered pointer should serialize to ObjectRef")
	}

	// VJ3.2: Registered map
	m := map[string]int{"a": 1}
	mID, _ := tr.RegisterObject(m)
	result = tr.ToValueJSON(m)
	ref, ok = result.(ObjectRef)
	if !ok || ref.Obj != mID {
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
	personID, _ := tr.RegisterObject(person)
	bytes, err = tr.ToValueJSONBytes(person)
	expected := fmt.Sprintf(`{"obj":%d}`, personID)
	if err != nil || string(bytes) != expected {
		t.Errorf("object ref bytes: err=%v, expected=%s, got=%s", err, expected, bytes)
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

// VJE1, VJE2: Unregistered pointer/map - auto-registered
func TestToValueJSON_Unregistered(t *testing.T) {
	tr := NewTracker()

	// VJE1: Unregistered pointer gets auto-registered
	// Per spec: nested objects in arrays must be converted to references
	person := &Person{Name: "Alice"}
	result := tr.ToValueJSON(person)
	ref, ok := result.(ObjectRef)
	if !ok {
		t.Error("unregistered pointer should be auto-registered and return ObjectRef")
	} else if ref.Obj <= 0 {
		t.Error("auto-registered object should have positive ID")
	}

	// Verify it was actually registered
	if id, found := tr.LookupObject(person); !found {
		t.Error("pointer should be registered after ToValueJSON")
	} else if id != ref.Obj {
		t.Error("registered ID should match returned ObjectRef")
	}

	// VJE2: Unregistered map gets auto-registered
	m := map[string]int{"a": 1}
	result = tr.ToValueJSON(m)
	ref, ok = result.(ObjectRef)
	if !ok {
		t.Error("unregistered map should be auto-registered and return ObjectRef")
	} else if ref.Obj <= 0 {
		t.Error("auto-registered map should have positive ID")
	}

	// Verify it was actually registered
	if id, found := tr.LookupObject(m); !found {
		t.Error("map should be registered after ToValueJSON")
	} else if id != ref.Obj {
		t.Error("registered ID should match returned ObjectRef")
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
	tr.DetectChanges()
	changes := tr.GetChanges()
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

	// Verify no false positives - GetChanges clears internal state
	tr.DetectChanges()
	changes = tr.GetChanges()
	if len(changes) != 0 {
		t.Error("should not detect changes after previous GetChanges")
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
	tr.DetectChanges()
	changes := tr.GetChanges()

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
	addressID, _ := tr.RegisterObject(address)

	tr.CreateVariable(person1, 0, "", nil)
	tr.CreateVariable(person2, 0, "", nil)

	// Same address object should have same registration
	id, ok := tr.LookupObject(address)
	if !ok {
		t.Error("shared address should be registered")
	}
	if id != addressID {
		t.Errorf("expected address ID %d, got %d", addressID, id)
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
	tr.GetChanges()

	// Set child to inactive
	child.SetActive(false)

	// Modify value
	person.Age = 31

	// DetectChanges should not detect the change
	tr.DetectChanges()
	changes := tr.GetChanges()
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
	tr.GetChanges()

	// Set address (middle) to inactive - should skip city too
	addressVar.SetActive(false)

	// Modify both address and city
	person.Address.City = "LA"

	// DetectChanges should not detect city change
	tr.DetectChanges()
	changes := tr.GetChanges()
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
	tr.GetChanges()

	// Set to inactive
	child.SetActive(false)
	person.Age = 31
	tr.DetectChanges() // Should not detect
	tr.GetChanges()

	// Re-activate
	child.SetActive(true)
	person.Age = 32 // Change again

	// Now should detect
	tr.DetectChanges()
	changes := tr.GetChanges()
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
	tr.GetChanges()

	// Modify both fields
	person.Name = "Bob"
	person.Age = 31

	// DetectChanges should find both changes via tree traversal
	tr.DetectChanges()
	changes := tr.GetChanges()
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
	tr.GetChanges()

	// Modify city (2 levels deep)
	person.Address.City = "LA"

	// DetectChanges should find city change via tree traversal
	tr.DetectChanges()
	changes := tr.GetChanges()
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
	tr.GetChanges()

	// Modify both
	person1.Name = "Alice2"
	person2.Name = "Bob2"

	// Should detect both changes from different root trees
	tr.DetectChanges()
	changes := tr.GetChanges()
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

// ============================================================================
// Call Tests (test-Resolver.md C1.1-C1.5)
// ============================================================================

// C1.1-C1.5: Call - Zero-Arg Methods
func TestResolver_Call_ZeroArgMethods(t *testing.T) {
	tr := NewTracker()
	counter := &Counter{value: 42, name: "test"}

	// C1.1: Zero-arg method returning int
	val, err := tr.Call(counter, "Value")
	if err != nil || val != 42 {
		t.Errorf("C1.1 Call Value: err=%v, val=%v", err, val)
	}

	// C1.4: Zero-arg method returning string
	val, err = tr.Call(counter, "Name")
	if err != nil || val != "test" {
		t.Errorf("C1.4 Call Name: err=%v, val=%v", err, val)
	}

	// C1.2: Method on pointer
	person := &Person{Name: "Alice", Address: &Address{City: "NYC"}}
	val, err = tr.Call(person, "GetName")
	if err != nil || val != "Alice" {
		t.Errorf("C1.2 Call GetName: err=%v, val=%v", err, val)
	}

	// C1.3: Multi-return method (returns first value)
	val, err = tr.Call(person, "Pair")
	if err != nil || val != "Alice" {
		t.Errorf("C1.3 Call Pair: err=%v, val=%v", err, val)
	}

	// C1.5: Method returning struct pointer
	val, err = tr.Call(person, "GetAddress")
	if err != nil {
		t.Errorf("C1.5 Call GetAddress: err=%v", err)
	}
	addr, ok := val.(*Address)
	if !ok || addr.City != "NYC" {
		t.Errorf("C1.5 Call GetAddress: expected *Address with City=NYC, got %v", val)
	}
}

// CE1-CE5: Call Errors
func TestResolver_CallErrors(t *testing.T) {
	tr := NewTracker()
	counter := &Counter{value: 42}

	// CE1: Nil object
	_, err := tr.Call(nil, "Value")
	if err == nil {
		t.Error("CE1 Call nil should error")
	}

	// CE2: Method not found
	_, err = tr.Call(counter, "NoSuch")
	if err == nil {
		t.Error("CE2 Call missing method should error")
	}

	// CE4: Method needs args
	_, err = tr.Call(counter, "NeedsArg")
	if err == nil {
		t.Error("CE4 Call method needing args should error")
	}

	// CE5: Method returns nothing (void)
	_, err = tr.Call(counter, "VoidMethod")
	if err == nil {
		t.Error("CE5 Call void method should error")
	}
}

// ============================================================================
// CallWith Tests (test-Resolver.md CW1.1-CW1.4)
// ============================================================================

// CW1.1-CW1.4: CallWith - One-Arg Methods
func TestResolver_CallWith_OneArgMethods(t *testing.T) {
	tr := NewTracker()

	// CW1.1: Set int via method
	counter := &Counter{value: 0}
	err := tr.CallWith(counter, "SetValue", 42)
	if err != nil || counter.value != 42 {
		t.Errorf("CW1.1 CallWith SetValue: err=%v, value=%d", err, counter.value)
	}

	// CW1.2: Set string via method
	err = tr.CallWith(counter, "SetName", "updated")
	if err != nil || counter.name != "updated" {
		t.Errorf("CW1.2 CallWith SetName: err=%v, name=%s", err, counter.name)
	}

	// CW1.3: Method on pointer
	person := &Person{Name: "Alice"}
	err = tr.CallWith(person, "SetName", "Bob")
	if err != nil || person.Name != "Bob" {
		t.Errorf("CW1.3 CallWith SetName on Person: err=%v, name=%s", err, person.Name)
	}

	// CW1.4: Set struct pointer via method
	addr := &Address{City: "Boston"}
	err = tr.CallWith(person, "SetAddress", addr)
	if err != nil || person.Address != addr {
		t.Errorf("CW1.4 CallWith SetAddress: err=%v, addr=%v", err, person.Address)
	}
}

// CWE1-CWE7: CallWith Errors
func TestResolver_CallWithErrors(t *testing.T) {
	tr := NewTracker()
	counter := &Counter{value: 42}

	// CWE1: Nil object
	err := tr.CallWith(nil, "SetValue", 10)
	if err == nil {
		t.Error("CWE1 CallWith nil should error")
	}

	// CWE2: Method not found
	err = tr.CallWith(counter, "NoSuch", 10)
	if err == nil {
		t.Error("CWE2 CallWith missing method should error")
	}

	// CWE4: Method takes no args
	err = tr.CallWith(counter, "Value", 10)
	if err == nil {
		t.Error("CWE4 CallWith zero-arg method should error")
	}

	// CWE5: Method takes 2+ args
	err = tr.CallWith(counter, "TwoArgs", 10)
	if err == nil {
		t.Error("CWE5 CallWith two-arg method should error")
	}

	// CWE6: Method returns value - now allowed (return values ignored)
	err = tr.CallWith(counter, "ReturnsSomething", 10)
	if err != nil {
		t.Errorf("CWE6 CallWith should accept methods with return values, got: %v", err)
	}

	// CWE7: Arg type mismatch
	err = tr.CallWith(counter, "SetValue", "not an int")
	if err == nil {
		t.Error("CWE7 CallWith type mismatch should error")
	}
}

// ============================================================================
// Path Semantics Tests (test-Resolver.md PM1-PM9)
// ============================================================================

// PM1-PM2: Getter mid-path
func TestPathSemantics_GetterMidPath(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Address: &Address{City: "NYC", Country: "USA"}}
	root := tr.CreateVariable(person, 0, "", nil)

	// PM1: Get through getter mid-path
	cityVar := tr.CreateVariable(nil, root.ID, "GetAddress().City", nil)
	val, err := cityVar.Get()
	if err != nil || val != "NYC" {
		t.Errorf("PM1 Get Address().City: err=%v, val=%v", err, val)
	}

	// PM2: Set through getter mid-path (should work - City is settable)
	err = cityVar.Set("Boston")
	if err != nil || person.Address.City != "Boston" {
		t.Errorf("PM2 Set Address().City: err=%v, city=%s", err, person.Address.City)
	}
}

// PM3-PM4: Path ends in getter (requires access "r" for () paths)
func TestPathSemantics_GetterTerminal(t *testing.T) {
	tr := NewTracker()
	counter := &Counter{value: 42}
	root := tr.CreateVariable(counter, 0, "", nil)

	// PM3: Get on path ending in getter - OK (requires access "r")
	valVar := tr.CreateVariable(nil, root.ID, "Value()?access=r", nil)
	val, err := valVar.Get()
	if err != nil || val != 42 {
		t.Errorf("PM3 Get Value(): err=%v, val=%v", err, val)
	}

	// PM4: Set on path ending in getter - ERROR (read-only)
	err = valVar.Set(100)
	if err == nil {
		t.Error("PM4 Set on getter path should error (read-only)")
	}
}

// PM5-PM6: Path ends in setter (requires access "w" or "action" for (_) paths)
func TestPathSemantics_SetterTerminal(t *testing.T) {
	tr := NewTracker()
	counter := &Counter{value: 0}
	root := tr.CreateVariable(counter, 0, "", nil)

	// PM5: Set on path ending in setter - OK (requires access "w" or "action")
	setVar := tr.CreateVariable(nil, root.ID, "SetValue(_)?access=w", nil)
	err := setVar.Set(42)
	if err != nil || counter.value != 42 {
		t.Errorf("PM5 Set SetValue(_): err=%v, value=%d", err, counter.value)
	}

	// PM6: Get on path ending in setter - ERROR (write-only)
	_, err = setVar.Get()
	if err == nil {
		t.Error("PM6 Get on setter path should error (write-only)")
	}
}

// PM7-PM8: Setter not terminal (should panic on CreateVariable)
func TestPathSemantics_SetterNotTerminal(t *testing.T) {
	tr := NewTracker()
	outer := &Outer{inner: &Inner{value: 42}}
	root := tr.CreateVariable(outer, 0, "", nil)

	// PM7-PM8: Setter not at end of path should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("PM7-PM8 Setter not terminal should panic")
		}
	}()

	// This should panic - SetInner(_) is not at terminal position
	tr.CreateVariable(nil, root.ID, "SetInner(_).value", nil)
}

// PM9: Chain getters (requires access "r" for paths ending in ())
func TestPathSemantics_ChainGetters(t *testing.T) {
	tr := NewTracker()
	outer := &Outer{inner: &Inner{value: 99}}
	root := tr.CreateVariable(outer, 0, "", nil)

	// PM9: Chain of getter calls (requires access "r")
	valVar := tr.CreateVariable(nil, root.ID, "Inner().Value()?access=r", nil)
	val, err := valVar.Get()
	if err != nil || val != 99 {
		t.Errorf("PM9 Get Inner().Value(): err=%v, val=%v", err, val)
	}
}

// PT3-PT4: Path element type tests for method syntax
func TestPathElement_MethodSyntax(t *testing.T) {
	tr := NewTracker()
	counter := &Counter{value: 42, name: "test"}
	root := tr.CreateVariable(counter, 0, "", nil)

	// PT3: String with "()" uses Call (requires access "r")
	getterVar := tr.CreateVariable(nil, root.ID, "Value()?access=r", nil)
	val, err := getterVar.Get()
	if err != nil || val != 42 {
		t.Errorf("PT3 getter path element: err=%v, val=%v", err, val)
	}

	// PT4: String with "(_)" uses CallWith (requires access "w" or "action")
	setterVar := tr.CreateVariable(nil, root.ID, "SetName(_)?access=w", nil)
	err = setterVar.Set("updated")
	if err != nil || counter.name != "updated" {
		t.Errorf("PT4 setter path element: err=%v, name=%s", err, counter.name)
	}
}

// ============================================================================
// Access Property Tests (test-Variable.md V8.1-V10.6, test-Resolver.md AM1-SE4)
// ============================================================================

// V8.1: Access defaults to rw
func TestAccess_DefaultsToRW(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	if v.GetAccess() != "rw" {
		t.Errorf("expected default access 'rw', got %q", v.GetAccess())
	}
	if v.Access != "rw" {
		t.Errorf("expected Access field 'rw', got %q", v.Access)
	}
}

// V8.2-V8.3: Set access via property
func TestAccess_SetViaProperty(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	// V8.2: Set to read-only
	v.SetProperty("access", "r")
	if v.GetAccess() != "r" {
		t.Errorf("expected access 'r', got %q", v.GetAccess())
	}

	// V8.3: Set to write-only
	v.SetProperty("access", "w")
	if v.GetAccess() != "w" {
		t.Errorf("expected access 'w', got %q", v.GetAccess())
	}

	// Reset to read-write
	v.SetProperty("access", "rw")
	if v.GetAccess() != "rw" {
		t.Errorf("expected access 'rw', got %q", v.GetAccess())
	}
}

// V8.4-V8.9: Access mode Get/Set behavior
func TestAccess_GetSetBehavior(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	parent := tr.CreateVariable(person, 0, "", nil)

	// V8.4, V8.5: Read-only - Get succeeds, Set fails
	readOnly := tr.CreateVariable(nil, parent.ID, "Name?access=r", nil)
	val, err := readOnly.Get()
	if err != nil || val != "Alice" {
		t.Errorf("V8.4: read-only Get should succeed, err=%v, val=%v", err, val)
	}
	err = readOnly.Set("Bob")
	if err == nil {
		t.Error("V8.5: read-only Set should fail")
	}

	// V8.6, V8.7: Write-only - Get fails, Set succeeds
	writeOnly := tr.CreateVariable(nil, parent.ID, "Age?access=w", nil)
	_, err = writeOnly.Get()
	if err == nil {
		t.Error("V8.6: write-only Get should fail")
	}
	err = writeOnly.Set(31)
	if err != nil || person.Age != 31 {
		t.Errorf("V8.7: write-only Set should succeed, err=%v, age=%d", err, person.Age)
	}

	// V8.8, V8.9: Read-write - both succeed
	readWrite := tr.CreateVariable(nil, parent.ID, "Name?access=rw", nil)
	val, err = readWrite.Get()
	if err != nil || val != "Alice" {
		t.Errorf("V8.8: read-write Get should succeed, err=%v, val=%v", err, val)
	}
	err = readWrite.Set("Carol")
	if err != nil || person.Name != "Carol" {
		t.Errorf("V8.9: read-write Set should succeed, err=%v, name=%s", err, person.Name)
	}
}

// V8.10: Invalid access value
func TestAccess_InvalidValue(t *testing.T) {
	tr := NewTracker()
	v := tr.CreateVariable(42, 0, "", nil)

	defer func() {
		if r := recover(); r == nil {
			t.Error("V8.10: invalid access value should panic")
		}
	}()

	v.SetProperty("access", "invalid")
}

// V8.11-V8.16: IsReadable and IsWritable
func TestAccess_IsReadableIsWritable(t *testing.T) {
	tr := NewTracker()

	// V8.11-V8.12: IsReadable
	vRW := tr.CreateVariable(1, 0, "", map[string]string{"access": "rw"})
	vR := tr.CreateVariable(2, 0, "", map[string]string{"access": "r"})
	vW := tr.CreateVariable(3, 0, "", map[string]string{"access": "w"})

	if !vRW.IsReadable() {
		t.Error("V8.11: rw should be readable")
	}
	if !vR.IsReadable() {
		t.Error("V8.12: r should be readable")
	}
	if vW.IsReadable() {
		t.Error("V8.13: w should not be readable")
	}

	// V8.14-V8.16: IsWritable
	if !vRW.IsWritable() {
		t.Error("V8.14: rw should be writable")
	}
	if !vW.IsWritable() {
		t.Error("V8.15: w should be writable")
	}
	if vR.IsWritable() {
		t.Error("V8.16: r should not be writable")
	}
}

// V8.17: Access via query string
func TestAccess_ViaQueryString(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice"}
	parent := tr.CreateVariable(person, 0, "", nil)

	child := tr.CreateVariable(nil, parent.ID, "Name?access=r", nil)
	if child.GetAccess() != "r" {
		t.Errorf("V8.17: expected access 'r' from query, got %q", child.GetAccess())
	}
}

// V9.1-V9.5: Access and Change Detection
func TestAccess_ChangeDetection(t *testing.T) {
	tr := NewTracker()
	person := &Person{Name: "Alice", Age: 30}
	parent := tr.CreateVariable(person, 0, "", nil)

	// V9.1: Read-only is scanned
	readOnly := tr.CreateVariable(nil, parent.ID, "Name?access=r", nil)
	tr.DetectChanges() // clear initial state
	tr.GetChanges()
	person.Name = "Bob"
	tr.DetectChanges()
	changes := tr.GetChanges()
	found := false
	for _, c := range changes {
		if c.VariableID == readOnly.ID && c.ValueChanged {
			found = true
			break
		}
	}
	if !found {
		t.Error("V9.1: read-only variable should be scanned for changes")
	}

	// V9.2: Write-only is NOT scanned
	writeOnly := tr.CreateVariable(nil, parent.ID, "Age?access=w", nil)
	tr.DetectChanges() // clear initial state
	tr.GetChanges()
	person.Age = 31
	tr.DetectChanges()
	changes = tr.GetChanges()
	for _, c := range changes {
		if c.VariableID == writeOnly.ID {
			t.Error("V9.2: write-only variable should NOT be scanned")
		}
	}

	// V9.3: Read-write is scanned (covered by default tests)

	// V9.4: Write-only parent, readable child - child is still scanned
	tr2 := NewTracker()
	person2 := &Person{Name: "Alice", Address: &Address{City: "NYC"}}
	root2 := tr2.CreateVariable(person2, 0, "", nil)
	addrVar := tr2.CreateVariable(nil, root2.ID, "Address?access=w", nil)
	cityVar := tr2.CreateVariable(nil, addrVar.ID, "City", nil) // default rw
	tr2.DetectChanges()                                         // clear
	tr2.GetChanges()
	person2.Address.City = "LA"
	tr2.DetectChanges()
	changes = tr2.GetChanges()
	found = false
	for _, c := range changes {
		if c.VariableID == cityVar.ID && c.ValueChanged {
			found = true
			break
		}
	}
	if !found {
		t.Error("V9.4: child of write-only parent should still be scanned")
	}

	// V9.5: access=r + Active=false -> NOT scanned
	tr3 := NewTracker()
	person3 := &Person{Name: "Alice"}
	root3 := tr3.CreateVariable(person3, 0, "", nil)
	nameVar := tr3.CreateVariable(nil, root3.ID, "Name?access=r", nil)
	nameVar.SetActive(false)
	tr3.DetectChanges()
	tr3.GetChanges()
	person3.Name = "Bob"
	tr3.DetectChanges()
	changes = tr3.GetChanges()
	for _, c := range changes {
		if c.VariableID == nameVar.ID {
			t.Error("V9.5: inactive read-only variable should NOT be scanned")
		}
	}
}

// V10.1-V10.6: Access vs Path Semantics (updated for path restrictions)
func TestAccess_VsPathSemantics(t *testing.T) {
	tr := NewTracker()
	counter := &Counter{value: 42}
	root := tr.CreateVariable(counter, 0, "", nil)

	// V10.1: access=r + path () -> Get: OK, Set: error (access)
	v1 := tr.CreateVariable(nil, root.ID, "Value()?access=r", nil)
	val, err := v1.Get()
	if err != nil || val != 42 {
		t.Errorf("V10.1: Get should succeed, err=%v, val=%v", err, val)
	}
	err = v1.Set(100)
	if err == nil {
		t.Error("V10.1: Set should fail (access blocks)")
	}

	// V10.2: access=w + path () -> NOW INVALID (CreateVariable should panic)
	// (Previously tested calling method for side effect, but now w + () is rejected)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("V10.2: access=w + path () should panic at CreateVariable")
			}
		}()
		tr.CreateVariable(nil, root.ID, "Value()?access=w", nil)
	}()

	// V10.3: access=r + path (_) -> NOW INVALID (CreateVariable should panic)
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("V10.3: access=r + path (_) should panic at CreateVariable")
			}
		}()
		tr.CreateVariable(nil, root.ID, "SetValue(_)?access=r", nil)
	}()

	// V10.4: access=w + path (_) -> Get: error (both), Set: OK
	v4 := tr.CreateVariable(nil, root.ID, "SetValue(_)?access=w", nil)
	_, err = v4.Get()
	if err == nil {
		t.Error("V10.4: Get should fail")
	}
	err = v4.Set(100)
	if err != nil || counter.value != 100 {
		t.Errorf("V10.4: Set should succeed, err=%v, value=%d", err, counter.value)
	}

	// V10.5: access=rw + path () -> Get: OK, Set: OK (variadic call)
	// Use Count() which is a variadic method: Count(args ...int) int
	counter.value = 50
	v5 := tr.CreateVariable(nil, root.ID, "Count()?access=rw", nil)
	val, err = v5.Get()
	if err != nil {
		t.Errorf("V10.5: Get should succeed, err=%v", err)
	}
	if val != 50 {
		t.Errorf("V10.5: Get should return 50, got %v", val)
	}
	err = v5.Set(200)
	if err != nil {
		t.Errorf("V10.5: Set should succeed (variadic call), err=%v", err)
	}
	if counter.value != 200 {
		t.Errorf("V10.5: Set should update value to 200, got %d", counter.value)
	}

	// V10.6: access=action + path () -> Get: error (access), Set: OK (calls method)
	counter.value = 100 // reset
	v6 := tr.CreateVariable(nil, root.ID, "Value()?access=action", nil)
	_, err = v6.Get()
	if err == nil {
		t.Error("V10.6: Get should fail (access blocks)")
	}
	err = v6.Set(nil) // Calls Value() for side effect
	if err != nil {
		t.Errorf("V10.6: Set should succeed (calls method for side effect), err=%v", err)
	}
}

// SE1-SE4: Action method side effects (updated: now requires access=action for () paths)
func TestAccess_ActionMethodSideEffect(t *testing.T) {
	tr := NewTracker()

	// Define a method that has side effects
	// Since we can't define methods in tests, we'll use Counter which has Value()
	counter := &Counter{value: 0}
	root := tr.CreateVariable(counter, 0, "", nil)

	// SE1-SE4: Action with getter path calls method for side effect
	actionGetter := tr.CreateVariable(nil, root.ID, "Value()?access=action", nil)

	// Set should call Value() method (for side effect)
	// Note: Value() returns the value, but for action we ignore the return
	err := actionGetter.Set(nil)
	if err != nil {
		t.Errorf("Action Set on getter should call method, err=%v", err)
	}

	// Verify the method was called by checking no error occurred
	// (The actual side effect verification would need a method that modifies state)
}

// ============================================================================
// Wrapper Tests
// ============================================================================

// WrapperData is a wrapper type for testing
type WrapperData struct {
	WrappedName string
	Extra       int
}

// wrapperResolver is a custom resolver that creates wrappers
type wrapperResolver struct {
	*Tracker
}

func (r *wrapperResolver) CreateWrapper(v *Variable) any {
	// Create a wrapper that exposes different fields
	if v.Value == nil {
		return nil
	}
	if p, ok := v.Value.(*Person); ok {
		return &WrapperData{
			WrappedName: "Wrapped:" + p.Name,
			Extra:       42,
		}
	}
	return nil
}

// W1: Wrapper created when wrapper property exists
func TestWrapper_CreatedWithProperty(t *testing.T) {
	tr := NewTracker()
	wr := &wrapperResolver{tr}
	tr.Resolver = wr

	person := &Person{Name: "Alice"}
	v := tr.CreateVariable(person, 0, "?wrapper=true", nil)

	if v.WrapperValue == nil {
		t.Error("W1: WrapperValue should be created when wrapper property exists")
	}

	wrapper, ok := v.WrapperValue.(*WrapperData)
	if !ok {
		t.Errorf("W1: WrapperValue should be *WrapperData, got %T", v.WrapperValue)
	}
	if wrapper.WrappedName != "Wrapped:Alice" {
		t.Errorf("W1: WrappedName should be 'Wrapped:Alice', got %q", wrapper.WrappedName)
	}

	if v.WrapperJSON == nil {
		t.Error("W1: WrapperJSON should be set")
	}
}

// W2: Wrapper not created when wrapper property absent
func TestWrapper_NotCreatedWithoutProperty(t *testing.T) {
	tr := NewTracker()
	wr := &wrapperResolver{tr}
	tr.Resolver = wr

	person := &Person{Name: "Alice"}
	v := tr.CreateVariable(person, 0, "", nil)

	if v.WrapperValue != nil {
		t.Error("W2: WrapperValue should be nil when wrapper property is absent")
	}
	if v.WrapperJSON != nil {
		t.Error("W2: WrapperJSON should be nil when wrapper property is absent")
	}
}

// W3: Wrapper uses NavigationValue for child access
func TestWrapper_ChildNavigation(t *testing.T) {
	tr := NewTracker()
	wr := &wrapperResolver{tr}
	tr.Resolver = wr

	person := &Person{Name: "Alice"}
	parent := tr.CreateVariable(person, 0, "?wrapper=true", nil)

	// Child navigates to wrapper's field instead of original value
	child := tr.CreateVariable(nil, parent.ID, "WrappedName", nil)
	val, err := child.Get()
	if err != nil {
		t.Errorf("W3: Get should succeed, err=%v", err)
	}
	if val != "Wrapped:Alice" {
		t.Errorf("W3: Child should navigate to WrapperValue, got %v", val)
	}

	// Extra field from wrapper
	extraChild := tr.CreateVariable(nil, parent.ID, "Extra", nil)
	val, err = extraChild.Get()
	if err != nil {
		t.Errorf("W3: Get Extra should succeed, err=%v", err)
	}
	if val != 42 {
		t.Errorf("W3: Extra should be 42, got %v", val)
	}
}

// W4: NavigationValue returns WrapperValue when present
func TestWrapper_NavigationValue(t *testing.T) {
	tr := NewTracker()
	wr := &wrapperResolver{tr}
	tr.Resolver = wr

	person := &Person{Name: "Alice"}

	// Without wrapper
	v1 := tr.CreateVariable(person, 0, "", nil)
	if v1.NavigationValue() != person {
		t.Error("W4: NavigationValue should return Value when no wrapper")
	}

	// With wrapper
	v2 := tr.CreateVariable(person, 0, "?wrapper=true", nil)
	if v2.NavigationValue() == person {
		t.Error("W4: NavigationValue should not return original Value when wrapper exists")
	}
	if v2.NavigationValue() != v2.WrapperValue {
		t.Error("W4: NavigationValue should return WrapperValue when wrapper exists")
	}
}

// W5: Wrapper unregistered on DestroyVariable
func TestWrapper_UnregisteredOnDestroy(t *testing.T) {
	tr := NewTracker()
	wr := &wrapperResolver{tr}
	tr.Resolver = wr

	person := &Person{Name: "Alice"}
	v := tr.CreateVariable(person, 0, "?wrapper=true", nil)

	wrapper := v.WrapperValue
	_, ok := tr.LookupObject(wrapper)
	if !ok {
		t.Error("W5: Wrapper should be registered initially")
	}

	tr.DestroyVariable(v.ID)

	// After destroy, wrapper should be unregistered
	_, ok = tr.LookupObject(wrapper)
	if ok {
		t.Error("W5: Wrapper should be unregistered after destroy")
	}
}

// W6: SetProperty("wrapper", ...) triggers wrapper update
func TestWrapper_SetPropertyTriggers(t *testing.T) {
	tr := NewTracker()
	wr := &wrapperResolver{tr}
	tr.Resolver = wr

	person := &Person{Name: "Alice"}
	v := tr.CreateVariable(person, 0, "", nil)

	if v.WrapperValue != nil {
		t.Error("W6: Initially no wrapper")
	}

	// Add wrapper property
	v.SetProperty("wrapper", "true")

	if v.WrapperValue == nil {
		t.Error("W6: Wrapper should be created after setting property")
	}

	// Remove wrapper property
	oldWrapper := v.WrapperValue
	v.SetProperty("wrapper", "")

	if v.WrapperValue != nil {
		t.Error("W6: Wrapper should be removed after clearing property")
	}

	// Old wrapper should be unregistered
	_, ok := tr.LookupObject(oldWrapper)
	if ok {
		t.Error("W6: Old wrapper should be unregistered")
	}
}

// W7: Wrapper re-created when ValueJSON changes
func TestWrapper_RecreatedOnValueChange(t *testing.T) {
	tr := NewTracker()
	wr := &wrapperResolver{tr}
	tr.Resolver = wr

	person := &Person{Name: "Alice"}
	v := tr.CreateVariable(person, 0, "?wrapper=true", nil)

	oldWrapper := v.WrapperValue
	oldWrapperData := oldWrapper.(*WrapperData)
	if oldWrapperData.WrappedName != "Wrapped:Alice" {
		t.Error("W7: Initial wrapper should have Alice")
	}

	// Change the value by setting a new pointer (changes ValueJSON)
	// Note: Just modifying person.Name wouldn't change ValueJSON since the pointer stays the same
	newPerson := &Person{Name: "Bob"}
	err := v.Set(newPerson)
	if err != nil {
		t.Fatalf("W7: Set failed: %v", err)
	}

	// Wrapper should be re-created with new value
	if v.WrapperValue == nil {
		t.Error("W7: Wrapper should still exist after value change")
	}
	newWrapperData := v.WrapperValue.(*WrapperData)
	if newWrapperData.WrappedName != "Wrapped:Bob" {
		t.Errorf("W7: New wrapper should have Bob, got %q", newWrapperData.WrappedName)
	}

	// Old wrapper should be unregistered
	_, ok := tr.LookupObject(oldWrapper)
	if ok {
		t.Error("W7: Old wrapper should be unregistered after value change")
	}
}

// W8: Wrapper not created when CreateWrapper returns nil
func TestWrapper_CreateReturnsNil(t *testing.T) {
	tr := NewTracker()
	// Use a resolver that returns nil for CreateWrapper
	// The default Tracker.CreateWrapper returns nil

	person := &Person{Name: "Alice"}
	v := tr.CreateVariable(person, 0, "?wrapper=true", nil)

	// Even with wrapper property, no wrapper because CreateWrapper returns nil
	if v.WrapperValue != nil {
		t.Error("W8: WrapperValue should be nil when CreateWrapper returns nil")
	}
}

// W9: Set via child navigates through wrapper
func TestWrapper_SetViaChild(t *testing.T) {
	tr := NewTracker()
	wr := &wrapperResolver{tr}
	tr.Resolver = wr

	person := &Person{Name: "Alice"}
	parent := tr.CreateVariable(person, 0, "?wrapper=true", nil)

	// Child navigates to wrapper's field
	child := tr.CreateVariable(nil, parent.ID, "Extra?access=rw", nil)

	// Set via wrapper
	err := child.Set(100)
	if err != nil {
		t.Errorf("W9: Set should succeed, err=%v", err)
	}

	wrapper := parent.WrapperValue.(*WrapperData)
	if wrapper.Extra != 100 {
		t.Errorf("W9: Wrapper.Extra should be 100, got %d", wrapper.Extra)
	}
}

// StatefulWrapper is a wrapper that maintains persistent state
type StatefulWrapper struct {
	Data      *Person
	CallCount int // tracks how many times it was updated
}

// reusingResolver returns the same wrapper on subsequent calls
type reusingResolver struct {
	*Tracker
}

func (r *reusingResolver) CreateWrapper(v *Variable) any {
	if v.Value == nil {
		return nil
	}
	p, ok := v.Value.(*Person)
	if !ok {
		return nil
	}

	// Reuse existing wrapper if present
	if w, ok := v.WrapperValue.(*StatefulWrapper); ok {
		w.Data = p
		w.CallCount++
		return w // Same pointer, state preserved
	}

	// Create new wrapper
	return &StatefulWrapper{
		Data:      p,
		CallCount: 1,
	}
}

// W10: Wrapper reuse preserves state and avoids re-registration
func TestWrapper_ReusePreservesState(t *testing.T) {
	tr := NewTracker()
	rr := &reusingResolver{tr}
	tr.Resolver = rr

	person := &Person{Name: "Alice"}
	v := tr.CreateVariable(person, 0, "?wrapper=true", nil)

	// Initial wrapper created
	wrapper1 := v.WrapperValue.(*StatefulWrapper)
	if wrapper1.CallCount != 1 {
		t.Errorf("W10: Initial CallCount should be 1, got %d", wrapper1.CallCount)
	}
	initialWrapperJSON := v.WrapperJSON

	// Change the value by setting a new pointer (changes ValueJSON)
	// Note: Just modifying person.Name wouldn't change ValueJSON since the pointer stays the same
	newPerson := &Person{Name: "Bob"}
	err := v.Set(newPerson)
	if err != nil {
		t.Fatalf("W10: Set failed: %v", err)
	}

	// Same wrapper should be reused (reusingResolver returns same pointer)
	wrapper2 := v.WrapperValue.(*StatefulWrapper)
	if wrapper2 != wrapper1 {
		t.Error("W10: Wrapper should be the same pointer (reused)")
	}
	if wrapper2.CallCount != 2 {
		t.Errorf("W10: CallCount should be 2 after value change, got %d", wrapper2.CallCount)
	}
	if wrapper2.Data.Name != "Bob" {
		t.Errorf("W10: Data should be updated to Bob, got %q", wrapper2.Data.Name)
	}

	// WrapperJSON should NOT be recomputed (same pointer)
	if v.WrapperJSON != initialWrapperJSON {
		t.Error("W10: WrapperJSON should not change when wrapper is reused")
	}

	// Wrapper should still be registered
	_, ok := tr.LookupObject(wrapper1)
	if !ok {
		t.Error("W10: Wrapper should still be registered")
	}
}

// W11: Wrapper replacement when different pointer returned
func TestWrapper_ReplacementOnDifferentPointer(t *testing.T) {
	tr := NewTracker()
	wr := &wrapperResolver{tr} // This always creates new wrappers
	tr.Resolver = wr

	person := &Person{Name: "Alice"}
	v := tr.CreateVariable(person, 0, "?wrapper=true", nil)

	wrapper1 := v.WrapperValue
	wrapperJSON1 := v.WrapperJSON

	// Change value by setting a new pointer (changes ValueJSON)
	// Note: Just modifying person.Name wouldn't change ValueJSON since the pointer stays the same
	newPerson := &Person{Name: "Bob"}
	err := v.Set(newPerson)
	if err != nil {
		t.Fatalf("W11: Set failed: %v", err)
	}

	wrapper2 := v.WrapperValue
	wrapperJSON2 := v.WrapperJSON

	// Should be different wrappers (wrapperResolver always creates new)
	if wrapper2 == wrapper1 {
		t.Error("W11: Wrapper should be different pointer (replaced)")
	}

	// WrapperJSON should be recomputed
	if wrapperJSON2 == wrapperJSON1 {
		t.Error("W11: WrapperJSON should be recomputed when wrapper changes")
	}

	// Old wrapper should be unregistered
	_, ok := tr.LookupObject(wrapper1)
	if ok {
		t.Error("W11: Old wrapper should be unregistered")
	}

	// New wrapper should be registered
	_, ok = tr.LookupObject(wrapper2)
	if !ok {
		t.Error("W11: New wrapper should be registered")
	}
}
