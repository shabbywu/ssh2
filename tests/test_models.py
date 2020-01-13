# -*- coding: utf-8 -*-
import pickle
from ssh2.models import AuthMethod, ClientConfig, ServerConfig, Session, session_scope, get_scoped_session


class TestModels:
    def test_auth_method(self):
        session = get_scoped_session()
        assert session.query(AuthMethod).count() == 0
        with session_scope() as s:
            auth = AuthMethod.from_password("test")
            s.add(auth)

        assert session.query(AuthMethod).count() == 1
        auth2 = session.query(AuthMethod).scalar()
        assert auth2 == auth
        assert auth2.content != 'test'
        assert auth2.content_decrypted == 'test'

    def test_client_config(self):
        session = get_scoped_session()
        assert session.query(ClientConfig).count() == 0
        with session_scope() as s:
            client_config = ClientConfig(user="somebody-1",
                                         auth=AuthMethod.from_password("test"),
                                         sessions=[
                                             Session(plugins=b"test", server=ServerConfig(host="127.0.0.1", port=22)),
                                             Session(plugins=b"test", server=ServerConfig(host="127.0.0.1", port=23)),
                                         ])
            s.add(client_config)

        assert session.query(ClientConfig).count() == 1
        assert session.query(AuthMethod).count() == 1
        assert session.query(Session).count() == 2
        assert session.query(ServerConfig).count() == 2
        client_config2 = session.query(ClientConfig).scalar()
        assert client_config == client_config2

    def test_server_config(self):
        session = get_scoped_session()
        assert session.query(ServerConfig).count() == 0
        with session_scope() as s:
            server_config = ServerConfig(host="127.0.0.1", port=22)
            s.add(server_config)

        assert session.query(ServerConfig).count() == 1
        server_config2 = session.query(ServerConfig).scalar()
        assert server_config == server_config2

    def test_session(self):
        session = get_scoped_session()
        assert session.query(Session).count() == 0
        with session_scope() as s:
            session_obj = Session(plugins=pickle.dumps(["test"]), server=ServerConfig(host="127.0.0.1", port=22),
                                  client=ClientConfig(user="somebody-1",
                                                      auth=AuthMethod.from_password("test")))
            s.add(session_obj)

        assert session.query(ClientConfig).count() == 1
        assert session.query(AuthMethod).count() == 1
        assert session.query(Session).count() == 1
        assert session.query(ServerConfig).count() == 1
        session_obj2 = session.query(Session).scalar()
        assert session_obj == session_obj2
        assert session_obj2.plugins == pickle.dumps(["test"])
