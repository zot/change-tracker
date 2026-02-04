allow client code to choose IDS -- add CreateVariableWithId()
- client will use positive IDs (except for the root variable 1)
- server will use negative IDs (if any)



We need to add wrapper support for variables. A wrapper is an object that acts as a stand-in for the variable's value. To manage this:
- add two wrapper fields to Variable: WrapperValue and WrapperJSON
- add CreateWrapper(variable) to Resolver
- when there is a wrapper property and ValueJSON changes (including on create)
  - if WrapperValue is currently non-nil, unregister it and nil out both wrapper fields
  - if ValueJSON changed is non-nil
    - create the wrapper with Resolve.CreateWrapper(variable)
    - register the wrapper
    - store it in WrapperValue
    - store the WrapperJSON for the registered value

## Variable Wrappers: Transforming Values

The `wrapper` property on variables enables value transformation at the backend. When a variable has a wrapper, the backend uses `Wrapper(variable)` to compute the outgoing value instead of sending the raw value directly.

### The Wrapper Property

Any variable can have a `wrapper` property set via path syntax:

```html
<!-- Direct wrapper usage in viewdef -->
<div data-ui-view="selectedContact?wrapper=ContactPresenter">

<!-- Wrapper with additional properties -->
<div data-ui-path="currentUser?wrapper=UserPresenter&editable=true">
```

The wrapper:
- Receives the **variable** (not just the value), enabling it to watch for changes
- Computes the outgoing JSON value sent to the frontend
- Can create/manage additional objects (like presenters)

### Variable Value Architecture

Variables need two distinct values:

1. **Monitored value** - Used to detect changes
   - For arrays: a copy so content changes are detected
   - For other values: tracks the raw value from the path

2. **Outgoing JSON value** - What gets sent to frontend
   - Without wrapper: monitored value in "value JSON" form (objects as `{obj: ID}` refs, not inline)
   - With wrapper: computed by `Wrapper(variable)`
   - Enables transformation (e.g., domain object ref → presenter object ref)

### Wrapper Use Cases

```html
<!-- Wrap a single object in a presenter -->
<div data-ui-view="contact?wrapper=ContactPresenter">

<!-- Wrap with editable form presenter -->
<sl-input data-ui-path="user.email?wrapper=EditableField&validate=email">

<!-- Custom computed value -->
<span data-ui-text="items?wrapper=CountDisplay">  <!-- shows "3 items" -->
```

