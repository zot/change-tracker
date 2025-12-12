# Test Design: Change
**Source Design:** crc-Change.md

## Test Scenarios

### Change Structure
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| C1.1 | Create Change | Change{VariableID: 1, Priority: High, ValueChanged: true} | Fields accessible |
| C1.2 | Change with properties | Change{PropertiesChanged: []string{"a", "b"}} | PropertiesChanged contains ["a", "b"] |
| C1.3 | Empty properties | Change{PropertiesChanged: nil} | PropertiesChanged is nil |

## Integration Tests (via Tracker.DetectChanges)

### DetectChanges Basic
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| DC1.1 | No changes | empty changed set | empty []Change |
| DC1.2 | Single value change | one value change | []Change with 1 entry |
| DC1.3 | Single property change | SetProperty called | []Change with 1 entry |
| DC1.4 | Value and property change | both changed | appropriate Change entries |

### DetectChanges Priority Ordering
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| DC2.1 | High before medium | high and medium changes | high changes first |
| DC2.2 | Medium before low | medium and low changes | medium changes first |
| DC2.3 | All priorities | high, medium, low changes | sorted High -> Medium -> Low |

### DetectChanges Split Changes
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| DC3.1 | Value and property different priorities | high value, low property | 2 Change entries for same var |
| DC3.2 | Multiple properties different priorities | props at high and low | 2 Change entries with grouped props |
| DC3.3 | Properties same priority | 2 props at medium | 1 Change entry with both props |

### DetectChanges Value Priority
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| DC4.1 | High priority value | var with ValuePriority=High | value change at High priority |
| DC4.2 | Low priority value | var with ValuePriority=Low | value change at Low priority |
| DC4.3 | Default priority value | var with no priority set | value change at Medium priority |

### DetectChanges Property Priority
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| DC5.1 | Property with :high suffix | SetProperty("name:high", "v") | property change at High priority |
| DC5.2 | Property with :low suffix | SetProperty("name:low", "v") | property change at Low priority |
| DC5.3 | Property without suffix | SetProperty("name", "v") | property change at Medium priority |

### DetectChanges Slice Reuse and Auto-Clear
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| DC6.1 | Multiple DetectChanges calls | call DetectChanges twice | slice reused (same backing array) |
| DC6.2 | Auto-clears internal state | DetectChanges clears automatically | subsequent call returns empty if no new changes |

### Auto-Clear Behavior (built into DetectChanges)
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| AC1.1 | Clears value changes | value change, DetectChanges | next DetectChanges returns empty (if no new changes) |
| AC1.2 | Clears property changes | property change, DetectChanges | next DetectChanges returns empty (if no new changes) |
| AC1.3 | Clears both | value + property, DetectChanges | next DetectChanges returns empty (if no new changes) |

## Error Scenarios

| ID | Scenario | Input | Expected Error |
|----|----------|-------|----------------|
| E1 | None expected | Change is a simple value type | No errors |

## Integration Tests

| ID | Scenario | Description |
|----|----------|-------------|
| I1 | Full change cycle | Create var, change value, SetProperty, DetectChanges (returns sorted, auto-clears) |
| I2 | Mixed priority workflow | Multiple vars with different priorities, verify sort order in DetectChanges result |
| I3 | Property change only | SetProperty, verify in DetectChanges result |
