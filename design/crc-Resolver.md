# Resolver
**Source Spec:** resolver.md, api.md

## Responsibilities

### Knows
- (Interface - no state)

### Does
- Get(obj, pathElement): retrieves value at path element within obj
- Set(obj, pathElement, value): assigns value at path element within obj
- Call(obj, methodName): invokes zero-arg method or variadic with no args, returns result
- CallWith(obj, methodName, value): invokes one-arg method or variadic method, ignores return value
- CreateWrapper(variable *Variable): creates a wrapper object for the variable (returns nil if no wrapper needed)

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
| Path ending in `()` | OK | OK with `rw` access (variadic call) | terminal |
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
- Method must take exactly one argument or be variadic with one parameter
- Return values are ignored
- Argument type must be assignable from passed value

### CreateWrapper Method
The CreateWrapper method allows custom resolvers to create wrapper objects that stand in for a variable's value when child variables navigate paths.

**Signature:**
```go
CreateWrapper(variable *Variable) any
```

**Behavior:**
- Called when a variable has the "wrapper" property set and its ValueJSON is non-nil
- Called again when ValueJSON changes during DetectChanges
- Returns a wrapper object that children will navigate through instead of the original value
- Returns nil if no wrapper is needed (default Tracker implementation always returns nil)

**Return Value Semantics:**
- **Same pointer as v.WrapperValue**: Wrapper is preserved with its state; no unregister/re-register, no WrapperJSON recomputation
- **Different pointer (non-nil)**: Old wrapper unregistered, new wrapper registered and WrapperJSON recomputed
- **nil**: Old wrapper unregistered and cleared

This allows CreateWrapper to:
- Update the existing wrapper in place and return it to preserve persistent state (caches, cursors, etc.)
- Return a new wrapper when the wrapper type or structure needs to change
- Return nil to remove the wrapper entirely

**Use Cases:**
- Provide a different interface to children than the underlying value
- Add computed properties or methods
- Implement adapters or facades
- Maintain persistent state across value changes (e.g., caches, selection state)

### Variable Access Property

The `access` property on Variables adds another layer of read/write control that is **independent** of path semantics:

| Access | Variable.Get() | Variable.Set() | Change Detection | Initial Value Computed |
|--------|----------------|----------------|------------------|------------------------|
| `rw` (default) | OK | OK | Scanned | Yes |
| `r` | OK | ERROR | Scanned | Yes |
| `w` | ERROR | OK | Skipped | Yes |
| `action` | ERROR | OK | Skipped | No |

The key difference between `w` and `action` is that `action` skips initial value computation during CreateVariable, preventing premature invocation of action-triggering paths like `AddContact(_)`.

Access checks occur at the Variable level before path resolution. Both access and path restrictions must pass for an operation to succeed. See crc-Variable.md for full details and test-Resolver.md for combined scenarios.

### Path Restrictions by Access Mode

CreateVariable validates that the access mode is compatible with the path ending:

| Access   | Valid Path Endings        | Invalid Path Endings |
|----------|---------------------------|---------------------|
| `rw`     | fields, indices, `()`     | `(_)`               |
| `r`      | fields, indices, `()`     | `(_)`               |
| `w`      | fields, indices, `(_)`    | `()`                |
| `action` | `()`, `(_)`               | (none)              |

Key rules:
- Paths ending in `(_)` require `access: "w"` or `access: "action"` (not `r` or `rw`)
- Paths ending in `()` are allowed with `rw`, `r`, or `action` access (supports variadic method calls)
- With `rw` access and `()` path: Get() calls method with no args, Set() calls method with args

CreateVariable validation errors:
- `access: "r"` or `access: "rw"` with path ending in `(_)` -> error (cannot read from setter)
- `access: "w"` with path ending in `()` -> error (use `rw`, `r`, or `action` for zero-arg methods)
