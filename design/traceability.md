# Traceability

## Level 1 to Level 2 (Specs to Design)

| Spec          | Design Elements                                                                                                                                                                                                                     |
|---------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| main.md       | crc-Tracker.md, crc-Variable.md, crc-Priority.md, crc-Change.md, crc-ObjectRegistry.md, seq-create-variable.md, seq-destroy-variable.md, seq-detect-changes.md                                                                         |
| api.md        | crc-Tracker.md, crc-Variable.md, crc-Priority.md, crc-Change.md, crc-Resolver.md, crc-ObjectRef.md, seq-create-variable.md, seq-destroy-variable.md, seq-detect-changes.md, seq-get-value.md, seq-set-value.md, seq-set-property.md, seq-to-value-json.md   |
| resolver.md   | crc-Resolver.md, seq-get-value.md, seq-set-value.md                                                                                                                                                                                                   |
| value-json.md | crc-ObjectRef.md, crc-ObjectRegistry.md, seq-to-value-json.md                                                                                                                                                                                         |

## Level 2 to Level 3 (Design to Implementation)

Note: All design elements are implemented in a single file (`tracker.go`) for simplicity.

### Tracker (crc-Tracker.md)
- [x] tracker.go - Tracker struct and all methods

### Variable (crc-Variable.md)
- [x] tracker.go - Variable struct and methods

### Priority (crc-Priority.md)
- [x] tracker.go - Priority type and constants

### Change (crc-Change.md)
- [x] tracker.go - Change struct

### Resolver (crc-Resolver.md)
- [x] tracker.go - Resolver interface and default implementation

### ObjectRef (crc-ObjectRef.md)
- [x] tracker.go - ObjectRef struct and helpers (IsObjectRef, GetObjectRefID)

### ObjectRegistry (crc-ObjectRegistry.md)
- [x] tracker.go - Weak reference registry (RegisterObject, UnregisterObject, LookupObject, GetObject)

### Sequences
- [x] seq-create-variable.md -> tracker.go:CreateVariable, RegisterObject
- [x] seq-destroy-variable.md -> tracker.go:DestroyVariable
- [x] seq-detect-changes.md -> tracker.go:DetectChanges (includes internal sortChanges, checkVariable)
- [x] seq-get-value.md -> tracker.go:Variable.Get, Tracker.Get
- [x] seq-set-value.md -> tracker.go:Variable.Set, Tracker.Set
- [x] seq-set-property.md -> tracker.go:Variable.SetProperty, Tracker.recordPropertyChange
- [x] seq-to-value-json.md -> tracker.go:ToValueJSON, LookupObject

## Test Design to Test Implementation

Note: All tests are implemented in a single file (`tracker_test.go`) for simplicity.

### Test Designs
| Test Design            | Implementation          |
|------------------------|-------------------------|
| test-Tracker.md        | [x] tracker_test.go     |
| test-Variable.md       | [x] tracker_test.go     |
| test-Priority.md       | [x] tracker_test.go     |
| test-Change.md         | [x] tracker_test.go     |
| test-Resolver.md       | [x] tracker_test.go     |
| test-ObjectRegistry.md | [x] tracker_test.go     |
| test-ValueJSON.md      | [x] tracker_test.go     |

## Implementation Traceability

All public functions and methods in tracker.go have traceability comments:

| Element                      | CRC Reference                                                                                                             | Sequence Reference     |
|------------------------------|---------------------------------------------------------------------------------------------------------------------------|------------------------|
| Package                      | crc-Tracker.md, crc-Variable.md, crc-Resolver.md, crc-ObjectRef.md, crc-ObjectRegistry.md, crc-Change.md, crc-Priority.md | -                      |
| Priority type                | crc-Priority.md                                                                                                           | -                      |
| Resolver interface           | crc-Resolver.md                                                                                                           | -                      |
| ObjectRef struct             | crc-ObjectRef.md                                                                                                          | -                      |
| IsObjectRef                  | crc-ObjectRef.md                                                                                                          | -                      |
| GetObjectRefID               | crc-ObjectRef.md                                                                                                          | -                      |
| Change struct                | crc-Change.md                                                                                                             | -                      |
| Tracker struct               | crc-Tracker.md                                                                                                            | -                      |
| NewTracker                   | -                                                                                                                         | seq-create-variable.md |
| Variable struct              | crc-Variable.md                                                                                                           | -                      |
| CreateVariable               | -                                                                                                                         | seq-create-variable.md |
| GetVariable                  | crc-Tracker.md                                                                                                            | -                      |
| DestroyVariable              | crc-Tracker.md                                                                                                            | -                      |
| DetectChanges                | crc-Tracker.md                                                                                                            | seq-detect-changes.md  |
| sortChanges (internal)       | crc-Tracker.md                                                                                                            | seq-detect-changes.md  |
| recordPropertyChange         | crc-Tracker.md                                                                                                            | seq-set-property.md    |
| Variables                    | crc-Tracker.md                                                                                                            | -                      |
| RootVariables                | crc-Tracker.md                                                                                                            | -                      |
| Children                     | crc-Tracker.md                                                                                                            | -                      |
| RegisterObject               | crc-Tracker.md, crc-ObjectRegistry.md                                                                                     | seq-create-variable.md |
| UnregisterObject             | crc-Tracker.md, crc-ObjectRegistry.md                                                                                     | -                      |
| LookupObject                 | crc-Tracker.md, crc-ObjectRegistry.md                                                                                     | seq-to-value-json.md   |
| GetObject                    | crc-Tracker.md, crc-ObjectRegistry.md                                                                                     | -                      |
| ToValueJSON                  | -                                                                                                                         | seq-to-value-json.md   |
| ToValueJSONBytes             | crc-Tracker.md                                                                                                            | -                      |
| Tracker.Get                  | -                                                                                                                         | seq-get-value.md       |
| Tracker.Set                  | -                                                                                                                         | seq-set-value.md       |
| Variable.Get                 | -                                                                                                                         | seq-get-value.md       |
| Variable.Set                 | -                                                                                                                         | seq-set-value.md       |
| Variable.Parent              | crc-Variable.md                                                                                                           | -                      |
| Variable.GetProperty         | crc-Variable.md                                                                                                           | -                      |
| Variable.GetPropertyPriority | crc-Variable.md                                                                                                           | -                      |
| Variable.SetProperty         | -                                                                                                                         | seq-set-property.md    |
| Variable.SetActive           | crc-Variable.md                                                                                                           | -                      |
