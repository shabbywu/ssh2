# -*- coding: utf-8 -*-
from typing import TYPE_CHECKING, Type

from aenum import NamedTuple, extend_enum, skip

if TYPE_CHECKING:
    from enum import Enum
    from ssh2.plugins import BasePlugin
else:
    from aenum import Enum


AdvancePluginType = NamedTuple("AdvancePluginType", ("key", "value", "backend"))


class AuthMethodType(str, Enum):
    PASSWORD = "PASSWORD"
    PUBLISH_KEY_PATH = "PUBLISH_KEY_PATH"
    PUBLISH_KEY_CONTENT = "PUBLISH_KEY_CONTENT"
    INTERACTIVE_PASSWORD = "INTERACTIVE_PASSWORD"


class PluginType(Enum):
    SSH_LOGIN = "SSH_LOGIN"
    EXPECT = "EXPECT"

    BACKEND_MAP = skip(
        dict(
            SSH_LOGIN="ssh2.plugins.ssh.SshLogin",
            EXPECT="ssh2.plugins.expect.ExpectPlugin",
        )
    )

    @classmethod
    def register(cls, plugin: Type["BasePlugin"]):
        extend_enum(cls, plugin.KIND, plugin.KIND)
        cls.BACKEND_MAP[plugin.KIND] = plugin  # type: ignore
        return plugin

    def get_backend(self) -> Type["BasePlugin"]:
        return self.BACKEND_MAP[self.value]  # type: ignore
