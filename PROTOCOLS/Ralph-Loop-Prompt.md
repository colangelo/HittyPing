# Ralph Loop prompt

## Spec 1

Files Created

1. RALPH-IMPLEMENT-CORE-SCHEDULING.md

The main prompt document containing:

- Mission and context files to read
- Iteration protocol (check → implement → test → update → commit → check completion)
- Implementation guidelines (code style, testing, git)
- Task sections reference
- Checkpoints for each section
- Completion promise: <promise>IMPLEMENT-CORE-SCHEDULING COMPLETE</promise>

1. run-ralph-implementation.sh

Helper script showing how to run the loop.

---
How to Run

Option A: Manual in Claude Code (Recommended)

/ralph-loop:ralph-loop "Implement the implement-core-scheduling OpenSpec change. Read RALPH-IMPLEMENT-CORE-SCHEDULING.md for full instructions. Work through tasks.md sequentially, updating checkboxes as you complete each task. Output <promise>IMPLEMENT-CORE-SCHEDULING COMPLETE</promise> when ALL tasks are done." --completion-promise "IMPLEMENT-CORE-SCHEDULING COMPLETE" --max-iterations 30

Option B: With openspec:apply integration

/ralph-loop:ralph-loop "/openspec:apply implement-core-scheduling - Follow RALPH-IMPLEMENT-CORE-SCHEDULING.md for iteration protocol. Output <promise>IMPLEMENT-CORE-SCHEDULING COMPLETE</promise> when all tasks in tasks.md are marked [x]." --completion-promise "IMPLEMENT-CORE-SCHEDULING COMPLETE" --max-iterations 30

---
What Ralph Will Do Each Iteration

1. Read tasks.md → find first unchecked task
2. Implement that task
3. Run tests
4. Mark task as [x] in tasks.md
5. Git commit
6. Repeat until all 65 tasks complete
7. Output completion promise

Estimated iterations: 20-40 (depending on task grouping)

## Spec example for OpenSpec

/ralph-loop:ralph-loop "/openspec:apply - Follow RALPH-OPENSPEC-APPLY.md iteration protocol. Commit after each task. Output <promise>OPENSPEC APPLY COMPLETE</promise> when all tasks in tasks.md are [x]." --completion-promise "OPENSPEC APPLY COMPLETE" --max-iterations 50

## Spec example for Beads

/ralph-loop:ralph-loop FEATURE=add-frontend-testing VERSION=v0.9.2 "
Follow RALPH-BEADS.md protocol.
Use 'bd list' to find next task, implement it, 'bd close <id>' when done.
Output <promise>ALL TASKS COMPLETE</promise> when 'bd list' returns no open issues.
" --completion-promise "ALL TASKS COMPLETE" --max-iterations 20
