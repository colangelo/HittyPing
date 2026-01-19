# OpenSpec ↔ Beads Workflow

This document describes the integration between OpenSpec (spec-driven development) and Beads (AI-native issue tracker).

## Quick Reference

```bash
# After openspec:proposal is approved and validated
./scripts/openspec-to-beads.py <change-id>

# Before openspec:archive (after completing tasks)
./scripts/beads-to-openspec.py
```

---

## Overview

The workflow bridges two systems:

| System | Purpose | Format |
|--------|---------|--------|
| **OpenSpec** | Spec-driven planning, requirements, design docs | `tasks.md` with `- [ ] X.Y.Z Task` format |
| **Beads** | Day-to-day task tracking for AI agents | SQLite + JSONL issues |

**Key insight**: OpenSpec captures *what* needs to be built (specs, requirements). Beads tracks *progress* during implementation.

---

## Workflow Stages

### 1. Planning (OpenSpec)

```
User Request
    ↓
openspec:proposal
    ↓
proposal.md + tasks.md + spec.md files
    ↓
User Approval
    ↓
openspec validate <id> --strict
```

### 2. Implementation (Beads)

```
./scripts/openspec-to-beads.py <change-id>
    ↓
Issues created in .beads/
    ↓
Work on issues (bd update, bd close)
    ↓
Progress tracked in beads
```

### 3. Archive (OpenSpec)

```
./scripts/beads-to-openspec.py
    ↓
tasks.md updated with completion status
    ↓
openspec archive <change-id>
    ↓
Clean specs preserved in archive/
```

---

## Script Reference

### openspec-to-beads.py

Converts `tasks.md` to Beads issues.

```bash
# Convert all tasks
./scripts/openspec-to-beads.py integrate-frontend

# Preview without creating (dry run)
./scripts/openspec-to-beads.py integrate-frontend --dry-run

# Only convert pending tasks
./scripts/openspec-to-beads.py integrate-frontend --only-pending

# Recreate even if mapping exists
./scripts/openspec-to-beads.py integrate-frontend --force
```

**Output:**

- Creates issues in `.beads/issues.jsonl`
- Creates mapping file: `.beads/openspec-mapping.json`

**Mapping format:**

```json
{
  "change_id": "integrate-frontend",
  "created_at": "2026-01-17T...",
  "source_file": "openspec/changes/integrate-frontend/tasks.md",
  "tasks": {
    "1.1.1": {
      "beads_id": "42",
      "status": "open",
      "description": "Start backend...",
      "phase": "Phase 1: OpenAPI Client Generation",
      "section": "1.1 Generate TypeScript Client",
      "line_number": 8
    }
  }
}
```

### beads-to-openspec.py

Syncs completed status back to `tasks.md`.

```bash
# Sync completed tasks
./scripts/beads-to-openspec.py

# Preview changes
./scripts/beads-to-openspec.py --dry-run

# Show detailed status table
./scripts/beads-to-openspec.py --verbose
```

**What it does:**

1. Reads `.beads/openspec-mapping.json`
2. Checks status of each mapped issue via `bd list --json`
3. Updates `tasks.md` marking completed tasks: `- [ ]` → `- [x]`
4. Updates mapping file with current status

---

## Task ID Format

OpenSpec tasks use hierarchical IDs: `X.Y.Z`

- **X**: Phase number (e.g., 1, 2, 3)
- **Y**: Section within phase (e.g., 1.1, 1.2)
- **Z**: Task within section (e.g., 1.1.1, 1.1.2)

Example from `tasks.md`:

```markdown
## Phase 1: OpenAPI Client Generation
### 1.1 Generate TypeScript Client
- [ ] 1.1.1 Start backend with `docker-compose up -d`
- [ ] 1.1.2 Run `just generate-api` to fetch OpenAPI spec
```

In Beads, issues are titled: `[1.1.1] Start backend with docker-compose up -d`

---

## Best Practices

### When converting to Beads

- Only convert after `openspec validate <id> --strict` passes
- Use `--only-pending` if some tasks are already marked complete
- Review the dry-run output before creating issues

### During implementation

- Use `bd update <id> --status in_progress` when starting a task
- Use `bd close <id>` when completing a task
- Add notes with `bd comments <id> add "..."`

### Before archiving

- Run `beads-to-openspec.py --verbose` to review status
- Ensure all required tasks are marked complete
- Then run `openspec archive <id>`

---

## Troubleshooting

### "No tasks found in tasks.md"

- Check the tasks.md format matches: `- [ ] X.Y.Z Description`
- Phase headers must be: `## Phase N: Title`
- Section headers must be: `### X.Y Section Title`

### "Found existing mapping"

- Mapping already exists from previous conversion
- Use `--force` to recreate, or delete `.beads/openspec-mapping.json`

### Issue not showing as completed

- Beads status must be: `done`, `closed`, or `completed`
- Run `bd list --json` to check actual status
