# -*- coding: utf-8 -*-
import yaml
from textwrap import dedent

from ssh2.utils.tempfile import generate_temp_file
from ssh2.cli.parser import YamlParser
from ssh2.models import get_scoped_session
from ssh2.models.auth_method import AuthMethodType, AuthMethod
from ssh2.models.client_config import ClientConfig
from ssh2.models.server_config import ServerConfig
from ssh2.models.session import Session
from ssh2.utils.crypto import b64encode


class TestYamlParser:
    def test_parse_auth_method_password(self):
        session = get_scoped_session()
        assert session.query(AuthMethod).count() == 0
        spec = dict(name="test", type="PASSWORD", content="password")
        instance, _ = YamlParser.parse_auth_method(spec)
        session.add(instance)
        session.commit()

        assert session.query(AuthMethod).count() == 1

    def test_parse_auth_method_publishkey_path(self):
        session = get_scoped_session()
        assert session.query(AuthMethod).count() == 0

        temppath = generate_temp_file()
        with open(temppath, "w") as fh:
            fh.write("test\n")

        spec = dict(name="test", type="PUBLISH_KEY_PATH", content=temppath.absolute(), save_private_key_in_db=True)
        instance, _ = YamlParser.parse_auth_method(spec)
        session.add(instance)
        session.commit()

        assert session.query(AuthMethod).count() == 1
        assert instance.content_decrypted == "test\n"

    def test_parse_auth_method_publishkey_content(self):
        session = get_scoped_session()
        assert session.query(AuthMethod).count() == 0
        spec = dict(name="test", type="PUBLISH_KEY_CONTENT", content=b64encode("test"))
        instance, _ = YamlParser.parse_auth_method(spec)
        session.add(instance)
        session.commit()

        assert session.query(AuthMethod).count() == 1
        assert instance.content_decrypted == 'test'

    def test_parse_client_config(self):
        session = get_scoped_session()
        assert session.query(ClientConfig).count() == 0
        assert session.query(AuthMethod).count() == 0
        auth_spec = dict(name="test", type="PASSWORD", content="password")
        auth, _ = YamlParser.parse_auth_method(auth_spec)

        client_spec = dict(user="user", name="name", auth=auth)
        client, _ = YamlParser.parse_client_config(client_spec)
        session.add(client)
        session.commit()

        assert session.query(ClientConfig).count() == 1
        assert session.query(AuthMethod).count() == 1

    def test_parse_server_config(self):
        session = get_scoped_session()
        assert session.query(ServerConfig).count() == 0
        server_spec = dict(name="test", host="127.0.0.1", port="123")
        server, _ = YamlParser.parse_server_config(server_spec)
        session.add(server)
        session.commit()

        assert session.query(ServerConfig).count() == 1

    def test_parse_session(self):
        s = get_scoped_session()
        assert s.query(ServerConfig).count() == 0
        assert s.query(ClientConfig).count() == 0
        assert s.query(AuthMethod).count() == 0
        assert s.query(Session).count() == 0
        auth_spec = dict(name="test", type="PASSWORD", content="password")
        auth, _ = YamlParser.parse_auth_method(auth_spec)

        client_spec = dict(user="user", name="name", auth=auth)
        client, _ = YamlParser.parse_client_config(client_spec)

        server_spec = dict(name="test", host="127.0.0.1", port="123")
        server, _ = YamlParser.parse_server_config(server_spec)

        session_spec = dict(name="test", tag="test", client=client, server=server, plugins=[
            dict(kind="SSH_LOGIN", args=dict()),
            dict(kind="EXPECT", args=dict(raw=['test']))
        ])
        session, _ = YamlParser.parse_session(session_spec)
        s.add(session)
        s.commit()

        assert s.query(ServerConfig).count() == 1
        assert s.query(ClientConfig).count() == 1
        assert s.query(AuthMethod).count() == 1
        assert s.query(Session).count() == 1

    def test_parse_simple(self):
        session = get_scoped_session()
        assert session.query(AuthMethod).count() == 0

        document = dedent("""
        kind: AuthMethod
        spec:
            name: str
            type: PASSWORD
            content: password
            expect_for_password: test
        """)
        parser = YamlParser(document)
        auth = parser.parse()[0]

        assert session.query(AuthMethod).count() == 1
        assert auth.name == 'str'
        assert auth.type == AuthMethodType.PASSWORD.value
        assert auth.content_decrypted == 'password'
        assert auth.expect_for_password == 'test'

    def test_parse_complex(self):
        s = get_scoped_session()
        assert s.query(ServerConfig).count() == 0
        assert s.query(ClientConfig).count() == 0
        assert s.query(AuthMethod).count() == 0
        assert s.query(Session).count() == 0

        temppath = generate_temp_file()
        with open(temppath, "w") as fh:
            fh.write("test\n")

        document = dedent(f"""
        kind: Session
        spec:
            tag: str
            name: str
            plugins:
                -   kind:   SSH_LOGIN
                    args:   
                -   kind:   EXPECT
                    args:
                        expect: str
                        send:   str
                        raw:
                        -   str
                        -   str
                        -   str
            client:
                spec:
                    user: str
                    name: str | nullable
                    auth:
                        spec:
                            name: str | nullable
                            type: PUBLISH_KEY_PATH
                            content: {temppath.absolute()}
                            save_private_key_in_db: true
            server:
                spec:
                    name: str
                    host: 127.0.0.1
                    port: 103
        """)
        session = YamlParser(document).parse()[0]

        assert s.query(ServerConfig).count() == 1
        assert s.query(ClientConfig).count() == 1
        assert s.query(AuthMethod).count() == 1
        assert s.query(Session).count() == 1

    def test_update_by_parser(self):
        s = get_scoped_session()
        assert s.query(AuthMethod).count() == 0

        temppath = generate_temp_file()
        with open(temppath, "w") as fh:
            fh.write("test\n")

        document = dedent(f"""
        kind: Session
        spec:
            tag: str
            name: str
            plugins:
                -   kind:   SSH_LOGIN
                    args:   
                -   kind:   EXPECT
                    args:
                        expect: str
                        send:   str
                        raw:
                        -   str
                        -   str
                        -   str
            client:
                spec:
                    user: str
                    name: str | nullable
                    auth:
                        spec:
                            name: str | nullable
                            type: PUBLISH_KEY_PATH
                            content: {temppath.absolute()}
                            save_private_key_in_db: true
            server:
                spec:
                    name: str
                    host: 127.0.0.1
                    port: 103
        """)

        parser = YamlParser(document)
        parser.parse()
        session = s.query(Session).scalar()

        assert s.query(ServerConfig).count() == 1
        assert s.query(ClientConfig).count() == 1
        assert s.query(AuthMethod).count() == 1
        assert s.query(Session).count() == 1
        assert session.name == 'str'
        assert session.tag == 'str'

        json = session.to_json()
        json['spec']['name'] = "名字"
        json['spec']['tag'] = "标签"

        YamlParser(yaml.dump(json, allow_unicode=True)).parse()
        session = s.query(Session).scalar()

        assert s.query(Session).count() == 1
        assert session.name == '名字'
        assert session.tag == '标签'
