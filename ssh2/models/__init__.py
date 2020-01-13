# -*- coding: utf-8 -*-
from typing import Union
from contextlib import contextmanager

from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, scoped_session, Session

from ssh2.conf import DEFAULT_DB, ECHO_SQL

_engine = None
BaseModel = declarative_base()


def get_scoped_session() -> Union[Session, scoped_session]:
    return scoped_session(sessionmaker(bind=get_engine(), expire_on_commit=False))


def get_engine():
    global _engine
    if _engine is None:
        _engine = create_engine(DEFAULT_DB, echo=ECHO_SQL)
    return _engine


@contextmanager
def session_scope(session_class=None):
    """Provide a transactional scope around a series of operations.
    """
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


from ssh2.models.auth_method import AuthMethod
from ssh2.models.client_config import ClientConfig
from ssh2.models.server_config import ServerConfig
from ssh2.models.session import Session


__all__ = [AuthMethod, ClientConfig, ServerConfig, Session, get_scoped_session, session_scope]
