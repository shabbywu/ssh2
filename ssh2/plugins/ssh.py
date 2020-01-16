# -*- coding: utf-8 -*-
import os
import stat

from ..constants import AuthMethodType, PluginType
from ..models import Session
from ..models.auth_method import AuthMethod
from ..models.client_config import ClientConfig
from ..models.server_config import ServerConfig
from ..plugins import BaseLoginPlugin


@PluginType.register
class SshLogin(BaseLoginPlugin):
    KIND = "SSH_LOGIN"

    def to_expect_cmds(self, session: Session):
        client: ClientConfig = session.client
        auth: AuthMethod = client.auth
        server: ServerConfig = session.server

        user_host = f"{client.user}@{server.host}"
        if auth.type == AuthMethodType.PASSWORD.value:
            return [
                f"spawn ssh -p {server.port} {user_host}",
                f'expect "{auth.expect_for_password}"',
                f'send "{auth.content_decrypted}\r"',
            ]

        elif auth.type in [
            AuthMethodType.PUBLISH_KEY_PATH.value,
            AuthMethodType.PUBLISH_KEY_CONTENT.value,
        ]:
            publishkey_path = self.get_publishkey_path(auth)
            os.chmod(path=publishkey_path, mode=stat.S_IRUSR)
            return [f"spawn ssh -i {publishkey_path} -p {server.port} {user_host}"]
