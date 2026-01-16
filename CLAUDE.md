# Project Instructions

## Design Workflow

Use the mini-spec skill for all design and implementation work.

**3-level architecture:**
- `specs/` - Human specs (WHAT & WHY)
- `design/` - Design docs (HOW - architecture)
- `src/` or root - Implementation (code)

**Commands:**
- "design this" → generates design docs only
- "implement this" → writes code, updates Artifacts checkboxes
- After code changes → unchecks Artifacts, asks about design updates

See `.claude/skills/mini-spec/SKILL.md` for the full methodology.

## Traceability Comment Format
- Use simple filenames WITHOUT directory paths
- Correct: `CRC: crc-Person.md`, `Spec: main.md`, `Sequence: seq-create-user.md`
- Wrong: `CRC: design/crc-Person.md`, `Spec: specs/main.md`

## Finding Implementations
- To find where a design element is implemented, grep for its filename (e.g., `grep "seq-get-file.md"`)

## Test Implementation
- Test designs are Level 2 artifacts in `design/test-*.md`
- ALWAYS read test designs BEFORE writing test code
- Test code MUST implement all scenarios from test designs
- Traceability: Test files reference test designs in comments: `// Test Design: test-ComponentName.md`
