# Sequence: Set Property
**Source Spec:** api.md

## Participants
- Client: caller setting a property
- Variable: the variable instance
- Tracker: records property changes

## Sequence

```
Client              Variable            Tracker
  |                    |                    |
  |  SetProperty       |                    |
  |  (name, value)     |                    |
  |------------------->|                    |
  |                    |                    |
  |                    | parseName(name)    |
  |                    |--------.           |
  |                    |        | baseName, |
  |                    |        | priority  |
  |                    |<-------'           |
  |                    |                    |
  |                    |    [if name has ":low"]
  |                    |    priority = PriorityLow
  |                    |                    |
  |                    |    [if name has ":medium"]
  |                    |    priority = PriorityMedium
  |                    |                    |
  |                    |    [if name has ":high"]
  |                    |    priority = PriorityHigh
  |                    |                    |
  |                    |    [if no suffix]  |
  |                    |    priority = PriorityMedium
  |                    |                    |
  |                    |    [if value is empty]
  |                    | delete Properties[baseName]
  |                    | delete PropertyPriorities[baseName]
  |                    |--------.           |
  |                    |<-------'           |
  |                    |                    |
  |                    |    [if value non-empty]
  |                    | Properties[baseName] = value
  |                    | PropertyPriorities[baseName] = priority
  |                    |--------.           |
  |                    |<-------'           |
  |                    |                    |
  |                    |    [if baseName == "priority"]
  |                    | ValuePriority = parsePriority(value)
  |                    |--------.           |
  |                    |<-------'           |
  |                    |                    |
  |                    |    [if baseName == "path"]
  |                    | Path = parsePath(value)
  |                    |--------.           |
  |                    |<-------'           |
  |                    |                    |
  |                    | recordPropertyChange   |
  |                    | (v.ID, baseName)   |
  |                    |------------------->|
  |                    |                    |
  |                    |                    | propertyChanges[varID]
  |                    |                    | .append(baseName)
  |                    |                    |--------.
  |                    |                    |<-------'
  |                    |                    |
  |                    |                    | changed[varID] = true
  |                    |                    |--------.
  |                    |                    |<-------'
  |                    |                    |
  |       (void)       |                    |
  |<-------------------|                    |
  |                    |                    |
```

## Notes
- Property name can include priority suffix: `:low`, `:medium`, `:high`
- Example: `SetProperty("label:high", "Important")` sets `Properties["label"]` with priority High
- Without suffix, property defaults to PriorityMedium
- Empty value removes both the property and its priority
- Setting the "priority" property updates ValuePriority (values: "low", "medium", "high")
- Setting the "path" property re-parses the path and updates the Path field
- Property changes are recorded immediately in the tracker (not during DetectChanges)
- The variable ID is added to the changed set when a property is set
