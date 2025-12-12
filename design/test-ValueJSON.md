# Test Design: Value JSON
**Source Design:** crc-ObjectRef.md, seq-to-value-json.md

## Test Scenarios

### ToValueJSON - Primitives
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| VJ1.1 | Nil value | nil | nil |
| VJ1.2 | String | "hello" | "hello" |
| VJ1.3 | Int | 42 | 42 |
| VJ1.4 | Int64 | int64(100) | 100 |
| VJ1.5 | Float64 | 3.14 | 3.14 |
| VJ1.6 | Bool true | true | true |
| VJ1.7 | Bool false | false | false |

### ToValueJSON - Arrays
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| VJ2.1 | Empty slice | []int{} | [] |
| VJ2.2 | Int slice | []int{1,2,3} | [1,2,3] |
| VJ2.3 | String slice | []string{"a","b"} | ["a","b"] |
| VJ2.4 | Nested slice | [][]int{{1},{2}} | [[1],[2]] |
| VJ2.5 | Array type | [3]int{1,2,3} | [1,2,3] |
| VJ2.6 | Pointer slice | []*T{p1,p2} (registered) | [{"obj":1},{"obj":2}] |

### ToValueJSON - Object References
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| VJ3.1 | Registered pointer | *T (registered as 1) | ObjectRef{Obj:1} |
| VJ3.2 | Registered map | map (registered as 2) | ObjectRef{Obj:2} |
| VJ3.3 | Same ptr twice | []*T{p,p} | [{"obj":1},{"obj":1}] |
| VJ3.4 | Mixed array | []any{"s",42,ptr} | ["s",42,{"obj":1}] |

### ToValueJSONBytes
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| VJ4.1 | Primitive | 42 | []byte("42") |
| VJ4.2 | Object ref | registered ptr | []byte(`{"obj":1}`) |
| VJ4.3 | Array | []int{1,2} | []byte("[1,2]") |

### IsObjectRef
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| VJ5.1 | Is ObjectRef | ObjectRef{Obj:1} | true |
| VJ5.2 | Not ObjectRef int | 42 | false |
| VJ5.3 | Not ObjectRef string | "hello" | false |
| VJ5.4 | Not ObjectRef map | map[string]any | false |

### GetObjectRefID
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| VJ6.1 | Valid ObjectRef | ObjectRef{Obj:5} | (5, true) |
| VJ6.2 | Not ObjectRef | 42 | (0, false) |
| VJ6.3 | Zero ID ref | ObjectRef{Obj:0} | (0, true) |

## Error Scenarios

| ID | Scenario | Input | Expected Error |
|----|----------|-------|----------------|
| VJE1 | Unregistered pointer | *T (not registered) | "unregistered pointer" |
| VJE2 | Unregistered map | map (not registered) | "unregistered map" |
| VJE3 | Struct value | T{} (not pointer) | treated as? (define behavior) |

## Change Detection via Value JSON

| ID | Scenario | Before | After | Changed? |
|----|----------|--------|-------|----------|
| CD1 | Same primitive | 42 | 42 | No |
| CD2 | Different primitive | 42 | 43 | Yes |
| CD3 | Same string | "a" | "a" | No |
| CD4 | Different string | "a" | "b" | Yes |
| CD5 | Same object ref | {"obj":1} | {"obj":1} | No |
| CD6 | Same array | [1,2] | [1,2] | No |
| CD7 | Different array len | [1,2] | [1,2,3] | Yes |
| CD8 | Different array elem | [1,2] | [1,3] | Yes |
| CD9 | Array order matters | [1,2] | [2,1] | Yes |

## JSON Serialization Format

| Go Type | Value JSON | JSON String |
|---------|------------|-------------|
| nil | nil | "null" |
| "text" | "text" | "\"text\"" |
| 42 | 42 | "42" |
| true | true | "true" |
| []int{1,2} | [1,2] | "[1,2]" |
| *T (id=5) | ObjectRef{Obj:5} | "{\"obj\":5}" |
