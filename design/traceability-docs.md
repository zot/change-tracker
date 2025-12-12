# Documentation Traceability

Maps source files to documentation sections.

## Specs to Documentation

### main.md
| Section | Documentation |
|---------|---------------|
| Overview | docs/requirements.md (Overview), docs/user-manual.md (Introduction) |
| Design Principles | docs/requirements.md (NFR1, NFR2) |
| Core Concepts - Tracker | docs/requirements.md (FR1, BR1), docs/design.md (Tracker), docs/user-manual.md (Creating Variables) |
| Core Concepts - Variables | docs/requirements.md (FR2, BR1), docs/design.md (Variable), docs/user-manual.md (Creating Variables) |
| Core Concepts - Priorities | docs/requirements.md (BR3, FR10), docs/design.md (Priority), docs/user-manual.md (Priority Levels) |
| Core Concepts - Object Registry | docs/requirements.md (FR12, BR4), docs/design.md (ObjectRegistry), docs/user-manual.md (Object Registry) |
| Core Concepts - Value JSON | docs/requirements.md (FR13), docs/design.md (ObjectRef), docs/user-manual.md (Object Registry) |
| Core Concepts - Change Detection | docs/requirements.md (FR7, BR2), docs/design.md (Change Detection Flow), docs/user-manual.md (Change Detection) |
| Core Concepts - Sorted Changes | docs/requirements.md (FR10, BR3), docs/design.md (Sort Changes Flow), docs/user-manual.md (Sorted Changes) |
| Use Cases | docs/user-manual.md (How-To Guides) |

### api.md
| Section | Documentation |
|---------|---------------|
| Priority type | docs/requirements.md (FR11), docs/design.md (Priority) |
| Tracker type | docs/requirements.md (FR1-FR10), docs/design.md (Tracker) |
| Variable type | docs/requirements.md (FR2, FR5, FR6), docs/design.md (Variable) |
| Resolver interface | docs/requirements.md (FR14), docs/design.md (Resolver) |
| ObjectRef type | docs/requirements.md (FR13), docs/design.md (ObjectRef) |
| Change type | docs/requirements.md (FR10), docs/design.md (Change) |
| NewTracker | docs/requirements.md (FR1), docs/developer-guide.md (Basic Usage) |
| CreateVariable | docs/requirements.md (FR2), docs/user-manual.md (Creating Variables) |
| GetVariable | docs/requirements.md (FR3), docs/developer-guide.md (Code Examples) |
| DestroyVariable | docs/requirements.md (FR4), docs/developer-guide.md (Code Examples) |
| DetectChanges | docs/requirements.md (FR7, FR8, FR9, FR10), docs/user-manual.md (Change Detection, Sorted Changes) |
| Variable.Get | docs/requirements.md (FR5), docs/user-manual.md (Getting and Setting Values) |
| Variable.Set | docs/requirements.md (FR6), docs/user-manual.md (Getting and Setting Values) |
| Variable.SetProperty | docs/requirements.md (FR11), docs/user-manual.md (Properties) |
| Object Registry Methods | docs/requirements.md (FR12), docs/user-manual.md (Object Registry) |
| ToValueJSON | docs/requirements.md (FR13), docs/design.md (Value JSON Serialization Flow) |

### resolver.md
| Section | Documentation |
|---------|---------------|
| Interface | docs/requirements.md (FR14), docs/design.md (Resolver) |
| Default Resolver | docs/design.md (Tracker as Default Resolver), docs/developer-guide.md (Custom Resolver) |
| Path Elements | docs/user-manual.md (Path Navigation) |
| Get Operations | docs/design.md (Value Get Flow), docs/user-manual.md (Getting and Setting Values) |
| Set Operations | docs/design.md (Value Set Flow), docs/user-manual.md (Getting and Setting Values) |
| Error Conditions | docs/requirements.md (TC3, TC4), docs/user-manual.md (Troubleshooting) |
| Custom Resolvers | docs/developer-guide.md (Custom Resolver), docs/user-manual.md (Use a Custom Resolver) |

### value-json.md
| Section | Documentation |
|---------|---------------|
| Purpose | docs/requirements.md (FR13), docs/design.md (Serialization System) |
| Format | docs/design.md (ObjectRef), docs/design.md (Value JSON Serialization Flow) |
| Registration Rules | docs/requirements.md (TC2), docs/user-manual.md (Object Registry) |
| Serialization Algorithm | docs/design.md (Value JSON Serialization Flow) |
| Examples | docs/user-manual.md (Object Registry) |

## Design to Documentation

### CRC Cards to docs/design.md
| CRC Card | Design Section |
|----------|----------------|
| crc-Tracker.md | System Components - Tracker |
| crc-Variable.md | System Components - Variable |
| crc-Priority.md | System Components - Priority |
| crc-Change.md | System Components - Change |
| crc-Resolver.md | System Components - Resolver |
| crc-ObjectRef.md | System Components - ObjectRef |
| crc-ObjectRegistry.md | System Components - ObjectRegistry |

### CRC Cards to docs/developer-guide.md
| CRC Card | Developer Guide Section |
|----------|-------------------------|
| crc-Tracker.md | Architecture - Key Components - Tracker |
| crc-Variable.md | Architecture - Key Components - Variable |
| crc-Resolver.md | Architecture - Key Components - Resolver |

### Sequences to docs/design.md
| Sequence | Design Section |
|----------|----------------|
| seq-create-variable.md | Data Flow - Variable Creation Flow |
| seq-detect-changes.md | Data Flow - Change Detection Flow |
| seq-get-value.md | Data Flow - Value Get Flow |
| seq-set-value.md | Data Flow - Value Set Flow |
| seq-set-property.md | Data Flow - Property Set Flow |
| seq-to-value-json.md | Data Flow - Value JSON Serialization Flow |

### Sequences to docs/user-manual.md
| Sequence | User Manual Section |
|----------|---------------------|
| seq-create-variable.md | Features - Creating Variables |
| seq-detect-changes.md | Features - Change Detection |
| seq-get-value.md | Features - Getting and Setting Values |
| seq-set-value.md | Features - Getting and Setting Values |
| seq-set-property.md | Features - Properties |
| seq-detect-changes.md | Features - Sorted Changes |

### Test Designs to docs/developer-guide.md
| Test Design | Developer Guide Section |
|-------------|-------------------------|
| test-Tracker.md | Testing - Test Design Methodology |
| test-Variable.md | Testing - Test File Structure |
| test-Resolver.md | Testing - Test Categories |
| test-ObjectRegistry.md | Testing - Test Categories |
| test-ValueJSON.md | Testing - Test Categories |
| test-Priority.md | Testing - Test Categories |
| test-Change.md | Testing - Test Categories |

## Coverage Summary

### Requirements Documentation (docs/requirements.md)
- **Specs Covered**: main.md (100%), api.md (100%), resolver.md (100%), value-json.md (100%)
- **Functional Requirements**: 15 documented
- **Non-Functional Requirements**: 5 documented
- **Technical Constraints**: 4 documented

### Design Documentation (docs/design.md)
- **CRC Cards Covered**: 7/7 (100%)
- **Sequences Covered**: 6/6 (100%)
- **Architecture Sections**: 4 systems documented
- **Design Patterns**: 4 patterns documented
- **Key Decisions**: 6 decisions documented

### Developer Guide (docs/developer-guide.md)
- **CRC Cards Referenced**: crc-Tracker.md, crc-Variable.md, crc-Resolver.md
- **Setup Instructions**: Prerequisites, Installation, Running Tests
- **Architecture Overview**: 4 systems, key components, patterns
- **Workflow Documentation**: Development workflow, Adding features, Testing

### User Manual (docs/user-manual.md)
- **Features Documented**: 9 features
- **How-To Guides**: 6 guides
- **Troubleshooting**: 6 common issues

## Gaps

### No Gaps Identified
All specs and design elements are mapped to documentation:
- All CRC cards appear in design documentation
- All sequences appear in data flow documentation
- All spec sections appear in requirements documentation
- Test designs referenced in developer guide
- User manual covers all major features

### Documentation Maintenance Notes
1. When adding new CRC cards, update docs/design.md System Components section
2. When adding new sequences, update docs/design.md Data Flow section
3. When adding new features, update all four documentation files
4. When modifying specs, review requirements.md for consistency
