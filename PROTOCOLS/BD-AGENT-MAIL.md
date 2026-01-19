# Multi-Agent Development Workflow

This document describes how to use **bd** (beads issue tracker), **mcp_agent_mail** (agent coordination), and **bv** (beads viewer) for parallel development with multiple Claude Code sessions.

## Overview

| Tool | Purpose | Key Commands |
|------|---------|--------------|
| **bd** | Task tracking with dependencies | `bd ready`, `bd update`, `bd close` |
| **mcp_agent_mail** | Agent messaging & file reservations | MCP tools in Claude Code |
| **bv** | TUI dashboard & agent briefings | `bv`, `bv --agent-brief` |

## Architecture

```txt
┌─────────────────────────────────────────────────────────────────────┐
│                         Human Operator                               │
│                              ↓                                       │
│                         bv (TUI Dashboard)                           │
│                              ↓                                       │
├─────────────────────────────────────────────────────────────────────┤
│  Claude Code Session 1   │  Claude Code Session 2   │  Session N    │
│  Agent: BlueLake         │  Agent: GreenCastle      │  Agent: ...   │
│  ────────────────────    │  ────────────────────    │  ──────────   │
│  • Registered via MCP    │  • Registered via MCP    │               │
│  • Claims tasks from bd  │  • Claims tasks from bd  │               │
│  • Reserves files        │  • Reserves files        │               │
│  • Sends/receives msgs   │  • Sends/receives msgs   │               │
├─────────────────────────────────────────────────────────────────────┤
│                    Shared Infrastructure                             │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐   │
│  │ .beads/          │  │ .agent-mail/     │  │ Git Repository   │   │
│  │ - beads.db       │  │ - messages/      │  │ - Code files     │   │
│  │ - issues.jsonl   │  │ - agents/        │  │ - Commits        │   │
│  │ - config.yaml    │  │ - file_reserv/   │  │                  │   │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘
```

## Setup

### Prerequisites

1. **bd** (beads CLI) installed: `brew install steveyegge/tap/beads`
2. **bv** (beads viewer) installed: `brew install steveyegge/tap/bv`
3. **mcp_agent_mail** MCP server configured in Claude Code

### Project Key Convention

All agents must use the same canonical absolute path as `project_key` to ensure they register under the same project:

```
/Users/ac/_projects/Infra/hittyping
```

> **Important:** Avoid symlinked or alternative paths (e.g., `/Users/ac/_sync/...`). Different paths create separate projects in agent-mail, breaking coordination.

### Initialize Project

```bash
# Initialize beads in your project (one-time)
cd /path/to/project
bd init --prefix hp

# Verify setup
bd list
bv
```

### Export for bv

After creating/updating tasks, export to JSONL so bv can read them:

```bash
bd export -o .beads/issues.jsonl
```

> **Note:** bd has auto-flush with 5-second debounce, but manual export ensures immediate sync.

---

## Workflow Steps

### Step 1: Start a Claude Code Session

Each Claude Code session should register as a unique agent. Use the `macro_start_session` MCP tool:

```
macro_start_session(
  human_key="/absolute/path/to/project",
  program="claude-code",
  model="opus-4.5",
  task_description="Working on Firefox addon activate-tab feature"
)
```

This will:

- Ensure the project exists in agent-mail
- Register/update the agent identity (auto-generates name like "BlueLake")
- Fetch recent inbox messages

**Agent Naming Rules:**

- Names are auto-generated adjective+noun combinations: `BlueLake`, `GreenCastle`, `RedStone`
- Names are NOT descriptive: avoid `BackendWorker`, `UIRefactorer`
- Each session gets a unique name to track who did what

### Step 2: Check Available Tasks

From the terminal or within Claude Code:

```bash
# See all tasks ready to work on (no blockers)
bd ready

# Example output:
# 📋 Ready work (3 issues with no blockers):
# 1. [P2] [task] mozeidon-r03: Add ACTIVATE_TAB to CommandName enum
# 2. [P2] [task] mozeidon-kcb: Add cli/core/tabs-activate.go
# 3. [P2] [task] mozeidon-7t4: Extract chrome addon source
```

### Step 3: Claim a Task

Update the task status and assign yourself:

```bash
bd update mozeidon-r03 --status in_progress --assignee BlueLake
```

This prevents other agents from working on the same task.

### Step 4: Reserve Files

Before editing, reserve the files you'll touch to prevent conflicts:

```
file_reservation_paths(
  project_key="/absolute/path/to/project",
  agent_name="BlueLake",
  paths=["firefox-addon/src/models/command.ts", "firefox-addon/src/services/tabs.ts"],
  ttl_seconds=3600,
  exclusive=true,
  reason="Implementing activate-tab command"
)
```

**File Reservation Behavior:**

- Other agents will see conflicts if they try to reserve the same files
- TTL auto-expires reservations if agent crashes
- Use `renew_file_reservations` for long-running work
- Supports glob patterns: `"src/**/*.ts"`

### Step 5: Do the Work

Implement the changes. Claude Code works as normal.

### Step 6: Release Files and Close Task

When done:

```bash
# Close the task with a reason
bd close mozeidon-r03 --reason "Added ACTIVATE_TAB enum value"

# Export for bv sync
bd export -o .beads/issues.jsonl
```

Release file reservations:

```
release_file_reservations(
  project_key="/absolute/path/to/project",
  agent_name="BlueLake"
)
```

### Step 7: Notify Other Agents (Optional)

If another agent is waiting on your work:

```
send_message(
  project_key="/absolute/path/to/project",
  sender_name="BlueLake",
  to=["GreenCastle"],
  subject="ACTIVATE_TAB enum ready",
  body_md="I've added the ACTIVATE_TAB enum. You can now implement the handler case."
)
```

### Step 8: Check Inbox

Periodically check for messages from other agents:

```
fetch_inbox(
  project_key="/absolute/path/to/project",
  agent_name="BlueLake",
  include_bodies=true
)
```

---

## bd Command Reference

### Viewing Tasks

```bash
# List all tasks
bd list

# List by status
bd list --status open
bd list --status in_progress
bd list --status closed

# Show task details
bd show mozeidon-r03

# View dependency tree
bd dep tree mozeidon-lb0

# Check for cycles
bd dep cycles
```

### Creating Tasks

```bash
# Basic task
bd create "Fix the bug"

# With priority (0=highest, 4=lowest)
bd create "Critical fix" -p 0

# With type and description
bd create "Add feature" -t feature -d "Detailed description here"

# With assignee
bd create "Review code" --assignee BlueLake
```

### Managing Dependencies

```bash
# Add blocking dependency (B blocks A)
bd dep add mozeidon-a mozeidon-b --type blocks

# Add parent-child relationship
bd dep add mozeidon-parent mozeidon-child --type parent-child

# View dependencies
bd dep tree mozeidon-a
```

### Updating Tasks

```bash
# Change status
bd update mozeidon-r03 --status in_progress

# Assign to agent
bd update mozeidon-r03 --assignee BlueLake

# Change priority
bd update mozeidon-r03 --priority 1
```

### Closing Tasks

```bash
# Close single task
bd close mozeidon-r03

# Close with reason
bd close mozeidon-r03 --reason "Implemented in commit abc123"

# Close multiple
bd close mozeidon-r03 mozeidon-8e3 --reason "Both done"
```

---

## mcp_agent_mail Tool Reference

### Session Management

| Tool | Purpose |
|------|---------|
| `macro_start_session` | Bootstrap: ensure project, register agent, get inbox |
| `register_agent` | Update agent profile and refresh last_active |
| `create_agent_identity` | Create a new unique agent (never overwrites) |
| `whois` | Look up agent profile and recent activity |

### Messaging

| Tool | Purpose |
|------|---------|
| `send_message` | Send new message to recipients |
| `reply_message` | Reply within existing thread |
| `fetch_inbox` | Poll for new messages |
| `mark_message_read` | Mark as read (no ack) |
| `acknowledge_message` | Mark as read + acknowledged |
| `search_messages` | FTS5 search across messages |
| `summarize_thread` | Get key points from thread |

### File Reservations

| Tool | Purpose |
|------|---------|
| `file_reservation_paths` | Reserve files/globs with TTL |
| `release_file_reservations` | Release your reservations |
| `renew_file_reservations` | Extend TTL without re-reserving |
| `force_release_file_reservation` | Force-release stale reservation |

### Macros (Compound Operations)

| Tool | Purpose |
|------|---------|
| `macro_start_session` | Project + agent + inbox in one call |
| `macro_prepare_thread` | Join thread with context |
| `macro_file_reservation_cycle` | Reserve + optional auto-release |
| `macro_contact_handshake` | Request + approve contact |

---

## bv (Beads Viewer) Reference

### Basic Usage

```bash
# Launch TUI dashboard
bv

# Export agent briefing (for context handoff)
bv --agent-brief ./agent-brief-output/

# Generate script for ready tasks
bv --emit-script

# View at historical point
bv --as-of main~5

# Check for drift from baseline
bv --check-drift
```

### Agent Brief Export

The `--agent-brief` flag exports a bundle for handing off context:

```bash
bv --agent-brief ./brief/
# Creates:
#   brief/triage.json    - Priority-sorted issues
#   brief/insights.json  - Analysis and patterns
#   brief/brief.md       - Human-readable summary
#   brief/helpers.md     - Suggested next steps
```

---

## Example: Parallel Development Session

### Terminal Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│ Terminal 1: bv Dashboard                                            │
│ ─────────────────────────────────────────────────────────────────── │
│ [Live view of all tasks, dependencies, agent assignments]           │
├─────────────────────────────────────────────────────────────────────┤
│ Terminal 2: Claude Code (Agent: BlueLake)                           │
│ ─────────────────────────────────────────────────────────────────── │
│ Working on: mozeidon-r03 (Firefox addon enum)                       │
│ Files reserved: firefox-addon/src/models/command.ts                 │
├─────────────────────────────────────────────────────────────────────┤
│ Terminal 3: Claude Code (Agent: GreenCastle)                        │
│ ─────────────────────────────────────────────────────────────────── │
│ Working on: mozeidon-kcb (CLI core function)                        │
│ Files reserved: cli/core/tabs-activate.go                           │
└─────────────────────────────────────────────────────────────────────┘
```

### Session 1 (BlueLake) - Firefox Addon Work

```
Human: Start a session and work on the Firefox addon tasks

Claude: [Calls macro_start_session, gets agent name "BlueLake"]
        [Runs: bd ready]
        [Runs: bd update mozeidon-r03 --status in_progress --assignee BlueLake]
        [Calls file_reservation_paths for firefox-addon/src/models/command.ts]
        [Implements the change]
        [Runs: bd close mozeidon-r03 --reason "Added ACTIVATE_TAB enum"]
        [Calls release_file_reservations]
        [Calls send_message to notify GreenCastle]
```

### Session 2 (GreenCastle) - CLI Work

```
Human: Start a session and work on the CLI tasks

Claude: [Calls macro_start_session, gets agent name "GreenCastle"]
        [Runs: bd ready]
        [Runs: bd update mozeidon-kcb --status in_progress --assignee GreenCastle]
        [Calls file_reservation_paths for cli/core/tabs-activate.go]
        [Implements the change]
        [Runs: bd close mozeidon-kcb --reason "Added TabsActivate function"]
        [Calls release_file_reservations]
```

---

## Conflict Prevention

### File Reservation Conflicts

When an agent tries to reserve already-reserved files:

```
file_reservation_paths response:
{
  "granted": [],
  "conflicts": [
    {
      "path": "firefox-addon/src/services/tabs.ts",
      "holders": ["BlueLake (expires in 45 min)"]
    }
  ]
}
```

**Resolution options:**

1. Wait for the holder to finish
2. Message the holder to coordinate
3. Work on a different task
4. Force-release if holder is stale (use with caution)

### Task Conflicts

The dependency system prevents conflicts:

- Tasks with blockers don't appear in `bd ready`
- Assigned tasks show the assignee in `bd list`
- `--status in_progress` signals active work

---

## Best Practices

### For Agents

1. **Always register first** - Use `macro_start_session` at the start
2. **Claim before working** - Update status to `in_progress` immediately
3. **Reserve files** - Even for quick changes, prevent overwrites
4. **Close promptly** - Don't leave tasks in_progress when done
5. **Communicate** - Send messages for handoffs or blockers
6. **Check inbox** - Poll periodically for coordination messages

### For Operators

1. **Use bv** - Keep a terminal with the TUI open
2. **Export regularly** - Run `bd export -o .beads/issues.jsonl` after changes
3. **Monitor conflicts** - Watch for stale reservations
4. **Review messages** - Check agent-mail for coordination issues

### Task Design

1. **Small, atomic tasks** - Each task should be completable in one session
2. **Clear dependencies** - Use `blocks` type for hard dependencies
3. **Specific file scope** - Tasks should touch minimal files
4. **Include context** - Use descriptions with file paths and requirements

---

## Beads + MCP-Agent-Mail Integration

Beads is the **source of truth** for task status and assignments. MCP-Agent-Mail handles agent coordination and file conflict prevention.

### Linking Conventions

Use the Beads issue ID to connect the two systems:

| Context | Convention | Example |
|---------|------------|---------|
| Message `thread_id` | Use Beads issue ID | `thread_id="hp-r03"` |
| Message subject | Prefix with issue ID | `"[hp-r03] Ready for review"` |
| File reservation `reason` | Include issue ID | `reason="hp-r03: implementing feature"` |

### Cleanup Behavior

| Situation | What happens |
|-----------|--------------|
| Agent crashes with file reservations | Auto-expires via TTL (no action needed) |
| Agent crashes with in_progress task | Manual cleanup: `bd update <id> --status open --assignee ""` |
| Stale reservation blocking work | Use `force_release_file_reservation` or wait for TTL |

### Upstream Documentation

For comprehensive mcp_agent_mail details, see: https://github.com/Dicklesworthstone/mcp_agent_mail

---

## Troubleshooting

### bv shows "No issues found"

```bash
# Export issues to JSONL
bd export -o .beads/issues.jsonl

# Verify
wc -l .beads/issues.jsonl
```

### Agent can't reserve files

Check who holds the reservation:

```bash
# In Claude Code, check the conflict response
# Or look at .agent-mail/file_reservations/
```

### Tasks not appearing in bd ready

Check dependencies:

```bash
bd dep tree <task-id>
bd show <task-id>
```

### Message not received

```
fetch_inbox(
  project_key="...",
  agent_name="...",
  include_bodies=true,
  limit=50
)
```

---

## Quick Reference Card

```bash
# === BD COMMANDS ===
bd ready                          # Show unblocked tasks
bd update <id> --status in_progress --assignee <agent>  # Claim task
bd close <id> --reason "Done"     # Complete task
bd dep tree <id>                  # View dependencies
bd export -o .beads/issues.jsonl  # Sync for bv

# === BV COMMANDS ===
bv                                # Launch TUI
bv --agent-brief ./brief/         # Export context bundle
bv --emit-script                  # Generate task script

# === MCP TOOLS (in Claude Code) ===
macro_start_session(...)          # Start session
file_reservation_paths(...)       # Reserve files
release_file_reservations(...)    # Release files
send_message(...)                 # Send message
fetch_inbox(...)                  # Check messages
```
