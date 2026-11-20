"""Exercise the installer without a real network request or binary."""

import hashlib
import os
from pathlib import Path
import subprocess
import tempfile
import unittest


class ForkInstallerTest(unittest.TestCase):
    script = Path(os.environ.get("INSTALLER", Path(__file__).resolve().parents[2] / "get-cli.sh"))
    release = "0.23.0-fallback.1"
    binary = b"#!/bin/sh\nprintf 'fixture\n'\n"

    def setUp(self):
        temp = tempfile.TemporaryDirectory()
        self.addCleanup(temp.cleanup)
        self.root = Path(temp.name)
        self.destination = self.root / "pangolin"
        self.env = dict(os.environ, PATH=f"{self.root}:{os.environ['PATH']}",
                        TMPDIR=str(self.root), FIXTURE=str(self.root), ARCH="amd64")
        self.env.pop("PANGOLIN_CLI_REPO", None)
        self.env.pop("PANGOLIN_CLI_VERSION", None)
        self.stub("uname", 'case "$1" in -s) echo Linux;; -m) echo "$ARCH";; esac')
        self.stub("curl", '''
printf '%s\n' "$2" >> "$FIXTURE/urls"
case "$2" in
    https://api.github.com/repos/hrsmwn-code/cli/releases/latest)
        printf '{"tag_name":"0.23.0-fallback.1"}\n';;
    https://github.com/hrsmwn-code/cli/releases/download/0.23.0-fallback.1/pangolin-cli_linux_amd64)
        cp "$FIXTURE/binary" "$4";;
    https://github.com/hrsmwn-code/cli/releases/download/0.23.0-fallback.1/SHA256SUMS)
        if [ "$3" = -o ]; then cp "$FIXTURE/checksums" "$4"; else cat "$FIXTURE/checksums"; fi;;
    *) exit 22;;
esac
''')
        (self.root / "binary").write_bytes(self.binary)
        self.checksum(hashlib.sha256(self.binary).hexdigest())

    def stub(self, name, body):
        path = self.root / name
        path.write_text(f"#!/bin/sh\n{body}\n")
        path.chmod(0o755)

    def checksum(self, digest):
        (self.root / "checksums").write_text(f"{digest}  pangolin-cli_linux_amd64\n")

    def install(self, *args):
        return subprocess.run(["sh", self.script, "--path", self.destination, *args],
                              env=self.env, capture_output=True, text=True)

    def test_latest_release_uses_fork_assets(self):
        result = self.install()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.destination.read_bytes(), self.binary)
        self.assertIn("Checksum verified", result.stdout)

    def test_pinned_release_rejects_bad_checksum(self):
        self.checksum("0" * 64)
        result = self.install("--version", self.release)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("Checksum mismatch", result.stderr)
        self.assertFalse(self.destination.exists())

    def test_pinned_release_needs_no_latest_release_lookup(self):
        result = self.install("--version", self.release)
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotIn("/releases/latest", (self.root / "urls").read_text())

    def test_unsupported_platform_is_rejected_before_asset_download(self):
        self.env["ARCH"] = "riscv64"
        result = self.install()
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("No fallback build for linux_riscv64", result.stderr)
        urls = self.root / "urls"
        if urls.exists():
            self.assertNotIn("/download/", urls.read_text())


if __name__ == "__main__":
    unittest.main()
