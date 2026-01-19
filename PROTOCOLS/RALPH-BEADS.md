# Ralph Loop Prompt: Beads Task Tracking

Generic prompt for implementing tasks tracked in Beads via Ralph Loop.

## Mission

Work through Beads issues sequentially until ALL tasks are closed.

**IMPORTANT**: Paths are relative to the git project root

## Prerequisites

Before starting, ensure:

1. OpenSpec proposal exists and is approved
2. Tasks have been converted: `./PROTOCOLS/scripts/openspec-to-beads.py <change-id>`
3. Mapping file exists: `.beads/openspec-mapping.json`

## Iteration Protocol

Each iteration:

1. **List Tasks**: Run `bd list` to see open issues
2. **Pick Next Task**: Select the first open task (or by priority/phase)
3. **Understand**: Read the issue details with `bd show <id>`
4. **Implement**: Complete the task
5. **Test**: Run relevant tests to verify implementation
6. **Close Issue**: Mark complete with `bd close <id>`
7. **Commit**: Create granular git commits with conventional format
8. **Check Completion**: If `bd list` shows no open issues, proceed to finalization

## Commands Reference

```bash
# List open tasks
bd list

# List all tasks including closed
bd list --all

# Show issue details
bd show <id>

# Mark task as in progress (optional)
bd update <id> --status in_progress

# Mark task complete
bd close <id>

# Add notes/comments
bd comments <id> add "note about implementation"

# Search tasks
bd search "keyword"
```

## Implementation Guidelines

### Task Selection

- Work in phase order (1.x before 2.x)
- Within a phase, work sequentially (1.1.1, 1.1.2, ...)
- If blocked, add a comment and move to next task

### Code Style

- Follow project conventions in `openspec/project.md`
- Match existing patterns in the codebase
- Write tests alongside implementation

### Git Commits

- Format: `feat:`, `fix:`, `test:`, `refactor:`, `docs:`
- One logical change per commit
- Commit AFTER closing the beads issue

### Testing

- Run tests after each task
- Fix failures before moving to next task

## Context Management

To avoid context rot during long implementation sessions:

- **Use subagents for implementation**: Delegate each task to a `Task` subagent with `subagent_type="general-purpose"` or specialized agents like `feature-dev:code-architect`
- **Keep main context clean**: The main chat should only track progress (`bd list`, `bd close`) and orchestrate subagents
- **Subagent scope**: Each subagent handles one task: read spec, implement, test, return result
- **Main chat commits**: After subagent completes, main chat does `bd close` and `git commit`

Example iteration:

```
1. Run `bd list` to find next task
2. Spawn subagent: "Implement hittyping-ji0: create main_test.go with imports"
3. Subagent returns completion status
4. Run `bd close hittyping-ji0`
5. Commit changes
6. Repeat
```

## Completion Check

After each iteration, run `bd list` to check remaining tasks.

If no open issues remain:

1. **Sync back to OpenSpec** (REQUIRED before archive):

```bash
./PROTOCOLS/scripts/beads-to-openspec.py
```

> **WARNING**: The OpenSpec archiver does not know about Beads. You MUST run this sync before `openspec archive` or tasks.md will have unchecked items.

2. **Verify tasks.md** is updated with all `[x]` marks

2. **Update docs**: CHANGELOG, ROADMAP, CLAUDE.md

3. **Run final validation**:

```bash
openspec validate <change-id> --strict
```

1. **Output the completion promise**:

```
<promise>ALL TASKS COMPLETE</promise>
```

## Important Rules

- Use `bd list` as source of truth for remaining work
- Close issues immediately after completing work
- Always commit after closing an issue
- If blocked, add comment: `bd comments <id> add "BLOCKED: reason"`
- Reference spec files in `openspec/specs/` for requirements
- Sync back to OpenSpec before archiving

## Example Prompt

```bash
/ralph-loop:ralph-loop FEATURE=integrate-frontend VERSION=v0.5.0 "
Follow RALPH-BEADS.md protocol.
Use 'bd list' to find next task, implement it, 'bd close <id>' when done.
Output <promise>ALL TASKS COMPLETE</promise> when 'bd list' returns no open issues.
" --completion-promise "ALL TASKS COMPLETE" --max-iterations 30
```

## Start

Run `bd list` and begin implementing tasks.
