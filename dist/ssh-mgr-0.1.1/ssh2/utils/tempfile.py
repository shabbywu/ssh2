# -*- coding: utf-8 -*-
import logging
import os
import tempfile
from pathlib import Path

from ..conf import TEMP_DIR, TEMP_FILE_PREFIX, TEMP_FILE_SUFFIX

logger = logging.getLogger(__name__)


def clean_temp_file_core():
    tmp_dir = Path(TEMP_DIR)
    for file in os.listdir(tmp_dir):
        logger.info("deleting %s", file)
        if file.startswith(TEMP_FILE_PREFIX) and file.endswith(TEMP_FILE_SUFFIX):
            Path(os.path.join(TEMP_DIR, file)).unlink()


def generate_temp_file() -> Path:
    tmp_dir = Path(TEMP_DIR)
    if not tmp_dir.exists() or not tmp_dir.is_dir():
        tmp_dir.mkdir()

    path = Path(tempfile.mkstemp(prefix=TEMP_FILE_PREFIX, suffix=TEMP_FILE_SUFFIX, dir=TEMP_DIR)[1])
    logger.debug("Generating temp path: %s", path)
    return path
