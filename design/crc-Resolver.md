# Resolver
**Source Spec:** resolver.md, api.md

## Responsibilities

### Knows
- (Interface - no state)

### Does
- Get(obj, pathElement): retrieves value at path element within obj
- Set(obj, pathElement, value): assigns value at path element within obj
- Call(obj, methodName): invokes zero-arg method, returns result
- CallWith(obj, methodName, value): invokes one-arg method (void only)

## Collaborators
- Tracker: tracker implements this interface as default resolver
- Variable: uses resolver via tracker for navigation

## Sequences
- seq-get-value.md: resolver used for path navigation (Call for getter elements)
- seq-set-value.md: resolver used for setting values (CallWith for setter elements)

## Notes

This is an interface type. The Tracker provides a default implementation using Go reflection.

### Path Element Types
| Type | Usage |
|------|-------|
| string | Struct field, map key |
| string ending in `()` | Zero-arg method (getter) - via Call |
| string ending in `(_)` | One-arg method (setter) - via CallWith |
| int | Slice/array index (0-based) |

### Path Semantics
| Pattern | Get | Set | Position |
|---------|-----|-----|----------|
| `methodName()` | OK (calls method) | OK (if not terminal) | anywhere |
| Path ending in `()` | OK | ERROR (read-only) | terminal |
| `methodName(_)` | ERROR (write-only) | OK (calls method) | terminal only |

### Default Implementation (Tracker)
The tracker's reflection-based resolver supports:
- Struct fields (exported only)
- Map keys (string keys)
- Slice/array indices
- Zero-argument method calls via Call (pathElement ends with "()")
- One-argument method calls via CallWith (pathElement ends with "(_)")

### Call Method Requirements
- Method must be exported
- Method must take zero arguments
- Method must return at least one value (first value used)

### CallWith Method Requirements
- Method must be exported
- Method must take exactly one argument
- Method must not return any values (void only)
- Argument type must be assignable from passed value

### Variable Access Property

The `access` property on Variables adds another layer of read/write control that is **independent** of path semantics:

| Access | Variable.Get() | Variable.Set() | Change Detection |
|--------|----------------|----------------|------------------|
| `rw` (default) | OK | OK | Scanned |
| `r` | OK | ERROR | Scanned |
| `w` | ERROR | OK | Skipped |

Access checks occur at the Variable level before path resolution. Both access and path restrictions must pass for an operation to succeed. See crc-Variable.md for full details and test-Resolver.md for combined scenarios.
