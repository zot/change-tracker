# Test Design: Variable
**Source Design:** crc-Variable.md

## Test Scenarios

### Get
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| V1.1 | Root variable get | root var | returns cached Value |
| V1.2 | Child field get | child with "Field" path | returns field value |
| V1.3 | Child nested get | child with "A.B" path | returns nested value |
| V1.4 | Child index get | child with "0" path | returns slice[0] |
| V1.5 | Method call get | child with "Method()" path | returns method result |
| V1.6 | Map key get | child with "key" path on map | returns map value |
| V1.7 | Value caching | after Get() | Value field updated |
| V1.8 | Chained children | grandchild variable | navigates through parent |

### Set
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| V2.1 | Set struct field | Set("newval") on field var | field updated |
| V2.2 | Set nested field | Set on A.B path | nested field updated |
| V2.3 | Set slice element | Set on index path | slice element updated |
| V2.4 | Set map value | Set on map key path | map value updated |
| V2.5 | Verify change | Set then DetectChanges | variable in changed set |

### Parent
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| V3.1 | Root has no parent | root var Parent() | returns nil |
| V3.2 | Child has parent | child var Parent() | returns parent *Variable |

### GetProperty / SetProperty
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| V4.1 | Get existing | GetProperty("key") | returns value |
| V4.2 | Get non-existent | GetProperty("missing") | returns "" |
| V4.3 | Set property | SetProperty("key", "val") | property set |
| V4.4 | Remove property | SetProperty("key", "") | property removed |
| V4.5 | Set with :low suffix | SetProperty("label:low", "x") | Properties["label"]="x", priority=Low |
| V4.6 | Set with :medium suffix | SetProperty("label:medium", "x") | Properties["label"]="x", priority=Medium |
| V4.7 | Set with :high suffix | SetProperty("label:high", "x") | Properties["label"]="x", priority=High |
| V4.8 | Set without suffix | SetProperty("label", "x") | Properties["label"]="x", priority=Medium (default) |
| V4.9 | Remove with suffix | SetProperty("label:high", "") | property and priority removed |
| V4.10 | Set priority property | SetProperty("priority", "high") | ValuePriority = PriorityHigh |
| V4.11 | Set priority:high prop | SetProperty("priority:high", "low") | ValuePriority=Low, PropertyPriorities["priority"]=High |
| V4.12 | Set path property | SetProperty("path", "A.B") | Path = ["A", "B"] |
| V4.13 | Records property change | SetProperty("key", "val") | tracker.propertyChanges has "key" |
| V4.14 | Adds to changed set | SetProperty("key", "val") | var ID in DetectChanges() result |

### GetPropertyPriority
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| V5.1 | Get existing priority | After SetProperty("x:high","v") | GetPropertyPriority("x") = High |
| V5.2 | Get default priority | After SetProperty("x","v") | GetPropertyPriority("x") = Medium |
| V5.3 | Get non-existent | GetPropertyPriority("missing") | returns PriorityMedium |

### Active/SetActive
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| V6.1 | Active defaults to true | CreateVariable | v.Active == true |
| V6.2 | SetActive false | v.SetActive(false) | v.Active == false |
| V6.3 | SetActive true | v.SetActive(true) | v.Active == true |
| V6.4 | Inactive skipped | SetActive(false), DetectChanges | variable not in changes |
| V6.5 | Inactive skips descendants | parent SetActive(false) | children not in changes |
| V6.6 | Re-activate | SetActive(false) then true | variable in changes again |

### Access Property
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| V8.1 | Access defaults to rw | CreateVariable | GetAccess() == "rw" |
| V8.2 | Set access via property | SetProperty("access", "r") | GetAccess() == "r" |
| V8.3 | Set write-only access | SetProperty("access", "w") | GetAccess() == "w" |
| V8.4 | Read-only Get succeeds | access: "r", Get() | returns value |
| V8.5 | Read-only Set fails | access: "r", Set(v) | error: read-only |
| V8.6 | Write-only Get fails | access: "w", Get() | error: write-only |
| V8.7 | Write-only Set succeeds | access: "w", Set(v) | value updated |
| V8.8 | Read-write Get succeeds | access: "rw", Get() | returns value |
| V8.9 | Read-write Set succeeds | access: "rw", Set(v) | value updated |
| V8.10 | Invalid access value | SetProperty("access", "x") | error: invalid access |
| V8.11 | IsReadable for rw | access: "rw" | IsReadable() == true |
| V8.12 | IsReadable for r | access: "r" | IsReadable() == true |
| V8.13 | IsReadable for w | access: "w" | IsReadable() == false |
| V8.14 | IsWritable for rw | access: "rw" | IsWritable() == true |
| V8.15 | IsWritable for w | access: "w" | IsWritable() == true |
| V8.16 | IsWritable for r | access: "r" | IsWritable() == false |
| V8.17 | Access via query string | path: "Field?access=r" | GetAccess() == "r" |

### Access and Change Detection
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| V9.1 | Read-only scanned | access: "r", value changes | appears in DetectChanges |
| V9.2 | Write-only not scanned | access: "w", value changes | NOT in DetectChanges |
| V9.3 | Read-write scanned | access: "rw", value changes | appears in DetectChanges |
| V9.4 | Write-only children scanned | parent access: "w", child access: "rw" | child in DetectChanges |
| V9.5 | Access + Active combo | access: "r", Active: false | NOT in DetectChanges |

### Access vs Path Semantics
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| V10.1 | Access r + path () | access: "r", path: "Value()" | Get: OK, Set: error (access) |
| V10.2 | Access w + path () | access: "w", path: "Value()" | Get: error (access), Set: OK (calls method) |
| V10.3 | Access r + path (_) | access: "r", path: "SetX(_)" | Get: error (path), Set: error (access) |
| V10.4 | Access w + path (_) | access: "w", path: "SetX(_)" | Get: error (both), Set: OK |
| V10.5 | Access rw + path () | access: "rw", path: "Value()" | Get: OK, Set: error (path restricts for readable vars) |
| V10.6 | Write-only method side effect | access: "w", path: "Trigger()" | Set: calls Trigger(), side effect occurs |

### ChildIDs
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| V7.1 | Empty on creation | CreateVariable (no children) | ChildIDs is nil or empty |
| V7.2 | Populated on child creation | CreateVariable with parent | parent.ChildIDs contains child ID |
| V7.3 | Removed on destroy | DestroyVariable(child) | parent.ChildIDs no longer contains ID |

## Error Scenarios

| ID | Scenario | Input | Expected Error |
|----|----------|-------|----------------|
| E1 | Get invalid path | non-existent field | error returned |
| E2 | Get unexported field | unexported field name | error returned |
| E3 | Get index out of bounds | index >= len(slice) | error returned |
| E4 | Set on nil parent value | parent value is nil | error returned |
| E5 | Set type mismatch | wrong type for field | error returned |
| E6 | Set non-settable | unexported field | error returned |

## Path Parsing Tests

| ID | Scenario | Input Path | Expected Path Elements |
|----|----------|------------|------------------------|
| P1 | Simple field | "Name" | ["Name"] |
| P2 | Dot-separated | "Address.City" | ["Address", "City"] |
| P3 | Integer index | "0" | [0] |
| P4 | Mixed path | "Items.0.Name" | ["Items", 0, "Name"] |
| P5 | Method call | "GetValue()" | ["GetValue()"] |
| P6 | Empty path | "" | [] |
| P7 | Path with query | "a.b?x=1" | ["a", "b"] (query in props) |
| P8 | Query only | "?x=1&y=2" | [] (empty path, props set) |
| P9 | Multiple query params | "a?x=1&y=2&z=3" | ["a"], props["x"]="1", etc. |
| P10 | Priority in query | "a?priority=high" | ["a"], ValuePriority=High |
