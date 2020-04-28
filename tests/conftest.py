# -*- coding: utf-8 -*-
from unittest import mock

import pytest
import sqlalchemy as sa
from sqlalchemy.orm import scoped_session, sessionmaker
from ssh2.models import create_dababases, get_engine


def pytest_addoption(parser):
    group = parser.getgroup("ssh2")
    group._addoption(
        "--reuse-db",
        action="store_true",
        dest="reuse_db",
        default=False,
        help="Re-use the testing database if it already exists, " "and do not remove it when the test finishes.",
    )


@pytest.fixture(autouse=True, scope="session")
def auto_create_db_and_drop(request):
    create_dababases()
    yield
    if not request.config.getvalue("reuse_db"):
        # TODO: drop db
        pass


@pytest.fixture(autouse=True, scope="function")
def sqlalchemy_transaction(request):
    """为使用了 sqlalchemy 操作 legacy db 的单元测试提供自动回滚，保证单元测试前后的状态一致"""
    session = None

    def fake_sessionmaker(*args, **kwargs):
        # copy from [pytest_flask_sqlalchemy](https://github.com/jeancochrane/pytest-flask-sqlalchemy/blob/master/pytest_flask_sqlalchemy/fixtures.py)
        # but remove flask requirement
        nonlocal session
        if session is not None:
            # 由于要控制 session 的rollback, 需确保在scope=`function`的范围内, 只有一个 session 实例
            return session

        engine = get_engine()
        connection = engine.connect()
        transaction = connection.begin()
        session = scoped_session(sessionmaker(bind=connection, expire_on_commit=False))

        # Make sure the session, connection, and transaction can't be closed by accident in
        # the codebase
        connection.force_close = connection.close
        transaction.force_rollback = transaction.rollback

        connection.close = lambda: None
        transaction.rollback = lambda: None
        session.close = lambda: None

        # Begin a nested transaction (any new transactions created in the codebase
        # will be held until this outer transaction is committed or closed)
        session.begin_nested()

        # Each time the SAVEPOINT for the nested transaction ends, reopen it
        @sa.event.listens_for(session, "after_transaction_end")
        def restart_savepoint(session, trans):
            if trans.nested and not trans._parent.nested:
                # ensure that state is expired the way
                # session.commit() at the top level normally does
                session.expire_all()

                session.begin_nested()

        # Force the connection to use nested transactions
        connection.begin = connection.begin_nested

        # If an object gets moved to the 'detached' state by a call to flush the session,
        # add it back into the session (this allows us to see changes made to objects
        # in the context of a test, even when the change was made elsewhere in
        # the codebase)
        @sa.event.listens_for(session, "persistent_to_detached")
        @sa.event.listens_for(session, "deleted_to_detached")
        def rehydrate_object(session, obj):
            session.add(obj)

        @request.addfinalizer
        def teardown_transaction():
            # Delete the session
            session.remove()

            # Rollback the transaction and return the connection to the pool
            transaction.force_rollback()
            connection.force_close()

        return session

    with mock.patch("ssh2.models.sessionmaker") as _sessionmaker:
        _sessionmaker.side_effect = fake_sessionmaker
        yield
