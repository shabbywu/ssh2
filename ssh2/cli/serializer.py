# -*- coding: utf-8 -*-
import pickle
import yaml
from sqlalchemy.orm import attributes

from ssh2.exceptions import AttrNotDefind, AttrNotFound
from ssh2.constants import AuthMethodType, PluginType
from ssh2.models import session_scope
from typing import Union
from ssh2.models.auth_method import AuthMethod
from ssh2.models.client_config import ClientConfig
from ssh2.models.server_config import ServerConfig
from ssh2.models.session import Session


KindMapper = {
    ClientConfig: 'ClientConfig',
    AuthMethod: 'AuthMethod',
    ServerConfig: 'ServerConfig',
    Session: 'Session',
}

AttrKindMapper = {
    AuthMethod: 'auth',
    ClientConfig: 'client',
    ServerConfig: 'server',
}


class BaseSerializer:
    def __init__(self, resource):
        assert isinstance(resource, (AuthMethod, ClientConfig, ServerConfig, Session)), "resource type error"
        assert getattr(resource, 'id', None), "resource must be persistent"
        self.resource = resource
        self.type = type(self.resource)

    def to_kind(self):
        return KindMapper[self.type]

    def to_attr(self):
        return AttrKindMapper[self.type]

    def be_json(self):
        obj = {}
        for field_name, field in self.type.__dict__.items():
            if not isinstance(field, attributes.InstrumentedAttribute):
                continue
            obj[field_name] = getattr(self.resource, field_name)


class YamlSerializer:
    def __init__(self, resource):
        self.resource = resource
