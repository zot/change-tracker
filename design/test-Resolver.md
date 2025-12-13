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
| CWE6 | Method returns value | CallWith(obj, "ReturnsSomething", v) | "must be void" |
| CWE7 | Arg type mismatch | CallWith(obj, "SetInt", "str") | "type mismatch" |

## Path Semantics Tests

### Path with Method Calls
| ID | Scenario | Path | Operation | Expected |
|----|----------|------|-----------|----------|
| PM1 | Getter mid-path | "Address().City" | Get | OK, returns city |
| PM2 | Getter mid-path | "Address().City" | Set | OK, sets city |
| PM3 | Path ends in getter | "Value()" | Get | OK, returns value |
| PM4 | Path ends in getter | "Value()" | Set | ERROR, read-only |
| PM5 | Path ends in setter | "SetValue(_)" | Set | OK, calls setter |
| PM6 | Path ends in setter | "SetValue(_)" | Get | ERROR, write-only |
| PM7 | Setter not terminal | "SetAddr(_).City" | Get | ERROR, setter must be terminal |
| PM8 | Setter not terminal | "SetAddr(_).City" | Set | ERROR, setter must be terminal |
| PM9 | Chain getters | "Outer().Inner().Value()" | Get | OK, chains calls |

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

### Access with Path Semantics (Combined)
| ID | Scenario | Access | Path | Get | Set |
|----|----------|--------|------|-----|-----|
| AP1 | rw + field | "rw" | "Field" | OK | OK |
| AP2 | r + field | "r" | "Field" | OK | ERROR (access) |
| AP3 | w + field | "w" | "Field" | ERROR (access) | OK |
| AP4 | rw + getter | "rw" | "Value()" | OK | ERROR (path) |
| AP5 | r + getter | "r" | "Value()" | OK | ERROR (access) |
| AP6 | w + getter | "w" | "Value()" | ERROR (access) | OK (side effect) |
| AP7 | rw + setter | "rw" | "SetX(_)" | ERROR (path) | OK |
| AP8 | r + setter | "r" | "SetX(_)" | ERROR (path) | ERROR (access) |
| AP9 | w + setter | "w" | "SetX(_)" | ERROR (access+path) | OK |

Note: For write-only variables (`access: "w"`), a `()` path allows Set to call the method for side effects (ignoring return value). The path-based Set restriction only applies to readable variables.

### Write-Only Method Side Effects
| ID | Scenario | Access | Path | Set Behavior |
|----|----------|--------|------|--------------|
| SE1 | Trigger side effect | "w" | "Trigger()" | Calls Trigger(), side effect occurs |
| SE2 | Counter increment | "w" | "Increment()" | Calls Increment(), counter updated |
| SE3 | Void method | "w" | "DoSomething()" | Calls method, no error |
| SE4 | Method with return | "w" | "Process()" | Calls method, return value ignored |

### Access and Change Detection
| ID | Scenario | Access | Expected in DetectChanges |
|----|----------|--------|---------------------------|
| AD1 | Read-write variable | "rw" | Yes, scanned |
| AD2 | Read-only variable | "r" | Yes, scanned |
| AD3 | Write-only variable | "w" | No, skipped |
| AD4 | Write-only parent, rw child | parent "w", child "rw" | Child scanned, parent not |

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
