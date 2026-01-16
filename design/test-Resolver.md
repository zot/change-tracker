# Test Design: Resolver
**Source Design:** crc-Resolver.md

## Test Scenarios (Default Tracker Implementation)

### Get - Struct Fields
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| R1.1 | Exported field | Get(struct, "Name") | field value |
| R1.2 | Nested struct | Get(struct, "Address") | Address struct |
| R1.3 | Pointer to struct | Get(*struct, "Name") | field value |
| R1.4 | Various types | Get(struct, "IntField") | int value |

### Get - Map Keys
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| R2.1 | String key exists | Get(map, "key") | map value |
| R2.2 | String key missing | Get(map, "missing") | error |
| R2.3 | Map of pointers | Get(map[string]*T, "k") | *T value |

### Get - Slice/Array Index
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| R3.1 | Valid index | Get(slice, 0) | first element |
| R3.2 | Middle index | Get(slice, 1) | second element |
| R3.3 | Array access | Get(array, 0) | first element |

### Set - Struct Fields
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| S1.1 | Set via pointer | Set(*struct, "Name", "new") | field updated |
| S1.2 | Set int field | Set(*struct, "Count", 42) | field updated |
| S1.3 | Set nested field | Set(*struct.Addr, "City", "NYC") | nested updated |

### Set - Map Keys
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| S2.1 | Set existing key | Set(map, "key", "val") | value updated |
| S2.2 | Set new key | Set(map, "new", "val") | key added |

### Set - Slice Index
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| S3.1 | Set valid index | Set(slice, 0, "new") | element updated |
| S3.2 | Set middle index | Set(slice, 1, "new") | element updated |

### Call - Zero-Arg Methods
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| C1.1 | Zero-arg method | Call(obj, "Value") | method return value |
| C1.2 | Method on pointer | Call(*obj, "Method") | return value |
| C1.3 | Multi-return method | Call(obj, "Pair") | first return value |
| C1.4 | String return | Call(obj, "Name") | string value |
| C1.5 | Struct return | Call(obj, "Address") | struct value |

### CallWith - One-Arg Methods
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| CW1.1 | Set int via method | CallWith(obj, "SetValue", 42) | nil, value updated |
| CW1.2 | Set string via method | CallWith(obj, "SetName", "Bob") | nil, value updated |
| CW1.3 | Method on pointer | CallWith(*obj, "SetCount", 5) | nil, value updated |
| CW1.4 | Set struct via method | CallWith(obj, "SetAddr", addr) | nil, value updated |

## Error Scenarios

### Get Errors
| ID | Scenario | Input | Expected Error |
|----|----------|-------|----------------|
| GE1 | Nil object | Get(nil, "field") | "nil object" |
| GE2 | Unexported field | Get(struct, "private") | "not found" or "unexported" |
| GE3 | Missing field | Get(struct, "NoSuch") | "not found" |
| GE4 | Missing map key | Get(map, "missing") | "key not found" |
| GE5 | Index out of bounds | Get(slice, 100) | "out of bounds" |
| GE6 | Unsupported type | Get(int, "field") | "unsupported type" |

### Set Errors
| ID | Scenario | Input | Expected Error |
|----|----------|-------|----------------|
| SE1 | Nil object | Set(nil, "f", v) | "nil object" |
| SE2 | Non-pointer struct | Set(struct, "f", v) | "need pointer" |
| SE3 | Unexported field | Set(*s, "private", v) | "not settable" |
| SE4 | Missing field | Set(*s, "NoSuch", v) | "not found" |
| SE5 | Type mismatch | Set(*s, "IntF", "str") | "type mismatch" |
| SE6 | Index out of bounds | Set(slice, 100, v) | "out of bounds" |

### Call Errors
| ID | Scenario | Input | Expected Error |
|----|----------|-------|----------------|
| CE1 | Nil object | Call(nil, "Method") | "nil object" |
| CE2 | Method not found | Call(obj, "NoSuch") | "method not found" |
| CE3 | Unexported method | Call(obj, "private") | "method not found" or "unexported" |
| CE4 | Method needs args | Call(obj, "NeedsArg") | "requires arguments" |
| CE5 | Method returns nothing | Call(obj, "VoidMethod") | "returns no values" |

### CallWith Errors
| ID | Scenario | Input | Expected Error |
|----|----------|-------|----------------|
| CWE1 | Nil object | CallWith(nil, "Set", v) | "nil object" |
| CWE2 | Method not found | CallWith(obj, "NoSuch", v) | "method not found" |
| CWE3 | Unexported method | CallWith(obj, "setPrivate", v) | "method not found" or "unexported" |
| CWE4 | Method takes no args | CallWith(obj, "NoArgs", v) | "doesn't take one argument" |
| CWE5 | Method takes 2+ args | CallWith(obj, "TwoArgs", v) | "doesn't take one argument" |
| CWE6 | Method returns value | CallWith(obj, "ReturnsSomething", v) | OK (return ignored) |
| CWE7 | Arg type mismatch | CallWith(obj, "SetInt", "str") | "type mismatch" |

## Path Semantics Tests

### Path with Method Calls
| ID | Scenario | Path | Operation | Expected |
|----|----------|------|-----------|----------|
| PM1 | Getter mid-path | "Address().City" | Get | OK, returns city |
| PM2 | Getter mid-path | "Address().City" | Set | OK, sets city |
| PM3 | Path ends in getter | "Value()" | Get | OK, returns value |
| PM4 | Path ends in getter (rw) | "Value()?access=rw" | Set | OK, calls method with args |
| PM5 | Path ends in setter | "SetValue(_)" | Set | OK, calls setter |
| PM6 | Path ends in setter | "SetValue(_)" | Get | ERROR, write-only |
| PM7 | Setter not terminal | "SetAddr(_).City" | Get | ERROR, setter must be terminal |
| PM8 | Setter not terminal | "SetAddr(_).City" | Set | ERROR, setter must be terminal |
| PM9 | Chain getters | "Outer().Inner().Value()" | Get | OK, chains calls |
| PM10 | Path ends in getter (r) | "Value()?access=r" | Set | ERROR, read-only access |

## Access Property Tests

### Variable Access Modes
| ID | Scenario | Access | Operation | Expected |
|----|----------|--------|-----------|----------|
| AM1 | Read-write (default) | "rw" | Get | OK |
| AM2 | Read-write (default) | "rw" | Set | OK |
| AM3 | Read-only | "r" | Get | OK |
| AM4 | Read-only | "r" | Set | ERROR, read-only variable |
| AM5 | Write-only | "w" | Get | ERROR, write-only variable |
| AM6 | Write-only | "w" | Set | OK |
| AM7 | Invalid access | "x" | SetProperty | ERROR, invalid access value |
| AM8 | Empty access | "" | behavior | defaults to "rw" |
| AM9 | Action | "action" | Get | ERROR, action variable |
| AM10 | Action | "action" | Set | OK |

### Access with Path Semantics (Combined)
| ID | Scenario | Access | Path | Get | Set |
|----|----------|--------|------|-----|-----|
| AP1 | rw + field | "rw" | "Field" | OK | OK |
| AP2 | r + field | "r" | "Field" | OK | ERROR (access) |
| AP3 | w + field | "w" | "Field" | ERROR (access) | OK |
| AP4 | rw + getter | "rw" | "Value()" | OK | OK (variadic call) |
| AP5 | r + getter | "r" | "Value()" | OK | ERROR (access) |
| AP6 | w + getter | "w" | "Value()" | N/A - CreateVariable fails | N/A |
| AP7 | rw + setter | "rw" | "SetX(_)" | N/A - CreateVariable fails | N/A |
| AP8 | r + setter | "r" | "SetX(_)" | N/A - CreateVariable fails | N/A |
| AP9 | w + setter | "w" | "SetX(_)" | ERROR (access+path) | OK |
| AP10 | action + field | "action" | "Field" | ERROR (access) | OK |
| AP11 | action + getter | "action" | "Value()" | ERROR (access) | OK (side effect) |
| AP12 | action + setter | "action" | "SetX(_)" | ERROR (access) | OK |

Note: Path restrictions are validated at CreateVariable time. Combinations marked "N/A - CreateVariable fails" will return an error before the variable is created. `rw`, `r`, and `action` allow `()` path endings. Only `w` and `action` allow `(_)` path endings.

### Write-Only and Action Method Side Effects
| ID | Scenario | Access | Path | Set Behavior |
|----|----------|--------|------|--------------|
| SE1 | Write-only with setter | "w" | "SetValue(_)" | Calls SetValue(arg) |
| SE2 | Write-only with field | "w" | "Field" | Sets field value |
| SE3 | Trigger side effect (action) | "action" | "Trigger()" | Calls Trigger(), side effect occurs |
| SE4 | Action with arg | "action" | "AddItem(_)" | Calls AddItem(arg) |
| SE5 | Action void method | "action" | "DoAction()" | Calls method, no error |
| SE6 | Action method with return | "action" | "Process()" | Calls method, return value ignored |

Note: Write-only (`w`) access can no longer use `()` paths - use `action` access for zero-arg method triggers.

### Access and Change Detection
| ID | Scenario | Access | Expected in DetectChanges |
|----|----------|--------|---------------------------|
| AD1 | Read-write variable | "rw" | Yes, scanned |
| AD2 | Read-only variable | "r" | Yes, scanned |
| AD3 | Write-only variable | "w" | No, skipped |
| AD4 | Write-only parent, rw child | parent "w", child "rw" | Child scanned, parent not |
| AD5 | Action variable | "action" | No, skipped |
| AD6 | Action parent, rw child | parent "action", child "rw" | Child scanned, parent not |

### Action vs Write-Only Creation Behavior
| ID | Scenario | Access | Expected at CreateVariable |
|----|----------|--------|----------------------------|
| AC1 | Write-only initial value | "w" | Initial value IS computed |
| AC2 | Action no initial value | "action" | Initial value NOT computed |
| AC3 | Write-only ValueJSON set | "w" | ValueJSON IS set |
| AC4 | Action ValueJSON nil | "action" | ValueJSON is nil |
| AC5 | Action avoids method call | "action", path: "Trigger()" | Method NOT called at creation |
| AC6 | Write-only with field path | "w", path: "Field" | Get() called at creation |

### Path Restriction Validation at CreateVariable
| ID | Scenario | Access | Path | Expected |
|----|----------|--------|------|----------|
| PR1 | rw allows () path | "rw" | "Value()" | OK |
| PR2 | rw rejects (_) path | "rw" | "SetValue(_)" | ERROR: cannot read from setter |
| PR3 | r allows () path | "r" | "Value()" | OK |
| PR4 | r rejects (_) path | "r" | "SetValue(_)" | ERROR: cannot read from setter |
| PR5 | w rejects () path | "w" | "Value()" | ERROR: use rw, r, or action for zero-arg methods |
| PR6 | w allows (_) path | "w" | "SetValue(_)" | OK |
| PR7 | action allows () path | "action" | "Trigger()" | OK |
| PR8 | action allows (_) path | "action" | "AddItem(_)" | OK |
| PR9 | default (rw) allows () | (none) | "Value()" | OK |
| PR10 | default (rw) rejects (_) | (none) | "SetX(_)" | ERROR |
| PR11 | nested path ending in () | "rw" | "Obj.Value()" | OK |
| PR12 | nested path ending in (_) | "r" | "Obj.SetX(_)" | ERROR |

## Path Element Type Tests

| ID | Scenario | Path Element | Object Type | Expected Behavior |
|----|----------|--------------|-------------|-------------------|
| PT1 | String on struct | "Field" | struct | field access |
| PT2 | String on map | "key" | map | key lookup |
| PT3 | String with "()" | "Method()" | any | Call invocation |
| PT4 | String with "(_)" | "SetMethod(_)" | any | CallWith invocation |
| PT5 | Int on slice | 0 | slice | index access |
| PT6 | Int on array | 0 | array | index access |
| PT7 | Invalid type | float64 | any | error |
