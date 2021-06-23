# -*- coding: utf-8 -*-
import json

import yaml
from ssh2.constants import AuthMethodType
from ssh2.exceptions import AttrNotDefind, AttrNotFound
from ssh2.models import session_scope
from ssh2.models.auth_method import AuthMethod
from ssh2.models.client_config import ClientConfig
from ssh2.models.server_config import ServerConfig
from ssh2.models.session import Session

KindMapper = {
    "ClientConfig": ClientConfig,
    "AuthMethod": AuthMethod,
    "ServerConfig": ServerConfig,
    "Session": Session,
}

AttrKindMapper = {"auth": AuthMethod, "client": ClientConfig, "server": ServerConfig}


class YamlParser:
    def __init__(self, doc: str):
        self.documents = yaml.load_all(doc)

    def parse(self):
        res = []
        for document in self.documents:
            with session_scope() as session:
                res.append(self._parse(document, db=session))

        return res

    def _parse(self, document, db=None):
        kind = document.pop("kind")
        spec = document.pop("spec", {})
        ref = document.pop("ref", {})
        obj_filter_by = ref.pop("field", None)
        obj_filter_value = ref.pop("value", None)
        if kind not in KindMapper:
            raise NotImplementedError(kind)
        cls = KindMapper[kind]
        return self.__parse(
            cls,
            spec,
            filter_by=obj_filter_by,
            filter_value=obj_filter_value,
            db=db,
        )

    def __parse(self, cls, spec=None, filter_by=None, filter_value=None, db=None):
        instance = None
        for attr, kind in AttrKindMapper.items():
            if attr in spec:
                obj_define = spec.pop(attr)
                ref = obj_define.pop("ref", {})
                obj_filter_by = ref.pop("field", None)
                obj_filter_value = ref.pop("value", None)
                obj = self.__parse(
                    cls=kind,
                    spec=obj_define.pop("spec", {}),
                    filter_by=obj_filter_by,
                    filter_value=obj_filter_value,
                    db=db,
                )
                spec[attr] = obj

        if filter_by and filter_value:
            instance = self.try_get_instance(db, cls, filter_by, filter_value)

        try:
            instance, created = self._CREATE_INSTANCE_HANDLER[cls].__get__(self)(spec, instance)
        except AttrNotFound as e:
            if instance is None:
                raise e
        else:
            if created:
                db.add(instance)

        return instance

    @classmethod
    def parse_auth_method(cls, spec: dict, instance: AuthMethod = None):
        name = cls.get_attr_from_spec(spec, "name", None)
        type = AuthMethodType(cls.get_attr_from_spec(spec, "type"))
        content = cls.get_attr_from_spec(spec, "content")
        expect_for_password = cls.get_attr_from_spec(spec, "expect_for_password", None)
        save_private_key_in_db = cls.get_attr_from_spec(spec, "save_private_key_in_db", False)

        if len(spec) != 0:
            raise AttrNotDefind(spec)

        if type == AuthMethodType.PASSWORD:
            new_instance = AuthMethod.from_password(content, expect_for_password=expect_for_password, name=name)
        elif type == AuthMethodType.PUBLISH_KEY_PATH:
            new_instance = AuthMethod.from_publishkey_file(
                content, save_private_key_in_db=save_private_key_in_db, name=name
            )
        elif type == AuthMethodType.PUBLISH_KEY_CONTENT:
            new_instance = AuthMethod.from_publishkey_content(content, name=name)
        elif type == AuthMethodType.INTERACTIVE_PASSWORD:
            new_instance = AuthMethod.interactive(expect_for_password=expect_for_password, name=name)
        else:
            raise NotImplementedError

        if isinstance(instance, AuthMethod):
            for key in ["name", "type", "content", "expect_for_password"]:
                setattr(instance, key, getattr(new_instance, key))
            return instance, False
        return new_instance, True

    @classmethod
    def parse_client_config(cls, spec: dict, instance: ClientConfig = None):
        user = cls.get_attr_from_spec(spec, "user")
        name = cls.get_attr_from_spec(spec, "name")
        auth = cls.get_attr_from_spec(spec, "auth")

        if len(spec) != 0:
            raise AttrNotDefind

        assert isinstance(auth, AuthMethod)
        if isinstance(instance, ClientConfig):
            for key in ["name", "user", "auth"]:
                setattr(instance, key, locals()[key])
            return instance, False

        return ClientConfig(user=user, name=name, auth=auth), True

    @classmethod
    def parse_server_config(cls, spec: dict, instance: ServerConfig = None):
        name = cls.get_attr_from_spec(spec, "name", None)
        host = cls.get_attr_from_spec(spec, "host")
        port = int(cls.get_attr_from_spec(spec, "port"))

        if len(spec) != 0:
            raise AttrNotDefind

        if isinstance(instance, ServerConfig):
            for key in ["name", "host", "port"]:
                setattr(instance, key, locals()[key])
            return instance, False

        return ServerConfig(name=name, host=host, port=port), True

    @classmethod
    def parse_session(cls, spec: dict, instance: Session = None):
        name = cls.get_attr_from_spec(spec, "name", None)
        tag = cls.get_attr_from_spec(spec, "tag")
        client = cls.get_attr_from_spec(spec, "client")
        server = cls.get_attr_from_spec(spec, "server")
        plugins = json.dumps(cls.get_attr_from_spec(spec, "plugins"))

        if len(spec) != 0:
            raise AttrNotDefind

        if isinstance(instance, Session):
            for key in ["name", "tag", "client", "server", "plugins"]:
                setattr(instance, key, locals()[key])
            return instance, False

        return (
            Session(name=name, tag=tag, client=client, server=server, plugins=plugins),
            False,
        )

    @staticmethod
    def get_attr_from_spec(spec, attr, default=object):
        try:
            return spec.pop(attr)
        except KeyError as e:
            if default is object:
                raise AttrNotFound(attr) from e
            return default

    @staticmethod
    def try_get_instance(db, model, filter_by, filter_value):
        try:
            return db.query(model).filter_by(**{filter_by: filter_value}).scalar()
        except Exception:
            return None

    _CREATE_INSTANCE_HANDLER = {
        ClientConfig: parse_client_config,
        AuthMethod: parse_auth_method,
        ServerConfig: parse_server_config,
        Session: parse_session,
    }
