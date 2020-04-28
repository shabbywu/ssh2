# -*- coding: utf-8 -*-
import os
import sys
from pathlib import Path

IN_TEST = "test" in sys.argv or "pytest" in sys.argv[0]

if not IN_TEST:
    DEFAULT_DB = os.environ.get("SSH2_DB_CONFIG", f"sqlite:///{ Path(os.path.expanduser('~')) / '.ssh/ssh2.db'}")
    ECHO_SQL = bool(os.environ.get("SSH2_ECHO_SQL", False))
else:
    DEFAULT_DB = "sqlite:///:memory:"
    ECHO_SQL = True


DEBUG = os.environ.get("DEBUG", False)

TEMP_FILE_PREFIX = "SSH2-TEMP-FILE-PREFIX"
TEMP_FILE_SUFFIX = ".suffix.ssh2"
TEMP_DIR = str(Path(os.path.expanduser("~")) / ".ssh/ssh2/")
