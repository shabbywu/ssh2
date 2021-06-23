# -*- coding: utf-8 -*-
from contextlib import contextmanager
from typing import Union

from sqlalchemy import create_engine
from sqlalchemy.orm import Session as DBSession
from sqlalchemy.orm import scoped_session, sessionmaker
from ssh2.conf import DEFAULT_DB, ECHO_SQL
from ssh2.models.auth_method import AuthMethod
from ssh2.models.base import BaseModel
from ssh2.models.client_config import ClientConfig
from ssh2.models.server_config import ServerConfig
from ssh2.models.session import Session

_engine = None


def get_scoped_session() -> Union[DBSession, scoped_session]:
    return scoped_session(sessionmaker(bind=get_engine(), expire_on_commit=False))


def get_engine():
    global _engine
    if _engine is None:
        _engine = create_engine(DEFAULT_DB, echo=ECHO_SQL)
    return _engine


@contextmanager
def session_scope(session_class=None):
    """Provide a transactional scope around a series of operations."""
    session = (session_class or get_scoped_session)()
    try:
        yield session
        session.commit()
    except Exception:
        session.rollback()
        raise
    finally:
        session.close()


def create_dababases():
    BaseModel.metadata.create_all(get_engine())


__all__ = [
    'AuthMethod',
    'ClientConfig',
    'ServerConfig',
    'Session',
    'get_scoped_session',
    'session_scope',
]
