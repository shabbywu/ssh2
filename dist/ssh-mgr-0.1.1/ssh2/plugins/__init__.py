# -*- coding: utf-8 -*-
from abc import ABCMeta, abstractmethod
from pathlib import Path
from typing import ClassVar, Optional

from ssh2 import conf
from ssh2.constants import AuthMethodType, PluginType
from ssh2.exceptions import ImportFromStringError
from ssh2.models.auth_method import AuthMethod
from ssh2.models.session import Session
from ssh2.utils import import_from_string
from ssh2.utils.tempfile import generate_temp_file


class BasePlugin(metaclass=ABCMeta):
    KIND: ClassVar[Optional[str]] = None

    @abstractmethod
    def to_expect_cmds(self, session: Session):
        raise NotImplementedError

    @abstractmethod
    def to_json(self):
        raise NotImplementedError

    @classmethod
    def from_dict(cls, plugin_definition: dict) -> "BasePlugin":
        kind = PluginType(plugin_definition.pop("kind"))
        plugin_args = plugin_definition.pop("args") or {}
        obj = kind.get_backend()(**plugin_args)  # type: ignore
        return obj


class BaseLoginPlugin(BasePlugin, metaclass=ABCMeta):
    @staticmethod
    def get_publishkey_path(auth: AuthMethod) -> Path:
        if auth.type == AuthMethodType.PUBLISH_KEY_CONTENT.value:
            path = generate_temp_file()
            with open(path, "w") as fh:
                fh.write(auth.content_decrypted)
        elif auth.type == AuthMethodType.PUBLISH_KEY_PATH.value:
            path = Path(auth.content_decrypted)
        else:
            raise
        return Path(path)

    def to_json(self):
        return dict(kind=self.KIND, args=dict())


# 加载默认插件
for plugin in ["ssh2.plugins.ssh:SshLogin", "ssh2.plugins.expect:ExpectPlugin"]:
    import_from_string(plugin)


# 加载额外插件
for plugin in ["ssh2_ioa:WeTermIOALogin"]:
    try:
        import_from_string(plugin)
    except ImportFromStringError as e:
        if conf.DEBUG:
            print(e)
