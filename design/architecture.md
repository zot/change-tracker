# Architecture

## Systems

### Core Tracking System
**Purpose**: Variable management and change detection
**Design Elements**: crc-Tracker.md, crc-Variable.md, crc-Priority.md, crc-Change.md, seq-create-variable.md, seq-detect-changes.md, seq-set-property.md, test-Tracker.md, test-Variable.md, test-Priority.md, test-Change.md

### Value Resolution System
**Purpose**: Navigate and modify values via paths
**Design Elements**: crc-Resolver.md, seq-get-value.md, seq-set-value.md, test-Resolver.md

### Serialization System
**Purpose**: Convert values to Value JSON format
**Design Elements**: crc-ObjectRef.md, seq-to-value-json.md, test-ValueJSON.md

### Object Registry System
**Purpose**: Weak reference tracking for object identity
**Design Elements**: crc-ObjectRegistry.md, test-ObjectRegistry.md

## Cross-Cutting Concerns
**Purpose**: Shared infrastructure and interfaces
**Design Elements**: crc-Resolver.md (interface used across systems), crc-Priority.md (priority type used by Variable)

## Quality Assurance
**Design Elements**: gap-analysis.md

## Dependencies

```
Core Tracking System
    |
    +---> Value Resolution System (for path navigation)
    |
    +---> Serialization System (for change detection)
    |
    +---> Object Registry System (for object identity)
```
