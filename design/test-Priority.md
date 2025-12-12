# Test Design: Priority
**Source Design:** crc-Priority.md

## Test Scenarios

### Priority Constants
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| PR1.1 | PriorityLow value | PriorityLow | -1 |
| PR1.2 | PriorityMedium value | PriorityMedium | 0 |
| PR1.3 | PriorityHigh value | PriorityHigh | 1 |
| PR1.4 | Priority ordering | Low < Medium < High | true |

### Priority String Parsing
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| PR2.1 | Parse "low" | "low" | PriorityLow |
| PR2.2 | Parse "medium" | "medium" | PriorityMedium |
| PR2.3 | Parse "high" | "high" | PriorityHigh |
| PR2.4 | Parse empty | "" | PriorityMedium (default) |
| PR2.5 | Parse invalid | "invalid" | PriorityMedium (default) |

### Priority Suffix Parsing
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| PR3.1 | Parse ":low" suffix | "name:low" | baseName="name", priority=Low |
| PR3.2 | Parse ":medium" suffix | "name:medium" | baseName="name", priority=Medium |
| PR3.3 | Parse ":high" suffix | "name:high" | baseName="name", priority=High |
| PR3.4 | No suffix | "name" | baseName="name", priority=Medium |
| PR3.5 | Multiple colons | "a:b:high" | baseName="a:b", priority=High |
| PR3.6 | Colon but no priority | "name:other" | baseName="name:other", priority=Medium |
