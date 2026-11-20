"""Check the CLI fork's site command and release image cannot self-update upstream."""

from pathlib import Path
import subprocess
import unittest


ROOT = Path(__file__).resolve().parents[2]


class SiteUpdatePolicyTest(unittest.TestCase):
    def test_native_site_does_not_link_server_directed_updater(self):
        imports = subprocess.check_output(
            ["go", "list", "-deps", "./cmd/up/site"], cwd=ROOT, text=True
        ).splitlines()
        self.assertFalse("github.com/fosrl/newt/updates" in imports,
                         "up site links Newt's server-directed updater")

    def test_release_image_marks_container_substrate(self):
        workflow = (ROOT / ".github/workflows/fork-image.yml").read_text()
        recipe = workflow.split("cat > Dockerfile.fork <<'EOF'\n", 1)[1].split("\n          EOF", 1)[0]
        final_stage = recipe.rsplit("FROM ", 1)[1]
        self.assertRegex(final_stage, r'(?m)^\s*ENV NEWT_SYSTEM_SUBSTRATE="?CONTAINER"?$')


if __name__ == "__main__":
    unittest.main()
