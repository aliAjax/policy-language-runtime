#!/usr/bin/env python3
import json
import shlex
import subprocess
import sys
from pathlib import Path


project = Path(__file__).resolve().parents[1]
plan_path = project.parents[1] / "_shared" / "bug_plan.json"
plan = json.loads(plan_path.read_text(encoding="utf-8"))
commands = [
    (bug["record"], command)
    for bug in plan["bugs"]
    for command in bug["verify_cmds"].splitlines()
    if command.strip()
]

print(f"running {len(commands)} targeted commands", flush=True)
failed = []
for record, command in commands:
    process = subprocess.run(
        shlex.split(command),
        cwd=project,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
    )
    state = "PASS" if process.returncode == 0 else "FAIL"
    print(f"{record} {state} {command}", flush=True)
    if process.returncode:
        print(process.stdout, flush=True)
        failed.append((record, command))

print(f"summary: {len(commands) - len(failed)}/{len(commands)} passed", flush=True)
sys.exit(1 if failed else 0)
