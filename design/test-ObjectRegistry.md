# Test Design: ObjectRegistry
**Source Design:** crc-ObjectRegistry.md

## Test Scenarios

### RegisterObject
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| OR1.1 | Register pointer | (*struct, 1) | true, object registered |
| OR1.2 | Register map | (map, 2) | true, object registered |
| OR1.3 | Register non-pointer | (int, 3) | false, not registered |
| OR1.4 | Register nil | (nil, 4) | false, not registered |
| OR1.5 | Re-register same | same ptr twice | updates varID |

### UnregisterObject
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| OR2.1 | Unregister existing | registered ptr | object removed |
| OR2.2 | Unregister non-existent | unregistered ptr | no error (no-op) |
| OR2.3 | After unregister lookup | unregistered ptr | Lookup returns false |

### LookupObject
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| OR3.1 | Lookup registered | registered ptr | (varID, true) |
| OR3.2 | Lookup unregistered | unknown ptr | (0, false) |
| OR3.3 | Lookup after GC | collected object | (0, false) |

### GetObject
| ID | Scenario | Input | Expected Output |
|----|----------|-------|-----------------|
| OR4.1 | Get registered | valid varID | object pointer |
| OR4.2 | Get unknown ID | invalid varID | nil |
| OR4.3 | Get after GC | varID of collected | nil |

## Weak Reference Behavior

| ID | Scenario | Description | Expected |
|----|----------|-------------|----------|
| WR1 | Object retention | Object still referenced | Lookup succeeds |
| WR2 | Object collection | No external refs, GC run | Lookup fails |
| WR3 | Registry cleanup | After GC | Entry removed |
| WR4 | Variable survives | Object collected | Variable still in tracker |

## Integration with CreateVariable

| ID | Scenario | Description | Expected |
|----|----------|-------------|----------|
| CV1 | Auto-register pointer | CreateVariable with *T | Registered automatically |
| CV2 | Auto-register map | CreateVariable with map | Registered automatically |
| CV3 | No register primitive | CreateVariable with int | Not in registry |
| CV4 | Child pointer value | Child resolves to *T | Registered on Get |

## Object Identity Tests

| ID | Scenario | Description | Expected |
|----|----------|-------------|----------|
| OI1 | Same object twice | Register same ptr | Same varID returned |
| OI2 | Different objects | Register two ptrs | Different varIDs |
| OI3 | Value JSON identity | Same ptr in array | Same ObjectRef |
