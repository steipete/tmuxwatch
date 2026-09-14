#!/usr/bin/env python3
"""Regression coverage for the release deployment-target gate."""

import os
import subprocess
import tempfile
import unittest
from pathlib import Path

GATE = Path(__file__).resolve().with_name("check-macos-target")


class DeploymentTargetTests(unittest.TestCase):
    def check(self, output, expected, exit_code=0):
        with tempfile.TemporaryDirectory() as directory:
            tool = Path(directory, "otool")
            tool.write_text("#!/bin/sh\nprintf '%s\\n' \"$OTOOL_OUTPUT\"\nexit \"$OTOOL_EXIT\"\n")
            tool.chmod(0o755)
            env = dict(os.environ, PATH=f"{directory}:{os.environ['PATH']}",
                       OTOOL_OUTPUT=output, OTOOL_EXIT=str(exit_code))
            result = subprocess.run([str(GATE), "fixture"], env=env, capture_output=True)
            self.assertEqual(result.returncode == 0, expected, result.stderr.decode())

    def test_macos_12(self):
        self.check("cmd LC_BUILD_VERSION\nplatform 1\nminos 12.0", True)

    def test_legacy_load_command(self):
        self.check("cmd LC_VERSION_MIN_MACOSX\nversion 12.0.0", True)

    def test_newer_target(self):
        self.check("cmd LC_BUILD_VERSION\nminos 15.0", False)

    def test_missing_target(self):
        self.check("cmd LC_UUID", False)

    def test_one_bad_slice(self):
        self.check("cmd LC_BUILD_VERSION\nminos 12.0\ncmd LC_BUILD_VERSION\nminos 15.0", False)

    def test_otool_failure_after_output(self):
        self.check("cmd LC_BUILD_VERSION\nminos 12.0", False, exit_code=1)


if __name__ == "__main__":
    unittest.main()
