# Test Design: Tracker
**Source Design:** crc-Tracker.md

## Test Scenarios

### NewTracker
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| T1.1 | Create new tracker | (none) | Non-nil tracker with Resolver set to self |
| T1.2 | Initial state | NewTracker() | Empty variables, empty changed set, empty rootIDs, nextID=1 |

### CreateVariable
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| T2.1 | Root variable with value | (value, 0, "", nil) | Variable with ID=1, Value cached, ValueJSON set |
| T2.2 | Root variable with properties | (value, 0, "", props) | Variable with Properties set |
| T2.3 | Child variable with path arg | (nil, parentID, "Field", nil) | Variable navigates to field value |
| T2.4 | Child with dot path | (nil, parentID, "A.B", nil) | Variable navigates nested path |
| T2.5 | Child with index path | (nil, parentID, "0", nil) | Variable navigates to slice element |
| T2.6 | Pointer value registered | (*struct, 0, "", nil) | Object in registry, ValueJSON = ObjectRef |
| T2.7 | Map value registered | (map, 0, "", nil) | Object in registry, ValueJSON = ObjectRef |
| T2.8 | Sequential IDs | Create 3 variables | IDs are 1, 2, 3 |
| T2.9 | Nil properties | (value, 0, "", nil) | Properties is empty map (not nil) |
| T2.10 | Empty path uses props | (nil, parentID, "", {"path": "Field"}) | Uses path from properties |
| T2.11 | Path arg overrides props | (nil, parentID, "A", {"path": "B"}) | Uses "A", ignores "B" |
| T2.12 | Path with query params | (nil, parentID, "a.b?w=1&h=2", nil) | Path=["a","b"], props["w"]="1", props["h"]="2" |
| T2.13 | Query props override map | (nil, parentID, "a?x=1", {"x": "2"}) | props["x"]="1" (query wins) |
| T2.14 | Path with priority prop | (nil, parentID, "a?priority=high", nil) | ValuePriority = PriorityHigh |
| T2.15 | Priority from props map | (nil, parentID, "a", {"priority": "low"}) | ValuePriority = PriorityLow |
| T2.16 | Active defaults to true | CreateVariable | v.Active == true |
| T2.17 | Root added to rootIDs | (value, 0, "", nil) | rootIDs contains v.ID |
| T2.18 | Child not in rootIDs | (nil, parentID, "path", nil) | rootIDs does not contain v.ID |
| T2.19 | Child added to parent ChildIDs | (nil, parentID, "path", nil) | parent.ChildIDs contains v.ID |
| T2.20 | rw rejects () path | (nil, parentID, "Value()?access=rw", nil) | error returned |
| T2.21 | rw rejects (_) path | (nil, parentID, "SetX(_)?access=rw", nil) | error returned |
| T2.22 | r allows () path | (nil, parentID, "Value()?access=r", nil) | Variable created |
| T2.23 | r rejects (_) path | (nil, parentID, "SetX(_)?access=r", nil) | error returned |
| T2.24 | w allows (_) path | (nil, parentID, "SetX(_)?access=w", nil) | Variable created |
| T2.25 | w rejects () path | (nil, parentID, "Value()?access=w", nil) | error returned |
| T2.26 | action allows () path | (nil, parentID, "Trigger()?access=action", nil) | Variable created |
| T2.27 | action allows (_) path | (nil, parentID, "AddItem(_)?access=action", nil) | Variable created |
| T2.28 | default (rw) rejects () | (nil, parentID, "Value()", nil) | error returned |
| T2.29 | default (rw) rejects (_) | (nil, parentID, "SetX(_)", nil) | error returned |

### CreateVariableWithId
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| T2.30 | Caller-specified ID | (5, value, 0, "", nil) | Variable with ID=5 |
| T2.31 | ID in use | (1, value, 0, "", nil) after another | Returns nil |
| T2.32 | Does not touch nextID | (10, value, 0, "", nil) | nextID unchanged |
| T2.33 | CreateVariable increments nextID | Create 2 variables | IDs are 1, 2; nextID becomes 3 |

### GetVariable
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| T3.1 | Existing variable | valid ID | Returns *Variable |
| T3.2 | Non-existent ID | invalid ID | Returns nil |
| T3.3 | After destroy | destroyed ID | Returns nil |

### DestroyVariable
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| T4.1 | Destroy existing | valid ID | Variable removed from tracker |
| T4.2 | Object unregistered | ID with pointer | Object removed from registry |
| T4.3 | Removed from changed | ID in changed set | ID removed from changed set |
| T4.4 | Destroy non-existent | invalid ID | No error (no-op) |
| T4.5 | Root removed from rootIDs | destroy root | rootIDs no longer contains ID |
| T4.6 | Child removed from ChildIDs | destroy child | parent.ChildIDs no longer contains ID |

### DetectChanges
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| T5.1 | No changes | unmodified values | returns empty []Change |
| T5.2 | Primitive change | modify int field | returns []Change with variable ID |
| T5.3 | String change | modify string field | returns []Change with variable ID |
| T5.4 | Multiple changes | modify 2 fields | returns []Change with both IDs |
| T5.5 | Nested change | modify nested field | returns []Change with child variable ID |
| T5.6 | Array element change | modify slice element | returns []Change with variable ID |
| T5.7 | Map value change | modify map value | returns []Change with variable ID |
| T5.8 | No false positives | same value reassigned | returns empty []Change |
| T5.9 | ValueJSON updated | after detection | ValueJSON reflects current |
| T5.10 | Clears internal state | after call | valueChanges and propertyChanges empty |
| T5.11 | Returns sorted changes | high and low priority | high priority changes first |
| T5.12 | Property changes included | SetProperty before DetectChanges | returns []Change with property |
| T5.13 | Slice reuse | call twice | same backing array capacity |
| T5.14 | Split by priority | high value, low property same var | 2 Change entries |
| T5.15 | Group same priority | 2 props same priority | 1 Change with both props |
| T5.16 | Tree traversal | multi-level tree | detects all changes via DFS |
| T5.17 | Skip inactive | inactive variable | not detected |
| T5.18 | Skip descendants | inactive parent | children not detected |
| T5.19 | Multiple roots | 2 root trees | both traversed |

### Variables
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| T8.1 | All variables | 3 created | slice of length 3 |
| T8.2 | Empty tracker | none created | empty slice |

### RootVariables
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| T9.1 | Only roots | mix of root/child | only parentID=0 vars |

### Children
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| T10.1 | Get children | parent with 2 children | slice of 2 children |
| T10.2 | No children | parent with no children | empty slice |

## Error Scenarios

| ID | Scenario | Input | Expected Error |
|----|----------|-------|----------------|
| E1 | Invalid parent ID | CreateVariable with bad parentID | error or nil parent handling |
| E2 | Path navigation failure | invalid path element | error from Get |
| E3 | rw access with () path | access: "rw", path: "Value()" | OK (variadic call supported) |
| E4 | rw access with (_) path | access: "rw", path: "SetX(_)" | error: cannot read from setter |
| E5 | r access with (_) path | access: "r", path: "SetX(_)" | error: cannot read from setter |
| E6 | w access with () path | access: "w", path: "Value()" | error: use rw, r, or action for zero-arg methods |

## Integration Tests

| ID | Scenario | Description |
|----|----------|-------------|
| I1 | Full lifecycle | Create, modify, DetectChanges (returns sorted, auto-clears) |
| I2 | Parent-child tree | Multi-level variable hierarchy |
| I3 | Object identity | Same pointer in multiple variables |
