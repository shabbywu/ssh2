# -*- coding: utf-8 -*-
from unittest import mock
import base64
from cryptography.fernet import Fernet

from ssh2.utils.crypto import get_default_secret_key, EncryptHandler


class TestUtils:
    def test_get_default_secret_key_with_envrion(self):
        expect_result = Fernet.generate_key()

        with mock.patch('ssh2.utils.crypto.os') as os:
            os.environ = dict(PY_SSH2_DEFAULT_SECRET_KEY=expect_result)
            assert get_default_secret_key() == expect_result

    def test_get_default_secret_key_with_default(self):
        expect_result = base64.urlsafe_b64encode(('test' * 8).encode())
        assert get_default_secret_key(lambda: 'test') == expect_result


class TestEncryptHandler:
    handler = EncryptHandler()

    def test_encrypt(self):
        assert self.handler.encrypt("test").startswith("crypt$")
        assert "test" not in self.handler.encrypt("test")

    def test_decrypt(self):
        assert self.handler.decrypt(self.handler.encrypt("test")) == "test"

    def test_encrypt_eq(self):
        p1 = self.handler.encrypt("test")
        p2 = self.handler.encrypt("test")
        assert p1 != p2
        assert self.handler.decrypt(p1) == self.handler.decrypt(p2)
