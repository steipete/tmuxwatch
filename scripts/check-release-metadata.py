#!/usr/bin/env python3
"""Keep source, package metadata, and the finalized changelog in sync."""

import json
import re
import sys
from pathlib import Path

version = json.loads(Path("package.json").read_text())["version"]
if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+", version):
    sys.exit("package.json must contain a stable SemVer version")
checks = {
    "cmd/tmuxwatch/main.go": rf'(?m)^var version = "{re.escape(version)}"$',
    "flake.nix": rf'version = "{re.escape(version)}";',
    "README.md": rf"tmuxwatch --version  # should print tmuxwatch {re.escape(version)}\n",
}
for name, pattern in checks.items():
    if not re.search(pattern, Path(name).read_text()):
        sys.exit(f"{name} does not match package.json version {version}")
finalized = re.search(r"(?m)^## \[([0-9]+\.[0-9]+\.[0-9]+)\] - \d{4}-\d{2}-\d{2}$",
                      Path("CHANGELOG.md").read_text())
if not finalized or finalized[1] != version:
    sys.exit(f"latest finalized changelog must match {version}")
if len(sys.argv) > 2 or (len(sys.argv) == 2 and sys.argv[1].removeprefix("v") != version):
    sys.exit(f"requested release must match {version}")
print(f"release metadata matches v{version}")
