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

### Get - Method Calls
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| R4.1 | Zero-arg method | Get(obj, "Value()") | method return value |
| R4.2 | Method on pointer | Get(*obj, "Method()") | return value |
| R4.3 | Multi-return method | Get(obj, "Pair()") | first return value |

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

## Error Scenarios

### Get Errors
| ID | Scenario | Input | Expected Error |
|----|----------|-------|----------------|
| GE1 | Nil object | Get(nil, "field") | "nil object" |
| GE2 | Unexported field | Get(struct, "private") | "not found" or "unexported" |
| GE3 | Missing field | Get(struct, "NoSuch") | "not found" |
| GE4 | Missing map key | Get(map, "missing") | "key not found" |
| GE5 | Index out of bounds | Get(slice, 100) | "out of bounds" |
| GE6 | Method not found | Get(obj, "NoMethod()") | "method not found" |
| GE7 | Method needs args | Get(obj, "NeedsArg()") | "requires arguments" |
| GE8 | Unsupported type | Get(int, "field") | "unsupported type" |

### Set Errors
| ID | Scenario | Input | Expected Error |
|----|----------|-------|----------------|
| SE1 | Nil object | Set(nil, "f", v) | "nil object" |
| SE2 | Non-pointer struct | Set(struct, "f", v) | "need pointer" |
| SE3 | Unexported field | Set(*s, "private", v) | "not settable" |
| SE4 | Missing field | Set(*s, "NoSuch", v) | "not found" |
| SE5 | Type mismatch | Set(*s, "IntF", "str") | "type mismatch" |
| SE6 | Index out of bounds | Set(slice, 100, v) | "out of bounds" |
| SE7 | Set on method | Set(obj, "Method()", v) | "cannot set method" |

## Path Element Type Tests

| ID | Scenario | Path Element | Object Type | Expected Behavior |
|----|----------|--------------|-------------|-------------------|
| PT1 | String on struct | "Field" | struct | field access |
| PT2 | String on map | "key" | map | key lookup |
| PT3 | String with parens | "Method()" | any | method call |
| PT4 | Int on slice | 0 | slice | index access |
| PT5 | Int on array | 0 | array | index access |
| PT6 | Invalid type | float64 | any | error |
