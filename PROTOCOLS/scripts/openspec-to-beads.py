#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = ["click>=8.0", "rich>=13.0"]
# ///
"""
Convert OpenSpec tasks.md to Beads issues.

Usage:
    ./scripts/openspec-to-beads.py <change-id>
    ./scripts/openspec-to-beads.py integrate-frontend --dry-run
    ./scripts/openspec-to-beads.py integrate-frontend --only-pending
"""

import json
import re
import subprocess
import sys
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path

import click
from rich.console import Console
from rich.table import Table

console = Console()

# Project root detection
PROJECT_ROOT = Path(__file__).parent.parent
OPENSPEC_DIR = PROJECT_ROOT / "openspec"
BEADS_DIR = PROJECT_ROOT / ".beads"
MAPPING_FILE = BEADS_DIR / "openspec-mapping.json"


@dataclass
class Task:
    """Represents a single task from tasks.md."""

    id: str  # e.g., "1.2.3"
    description: str
    completed: bool
    phase: str  # Phase title
    section: str | None  # Section title (optional)
    line_number: int
    raw_line: str


@dataclass
class Phase:
    """Represents a phase grouping tasks."""

    number: int
    title: str
    sections: dict[str, list[Task]] = field(default_factory=dict)


def parse_tasks_md(tasks_path: Path) -> list[Task]:
    """Parse tasks.md and extract all tasks with their context."""
    if not tasks_path.exists():
        console.print(f"[red]Error:[/red] {tasks_path} not found")
        sys.exit(1)

    tasks: list[Task] = []
    current_phase = ""
    current_section: str | None = None

    with open(tasks_path) as f:
        lines = f.readlines()

    for line_num, line in enumerate(lines, 1):
        line = line.rstrip()

        # Phase header: ## Phase N: Title
        phase_match = re.match(r"^##\s+Phase\s+(\d+):\s+(.+)$", line)
        if phase_match:
            current_phase = f"Phase {phase_match.group(1)}: {phase_match.group(2)}"
            current_section = None
            continue

        # Section header: ### X.Y Section Title
        section_match = re.match(r"^###\s+(\d+\.\d+)\s+(.+)$", line)
        if section_match:
            current_section = f"{section_match.group(1)} {section_match.group(2)}"
            continue

        # Task line: - [ ] X.Y.Z Description or - [x] X.Y.Z Description
        task_match = re.match(r"^-\s+\[([ xX])\]\s+(\d+\.\d+\.\d+)\s+(.+)$", line)
        if task_match:
            completed = task_match.group(1).lower() == "x"
            task_id = task_match.group(2)
            description = task_match.group(3)

            tasks.append(
                Task(
                    id=task_id,
                    description=description,
                    completed=completed,
                    phase=current_phase,
                    section=current_section,
                    line_number=line_num,
                    raw_line=line,
                )
            )

    return tasks


def load_mapping() -> dict:
    """Load existing beads-openspec mapping."""
    if MAPPING_FILE.exists():
        with open(MAPPING_FILE) as f:
            return json.load(f)
    return {"tasks": {}, "created_at": None, "change_id": None}


def save_mapping(mapping: dict) -> None:
    """Save beads-openspec mapping."""
    MAPPING_FILE.parent.mkdir(parents=True, exist_ok=True)
    with open(MAPPING_FILE, "w") as f:
        json.dump(mapping, f, indent=2)


def create_beads_issue(task: Task, change_id: str, dry_run: bool = False) -> str | None:
    """Create a single beads issue for a task. Returns issue ID if created."""
    # Build issue title with task ID prefix for traceability
    title = f"[{task.id}] {task.description}"

    # Build description with context
    description_parts = [
        f"**OpenSpec Change:** `{change_id}`",
        f"**Phase:** {task.phase}",
    ]
    if task.section:
        description_parts.append(f"**Section:** {task.section}")
    description_parts.append(f"**Task ID:** `{task.id}`")
    description_parts.append("")
    description_parts.append("---")
    description_parts.append("")
    description_parts.append(task.description)

    description = "\n".join(description_parts)

    if dry_run:
        console.print(f"  [dim]Would create:[/dim] {title}")
        return None

    # Create markdown file for bd create --file
    temp_file = Path("/tmp/beads-task.md")
    with open(temp_file, "w") as f:
        f.write(f"## {title}\n\n{description}\n")

    try:
        result = subprocess.run(
            ["bd", "create", "--file", str(temp_file)],
            capture_output=True,
            text=True,
            cwd=PROJECT_ROOT,
        )
        if result.returncode != 0:
            console.print(f"  [red]Error creating issue:[/red] {result.stderr}")
            return None

        # Parse issue ID from output
        # Format: "EMS-Roster-71m: [1.1.1] Task description [P2, task]"
        match = re.search(r"^\s*([A-Za-z0-9-]+):\s+\[" + re.escape(task.id) + r"\]", result.stdout, re.MULTILINE)
        if match:
            issue_id = match.group(1)
            console.print(f"  [green]Created:[/green] {issue_id} {title}")
            return issue_id

        console.print(f"  [yellow]Warning:[/yellow] Could not parse issue ID from: {result.stdout.strip()}")
        return None

    finally:
        temp_file.unlink(missing_ok=True)


def group_tasks_by_phase(tasks: list[Task]) -> dict[str, list[Task]]:
    """Group tasks by their phase for organized display."""
    phases: dict[str, list[Task]] = {}
    for task in tasks:
        if task.phase not in phases:
            phases[task.phase] = []
        phases[task.phase].append(task)
    return phases


@click.command()
@click.argument("change_id")
@click.option("--dry-run", is_flag=True, help="Show what would be created without creating")
@click.option("--only-pending", is_flag=True, help="Only convert pending (unchecked) tasks")
@click.option("--force", is_flag=True, help="Recreate issues even if mapping exists")
def main(change_id: str, dry_run: bool, only_pending: bool, force: bool):
    """Convert OpenSpec tasks.md to Beads issues.

    CHANGE_ID is the openspec change directory name (e.g., 'integrate-frontend')
    """
    tasks_path = OPENSPEC_DIR / "changes" / change_id / "tasks.md"

    console.print(f"\n[bold]OpenSpec → Beads Conversion[/bold]")
    console.print(f"Change: [cyan]{change_id}[/cyan]")
    console.print(f"Source: [dim]{tasks_path}[/dim]\n")

    # Parse tasks
    all_tasks = parse_tasks_md(tasks_path)

    if not all_tasks:
        console.print("[yellow]No tasks found in tasks.md[/yellow]")
        return

    # Filter if only-pending
    tasks = [t for t in all_tasks if not t.completed] if only_pending else all_tasks

    # Load existing mapping
    mapping = load_mapping()

    # Check if already converted
    if mapping.get("change_id") == change_id and not force:
        existing_count = len(mapping.get("tasks", {}))
        console.print(
            f"[yellow]Warning:[/yellow] Found existing mapping with {existing_count} tasks."
        )
        console.print("Use --force to recreate issues.\n")

        # Show summary table
        table = Table(title="Existing Mapping")
        table.add_column("Task ID", style="cyan")
        table.add_column("Beads #", style="green")
        table.add_column("Status")

        for task_id, info in mapping.get("tasks", {}).items():
            table.add_row(task_id, f"#{info['beads_id']}", info.get("status", "unknown"))

        console.print(table)
        return

    # Summary before conversion
    pending_count = sum(1 for t in all_tasks if not t.completed)
    completed_count = sum(1 for t in all_tasks if t.completed)

    console.print(f"Found [bold]{len(all_tasks)}[/bold] total tasks:")
    console.print(f"  • [green]{completed_count}[/green] completed")
    console.print(f"  • [yellow]{pending_count}[/yellow] pending\n")

    if only_pending:
        console.print(f"Converting [bold]{len(tasks)}[/bold] pending tasks only.\n")
    else:
        console.print(f"Converting [bold]{len(tasks)}[/bold] tasks.\n")

    if dry_run:
        console.print("[dim]Dry run mode - no issues will be created[/dim]\n")

    # Group by phase for organized output
    phases = group_tasks_by_phase(tasks)

    # New mapping
    new_mapping = {
        "change_id": change_id,
        "created_at": datetime.now().isoformat(),
        "source_file": str(tasks_path.relative_to(PROJECT_ROOT)),
        "tasks": {},
    }

    created_count = 0
    for phase_name, phase_tasks in phases.items():
        console.print(f"[bold]{phase_name}[/bold]")

        for task in phase_tasks:
            issue_id = create_beads_issue(task, change_id, dry_run)
            if issue_id:
                new_mapping["tasks"][task.id] = {
                    "beads_id": issue_id,
                    "status": "open",
                    "description": task.description,
                    "phase": task.phase,
                    "section": task.section,
                    "line_number": task.line_number,
                }
                created_count += 1

        console.print()

    # Save mapping (unless dry run)
    if not dry_run and created_count > 0:
        save_mapping(new_mapping)
        console.print(f"[green]✓[/green] Created {created_count} issues")
        console.print(f"[green]✓[/green] Saved mapping to {MAPPING_FILE.relative_to(PROJECT_ROOT)}")
    elif dry_run:
        console.print(f"[dim]Would create {len(tasks)} issues[/dim]")


if __name__ == "__main__":
    main()
