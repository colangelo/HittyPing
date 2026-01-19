#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = ["click>=8.0", "rich>=13.0"]
# ///
"""
Sync completed Beads issues back to OpenSpec tasks.md.

Usage:
    ./scripts/beads-to-openspec.py
    ./scripts/beads-to-openspec.py --dry-run
    ./scripts/beads-to-openspec.py --verbose
"""

import json
import re
import subprocess
import sys
from pathlib import Path

import click
from rich.console import Console
from rich.table import Table

console = Console()

# Project root detection (script is in PROTOCOLS/scripts/, so go up 2 levels)
PROJECT_ROOT = Path(__file__).parent.parent.parent
OPENSPEC_DIR = PROJECT_ROOT / "openspec"
BEADS_DIR = PROJECT_ROOT / ".beads"
MAPPING_FILE = BEADS_DIR / "openspec-mapping.json"


def load_mapping() -> dict:
    """Load existing beads-openspec mapping."""
    if not MAPPING_FILE.exists():
        console.print(f"[red]Error:[/red] No mapping file found at {MAPPING_FILE}")
        console.print("Run openspec-to-beads.py first to create the mapping.")
        sys.exit(1)

    with open(MAPPING_FILE) as f:
        return json.load(f)


def save_mapping(mapping: dict) -> None:
    """Save updated mapping."""
    with open(MAPPING_FILE, "w") as f:
        json.dump(mapping, f, indent=2)


def get_beads_issue_status(issue_id: str) -> dict | None:
    """Get the current status of a beads issue."""
    try:
        result = subprocess.run(
            ["bd", "show", issue_id, "--json"],
            capture_output=True,
            text=True,
            cwd=PROJECT_ROOT,
        )
        if result.returncode != 0:
            return None

        return json.loads(result.stdout)
    except (json.JSONDecodeError, subprocess.SubprocessError):
        return None


def get_all_beads_issues() -> list[dict]:
    """Get all beads issues including closed."""
    try:
        result = subprocess.run(
            ["bd", "list", "--limit", "0", "--all", "--json"],
            capture_output=True,
            text=True,
            cwd=PROJECT_ROOT,
        )
        if result.returncode != 0:
            console.print(f"[red]Error listing issues:[/red] {result.stderr}")
            return []

        return json.loads(result.stdout)
    except (json.JSONDecodeError, subprocess.SubprocessError) as e:
        console.print(f"[red]Error:[/red] {e}")
        return []


def update_tasks_md(tasks_path: Path, completed_tasks: set[str], dry_run: bool = False) -> int:
    """Update tasks.md marking completed tasks. Returns count of changes."""
    if not tasks_path.exists():
        console.print(f"[red]Error:[/red] {tasks_path} not found")
        return 0

    with open(tasks_path) as f:
        lines = f.readlines()

    changes = 0
    new_lines = []

    for line in lines:
        # Match task line: - [ ] X.Y.Z Description
        task_match = re.match(r"^(-\s+\[)([ ])(]\s+)(\d+\.\d+\.\d+)(\s+.+)$", line)
        if task_match:
            task_id = task_match.group(4)
            if task_id in completed_tasks:
                # Change [ ] to [x]
                new_line = f"{task_match.group(1)}x{task_match.group(3)}{task_id}{task_match.group(5)}\n"
                if new_line != line:
                    changes += 1
                    if dry_run:
                        console.print(f"  [dim]Would mark:[/dim] {task_id} as completed")
                    else:
                        console.print(f"  [green]Marked:[/green] {task_id} as completed")
                    new_lines.append(new_line)
                    continue

        new_lines.append(line)

    if not dry_run and changes > 0:
        with open(tasks_path, "w") as f:
            f.writelines(new_lines)

    return changes


@click.command()
@click.option("--dry-run", is_flag=True, help="Show what would change without modifying files")
@click.option("--verbose", "-v", is_flag=True, help="Show detailed issue status")
def main(dry_run: bool, verbose: bool):
    """Sync completed Beads issues back to OpenSpec tasks.md.

    Reads the mapping file created by openspec-to-beads.py and updates
    tasks.md to mark completed tasks.
    """
    console.print("\n[bold]Beads → OpenSpec Sync[/bold]\n")

    # Load mapping
    mapping = load_mapping()
    change_id = mapping.get("change_id")
    tasks_mapping = mapping.get("tasks", {})

    if not change_id:
        console.print("[red]Error:[/red] No change_id in mapping file")
        sys.exit(1)

    console.print(f"Change: [cyan]{change_id}[/cyan]")
    console.print(f"Mapped tasks: [bold]{len(tasks_mapping)}[/bold]\n")

    if dry_run:
        console.print("[dim]Dry run mode - no files will be modified[/dim]\n")

    # Get all beads issues to check status
    all_issues = get_all_beads_issues()
    issues_by_id = {str(i.get("id")): i for i in all_issues}

    # Check status of each mapped task
    completed_tasks: set[str] = set()
    status_updates: dict[str, str] = {}

    # Initialize table for verbose output
    table = Table(title="Issue Status")
    table.add_column("Task ID", style="cyan")
    table.add_column("Beads #", style="green")
    table.add_column("Status")
    table.add_column("Description")

    for task_id, task_info in tasks_mapping.items():
        beads_id = task_info.get("beads_id")
        issue = issues_by_id.get(str(beads_id))

        if issue:
            status = issue.get("status", "unknown")
            is_done = status in ("done", "closed", "completed")

            if is_done:
                completed_tasks.add(task_id)

            # Track status for mapping update
            status_updates[task_id] = status

            status_style = "[green]" if is_done else "[yellow]"
            table.add_row(
                task_id,
                f"#{beads_id}",
                f"{status_style}{status}[/]",
                task_info.get("description", "")[:50],
            )
        else:
            table.add_row(
                task_id,
                f"#{beads_id}",
                "[red]not found[/]",
                task_info.get("description", "")[:50],
            )

    if verbose:
        console.print(table)
        console.print()

    # Summary
    console.print(f"Completed in beads: [green]{len(completed_tasks)}[/green] / {len(tasks_mapping)}")

    if not completed_tasks:
        console.print("\n[dim]No completed tasks to sync.[/dim]")
        return

    # Update tasks.md
    tasks_path = OPENSPEC_DIR / "changes" / change_id / "tasks.md"
    console.print(f"\nUpdating: [dim]{tasks_path.relative_to(PROJECT_ROOT)}[/dim]\n")

    changes = update_tasks_md(tasks_path, completed_tasks, dry_run)

    # Update mapping with current status
    if not dry_run:
        for task_id, status in status_updates.items():
            if task_id in mapping["tasks"]:
                mapping["tasks"][task_id]["status"] = status
        save_mapping(mapping)

    # Final summary
    if changes > 0:
        if dry_run:
            console.print(f"\n[dim]Would update {changes} tasks in tasks.md[/dim]")
        else:
            console.print(f"\n[green]✓[/green] Updated {changes} tasks in tasks.md")
            console.print(f"[green]✓[/green] Updated mapping file status")
    else:
        console.print("\n[dim]No changes needed - tasks.md already up to date[/dim]")


if __name__ == "__main__":
    main()
