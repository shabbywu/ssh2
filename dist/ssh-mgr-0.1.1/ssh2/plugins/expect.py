# -*- coding: utf-8 -*-
from typing import List

from ssh2.constants import PluginType
from ssh2.models import Session
from ssh2.plugins import BasePlugin


@PluginType.register
class ExpectPlugin(BasePlugin):
    KIND = "EXPECT"

    def __init__(self, expect=None, send=None, raw: List[str] = None):
        self.expect = expect
        self.send = send
        self.raw = raw

    def to_expect_cmds(self, session: Session):
        if self.raw:
            return self.raw

        return [f'expect "{self.expect}"', f"send {self.send}"]

    def to_json(self):
        return dict(kind=self.KIND, args=dict(expect=self.expect, send=self.send, raw=self.raw))
