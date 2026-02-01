# Requirements

## Feature: Core Tracker
**Source:** specs/main.md

- **R1:** Tracker creates and manages variables with unique integer IDs
- **R2:** Tracker maintains a set of root variable IDs (variables with ParentID == 0)
- **R3:** Tracker maintains an object registry (weak map from objects to IDs)
- **R4:** Tracker tracks which variables have changed
- **R5:** Tracker has a Resolver field that defaults to itself (using Go reflection)
- **R6:** Tracker serializes values to Value JSON form

## Feature: Variables
**Source:** specs/main.md

- **R7:** Variable has ID, ParentID, ChildIDs, Active, Access, Properties, PropertyPriorities, Path, Value, ValueJSON, ValuePriority fields
- **R8:** Variable value is computed by starting at parent's cached value and applying each path element
- **R9:** Variables form a tree structure via parent-child relationships
- **R10:** ChildIDs is maintained automatically when variables are created/destroyed

## Feature: Variable Errors
**Source:** specs/api.md

- **R11:** Variable has Error field to store structured error from last Get/Set operation

## Feature: Active Flag
**Source:** specs/main.md

- **R12:** Active field controls whether a variable participates in change detection (default: true)
- **R13:** When Active is false, variable and all descendants are skipped during change detection
- **R14:** Active field can be toggled at any time; changes take effect on next DetectChanges()

## Feature: Access Modes
**Source:** specs/main.md

- **R15:** Access mode "rw" (default) allows Get, Set, and change detection; initial value computed
- **R16:** Access mode "r" allows Get and change detection; Set fails; initial value computed
- **R17:** Access mode "w" allows Set only; Get fails; excluded from change detection; initial value computed
- **R18:** Access mode "action" allows Set only; Get fails; excluded from change detection; initial value NOT computed
- **R19:** Path ending in `(_)` requires access "w" or "action"
- **R20:** Path ending in `()` is allowed with access "rw", "r", or "action"

## Feature: Priorities
**Source:** specs/main.md

- **R21:** Values and properties can have priority levels: Low, Medium (default), High
- **R22:** Value priority is set via the "priority" property
- **R23:** Property priority is set via suffix on property name (:low, :medium, :high)
- **R24:** Properties without priority suffix default to Medium

## Feature: Object Registry
**Source:** specs/main.md

- **R25:** Object registry uses Go 1.24+ weak references (weak.Pointer)
- **R26:** Objects registered automatically via ToValueJSON() when pointers/maps encountered
- **R27:** Registered objects don't prevent garbage collection
- **R28:** When object is collected, registry entry is automatically cleaned up
- **R29:** Same object in multiple locations serializes to same {"obj": id}

## Feature: Value JSON
**Source:** specs/value-json.md

- **R30:** Value JSON has three types: primitives, arrays, object references
- **R31:** Primitives (string, number, bool, nil) serialize as standard JSON
- **R32:** Arrays/slices serialize as JSON arrays with elements in Value JSON form
- **R33:** Registered objects (pointers, maps) serialize as {"obj": ID}
- **R34:** Unregistered pointers/maps are auto-registered during ToValueJSON()

## Feature: Change Detection
**Source:** specs/main.md

- **R35:** DetectChanges() performs depth-first traversal starting from root variables
- **R36:** For each active variable, convert current value to Value JSON and compare to stored
- **R37:** Inactive variables and their descendants are skipped
- **R38:** After comparison, current Value JSON becomes new stored Value JSON
- **R39:** Changes are sorted by priority (high → medium → low) and returned
- **R40:** Internal change records are cleared after DetectChanges()
- **R41:** Variable may appear multiple times if changes at different priority levels

## Feature: Property Changes
**Source:** specs/main.md

- **R42:** SetProperty() records changes in the tracker for DetectChanges
- **R43:** Setting "priority" property updates variable's ValuePriority
- **R44:** Setting "path" property re-parses and updates Path field
- **R45:** Setting "access" property updates Access field (validates: r, w, rw, action)

## Feature: Resolver Interface
**Source:** specs/resolver.md

- **R46:** Resolver interface has Get(obj, pathElement) and Set(obj, pathElement, value) methods
- **R47:** Resolver interface has Call(obj, methodName) for zero-arg methods
- **R48:** Resolver interface has CallWith(obj, methodName, value) for one-arg methods
- **R49:** Path elements can be string (field, map key, method) or int (index)
- **R50:** Method calls use "()" suffix for getters, "(_)" suffix for setters

## Feature: Default Resolver (Tracker)
**Source:** specs/resolver.md

- **R51:** Tracker implements Resolver using Go reflection
- **R52:** Get on struct returns field value by name (exported fields only)
- **R53:** Get on map returns value by string key
- **R54:** Get on slice/array returns element at integer index
- **R55:** Call invokes zero-arg method and returns first return value
- **R56:** Set on struct field requires pointer to struct
- **R57:** Set on slice requires index within bounds
- **R58:** CallWith invokes method with exactly one argument or variadic

## Feature: Structured Errors
**Source:** specs/api.md

- **R59:** VariableErrorType enum categorizes errors (PathError, NotFound, BadSetterCall, BadAccess, BadIndex, BadReference, BadParent, BadCall, NilPath)
- **R60:** VariableError has ErrorType, Message, and Cause fields
- **R61:** Variable.Error field stores last error from Get/Set operations
- **R62:** All resolver and variable operations return VariableError for failures

## Feature: Wrapper Support
**Source:** specs/wrapper.md

- **R63:** Variable can have optional wrapper via "wrapper" property
- **R64:** Resolver.CreateWrapper(variable) creates wrapper when wrapper property set and ValueJSON non-nil
- **R65:** Child variables navigate through WrapperValue instead of Value when wrapper present
- **R66:** NavigationValue() returns WrapperValue if present, otherwise Value
- **R67:** Wrapper is destroyed when wrapper property cleared, ValueJSON becomes nil, or variable destroyed
- **R68:** CreateWrapper can return same wrapper (preserves state) or new wrapper (replaces old)
- **R69:** WrapperValue and WrapperJSON fields store wrapper and its serialized form
