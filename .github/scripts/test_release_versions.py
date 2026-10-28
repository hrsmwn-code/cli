"""Check the release workflow stamps module versions despite go.work."""

import os
from pathlib import Path
import subprocess
import unittest


class ReleaseVersionsTest(unittest.TestCase):
    def test_module_versions_in_linker_flags(self):
        root = Path(__file__).resolve().parents[2]
        workflow = (root / ".github/workflows/fork-fallback.yml").read_text()
        commands = []
        for line in workflow.splitlines():
            command = line.strip()
            if command.startswith(("NEWT_VERSION=", "OLM_VERSION=", "LDF=")):
                commands.append(command)
        result = subprocess.run(
            ["bash", "-euo", "pipefail", "-c", "\n".join(commands) + '\nprintf "%s" "$LDF"'],
            cwd=root, env=dict(os.environ, V="0.19.1-fallback.1", GITHUB_REPOSITORY="hrsmwn-code/cli"),
            check=True, capture_output=True, text=True)
        requirements = {}
        for line in (root / "go.mod").read_text().splitlines():
            fields = line.split()
            if len(fields) == 2 and fields[0] in ("github.com/fosrl/newt", "github.com/fosrl/olm"):
                requirements[fields[0]] = fields[1].removeprefix("v")
        for module, name in [("newt", "Newt"), ("olm", "Olm")]:
            self.assertIn(f"{name}VersionOverride={requirements['github.com/fosrl/' + module]}", result.stdout)
        self.assertIn("Version=0.19.1-fallback.1", result.stdout)
        self.assertIn("ForkRepo=hrsmwn-code/cli", result.stdout)


if __name__ == "__main__":
    unittest.main()
