# -*- coding: utf-8 -*-
import os
import stat
from ssh2.models import Session
from ssh2.plugins import BaseLoginPlugin

from ssh2.models.server_config import ServerConfig
from ssh2.models.client_config import ClientConfig
from ssh2.models.auth_method import AuthMethod
from ssh2.constants import AuthMethodType, PluginType


@PluginType.register
class SshLogin(BaseLoginPlugin):
    KIND = 'SSH_LOGIN'

    def to_expect_cmds(self, session: Session):
        client: ClientConfig = session.client
        auth: AuthMethod = client.auth
        server: ServerConfig = session.server

        user_host = f"{client.user}@{server.host}"
        if auth.type == AuthMethodType.PASSWORD.value:
            return [f"spawn ssh -p {server.port} {user_host}",
                    f"expect \"{auth.expect_for_password}\"",
                    f"send \"{auth.content_decrypted}\r\""]

        elif auth.type in [AuthMethodType.PUBLISH_KEY_PATH.value, AuthMethodType.PUBLISH_KEY_CONTENT.value]:
            publishkey_path = self.get_publishkey_path(auth)
            os.chmod(path=publishkey_path, mode=stat.S_IRUSR)
            return [f"spawn ssh -i {publishkey_path} -p {server.port} {user_host}"]
