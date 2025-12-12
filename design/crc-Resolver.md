# Resolver
**Source Spec:** resolver.md, api.md

## Responsibilities

### Knows
- (Interface - no state)

### Does
- Get(obj, pathElement): retrieves value at path element within obj
- Set(obj, pathElement, value): assigns value at path element within obj

## Collaborators
- Tracker: tracker implements this interface as default resolver
- Variable: uses resolver via tracker for navigation

## Sequences
- seq-get-value.md: resolver used for path navigation
- seq-set-value.md: resolver used for setting values

## Notes

This is an interface type. The Tracker provides a default implementation using Go reflection.

### Path Element Types
| Type | Usage |
|------|-------|
| string | Struct field, map key, or method name (with "()" suffix) |
| int | Slice/array index (0-based) |

### Default Implementation (Tracker)
The tracker's reflection-based resolver supports:
- Struct fields (exported only)
- Map keys (string keys)
- Slice/array indices
- Zero-argument method calls (pathElement ends with "()")
