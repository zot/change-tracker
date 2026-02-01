# VariableError
**Source Spec:** api.md
**Requirements:** R11, R59, R60, R61, R62

## Responsibilities

### Knows
- ErrorType: VariableErrorType - the category of error
- Message: string - human-readable error message
- Cause: error - underlying error if any (nil otherwise)

### Does
- Error(): returns formatted error message including cause if present

## Collaborators
- VariableErrorType: uses the error type enum

## Notes

### VariableErrorType Enum

| Value | Description |
|-------|-------------|
| NoError | No error (zero value) |
| PathError | Path navigation failed (field not found, type mismatch, etc.) |
| NotFound | Variable or parent not found |
| BadSetterCall | Setter call `(_)` in wrong position |
| BadAccess | Access mode violation |
| BadIndex | Invalid index (out of bounds, bad format) |
| BadReference | Invalid object reference |
| BadParent | Parent variable not found |
| BadCall | Method call failed |
| NilPath | Nil value encountered during path navigation |

### Error Construction

The `verror()` helper creates VariableError instances:
```go
verror(typ VariableErrorType, msg string, args ...any) *VariableError
```
- Automatically extracts `error` type from args as Cause
- Prepends error type name to message

### Usage in Variable

Variable has a `verror()` method that:
1. Creates a VariableError
2. Stores it in `v.Error`
3. Returns the error

This allows callers to check `variable.Error` after operations to get structured error information.
