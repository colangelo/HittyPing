# Ralph Loop Prompt: OpenSpec Apply

Generic prompt for implementing any active OpenSpec change proposal via Ralph Loop.

## Mission

Implement the active OpenSpec change proposal by working through `tasks.md` sequentially until ALL tasks are marked `[x]`.

**IMPORTANT**: Paths are relative to the git project root

## Iteration Protocol

Each iteration:

1. **Discover**: Run `openspec list` to identify the active change
2. **Read Context**: Use `/openspec:apply <change-id>` guidance - read `proposal.md`, `design.md`, `tasks.md`
3. **Find Next Task**: Identify the FIRST unchecked task `- [ ]` in `tasks.md`
4. **Implement**: Complete that task (or small logical group)
5. **Test**: Run relevant tests to verify implementation
6. **Update tasks.md**: Mark completed tasks as `- [x]`
7. **Commit**: Create granular git commits with conventional format
8. **Check Completion**: If ALL tasks are `[x]`, output completion promise

## Implementation Guidelines

### Before Starting

```bash
openspec list                    # Find active change
openspec show <change-id>        # Review proposal details
```

### Code Style

- Follow project conventions in `openspec/project.md`
- Match existing patterns in the codebase
- Write tests alongside implementation

### Git Commits

- Format: `feat:`, `fix:`, `test:`, `refactor:`, `docs:`
- One logical change per commit
- Commit AFTER updating `tasks.md`
- Exclude `Co-Authored-By: Claude ...`

### Testing

- Run tests after each task
- Fix failures before moving to next task

## Completion Check

After each iteration, check if ALL tasks in `tasks.md` are marked `- [x]`.

If complete:

1. Update docs: CHANGELOG, ROADMAP, CLAUDE.md, readme

2. Run final validation:

```bash
openspec validate <change-id> --strict
```

1. Output the completion promise:

```xml
<promise>OPENSPEC APPLY COMPLETE</promise>
```

## Important Rules

- Let OpenSpec guide you - read `proposal.md`, `design.md`, `tasks.md`
- Work on ONE task at a time
- Always update `tasks.md` immediately after completing work
- Always commit after updating `tasks.md`
- If blocked, note the blocker and continue to next task
- Reference spec files in `openspec/specs/` for requirements

## Start

Run `openspec list` and begin implementing the active change.
