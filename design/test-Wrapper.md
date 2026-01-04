# Test Design: Wrapper

**Source Spec:** wrapper.md
**CRC Cards:** crc-Variable.md, crc-Resolver.md

## Overview

Tests for wrapper support, which allows a custom resolver to provide an alternative object for child variable navigation.

## Test Categories

### Wrapper Creation Tests

| ID | Scenario | Description | Expected |
|----|----------|-------------|----------|
| W1 | Created with property | CreateVariable with `wrapper=true` | WrapperValue and WrapperJSON set |
| W2 | Not created without property | CreateVariable without wrapper property | WrapperValue and WrapperJSON nil |
| W8 | CreateWrapper returns nil | Wrapper property set but CreateWrapper returns nil | WrapperValue nil |

### Navigation Tests

| ID | Scenario | Description | Expected |
|----|----------|-------------|----------|
| W3 | Child navigation | Child Get() navigates through WrapperValue | Returns wrapper field value |
| W4 | NavigationValue method | Variable.NavigationValue() | Returns WrapperValue if present, else Value |
| W9 | Set via child | Child Set() navigates through WrapperValue | Modifies wrapper field |

### Lifecycle Tests

| ID | Scenario | Description | Expected |
|----|----------|-------------|----------|
| W5 | Unregistered on destroy | DestroyVariable with wrapper | Wrapper unregistered from object registry |
| W6 | SetProperty triggers | SetProperty("wrapper", "true") | Creates wrapper; clearing removes it |
| W7 | Recreated on value change | DetectChanges after value modification | Old wrapper unregistered, new wrapper created |

### Wrapper Reuse Tests

| ID | Scenario | Description | Expected |
|----|----------|-------------|----------|
| W10 | Reuse preserves state | CreateWrapper returns same pointer | State preserved, WrapperJSON unchanged |
| W11 | Replacement on different pointer | CreateWrapper returns new pointer | Old unregistered, new registered, WrapperJSON recomputed |

## Test Helpers

### wrapperResolver

Custom resolver that creates `WrapperData` wrappers:

```go
type WrapperData struct {
    WrappedName string
    Extra       int
}

type wrapperResolver struct {
    *Tracker
}

func (r *wrapperResolver) CreateWrapper(v *Variable) any {
    if p, ok := v.Value.(*Person); ok {
        return &WrapperData{
            WrappedName: "Wrapped:" + p.Name,
            Extra:       42,
        }
    }
    return nil
}
```

### reusingResolver

Custom resolver that reuses existing wrapper to preserve state:

```go
type StatefulWrapper struct {
    Data      *Person
    CallCount int
}

type reusingResolver struct {
    *Tracker
}

func (r *reusingResolver) CreateWrapper(v *Variable) any {
    if w, ok := v.WrapperValue.(*StatefulWrapper); ok {
        w.Data = v.Value.(*Person)
        w.CallCount++
        return w  // Same pointer preserves state
    }
    return &StatefulWrapper{Data: v.Value.(*Person), CallCount: 1}
}
```

## Traceability

| Test ID | Implementation |
|---------|----------------|
| W1 | TestWrapper_CreatedWithProperty |
| W2 | TestWrapper_NotCreatedWithoutProperty |
| W3 | TestWrapper_ChildNavigation |
| W4 | TestWrapper_NavigationValue |
| W5 | TestWrapper_UnregisteredOnDestroy |
| W6 | TestWrapper_SetPropertyTriggers |
| W7 | TestWrapper_RecreatedOnValueChange |
| W8 | TestWrapper_CreateReturnsNil |
| W9 | TestWrapper_SetViaChild |
| W10 | TestWrapper_ReusePreservesState |
| W11 | TestWrapper_ReplacementOnDifferentPointer |
