# -*- coding: utf-8 -*-
import json

from ssh2.models import AuthMethod, ClientConfig, ServerConfig, Session, get_scoped_session, session_scope
from ssh2.plugins.ssh import SshLogin


class TestSshBackend:
    def test_to_expect_cmds(self):
        session = get_scoped_session()
        assert session.query(Session).count() == 0
        with session_scope() as s:
            session_obj = Session(
                plugins=json.dumps([SshLogin().to_json()]),
                server=ServerConfig(host="127.0.0.1", port=22),
                client=ClientConfig(
                    user="somebody-1",
                    auth=AuthMethod.from_password("test", "password:"),
                ),
            )
            s.add(session_obj)

        assert session.query(ClientConfig).count() == 1
        assert session.query(AuthMethod).count() == 1
        assert session.query(Session).count() == 1
        assert session.query(ServerConfig).count() == 1
        session_obj: Session = session.query(Session).scalar()
        assert (
            session_obj.to_expect_cmds()
            == 'set timeout 20\n\ntrap {\n    set rows [stty rows]\n    set cols [stty columns]\n    stty rows $rows columns $cols < $spawn_out(slave,name)\n} WINCH\nspawn ssh -p 22 somebody-1@127.0.0.1\nexpect "password:"\nsend "test\r"\ninteract'
        )
