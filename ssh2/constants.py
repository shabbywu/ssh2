# -*- coding: utf-8 -*-
import typing

from aenum import Enum, NamedTuple, extend_enum, skip

if typing.TYPE_CHECKING:
    from .plugins import BasePlugin


AdvancePluginType = NamedTuple("AdvancePluginType", ("key", "value", "backend"))


class AuthMethodType(Enum):
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
    def register(cls, plugin: "BasePlugin"):
        extend_enum(cls, plugin.KIND, plugin.KIND)
        cls.BACKEND_MAP[plugin.KIND] = plugin
        return plugin

    def get_backend(self):
        return self.BACKEND_MAP[self.value]
