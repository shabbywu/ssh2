# -*- coding: utf-8 -*-
import base64
import getpass
import os
from decimal import Decimal
from typing import Union

import six
from cryptography.fernet import Fernet

_PROTECTED_TYPES = (
    type(None),
    int,
    float,
    Decimal,
)


def is_protected_type(obj):
    """Determine if the object instance is of a protected type.

    Objects of protected types are preserved as-is when passed to
    force_text(strings_only=True).
    """
    return isinstance(obj, _PROTECTED_TYPES)


def force_text(s, encoding="utf-8", errors="strict"):
    """
    Similar to django's force_text function
    """
    # Handle the common case first for performance reasons.
    if issubclass(type(s), six.text_type):
        return s
    try:
        if not issubclass(type(s), six.string_types):
            if six.PY3:
                if isinstance(s, bytes):
                    s = six.text_type(s, encoding, errors)
                else:
                    s = six.text_type(s)
            elif hasattr(s, "__unicode__"):
                s = six.text_type(s)
            else:
                s = six.text_type(bytes(s), encoding, errors)
        else:
            # Note: We use .decode() here, instead of six.text_type(s, encoding,
            # errors), so that if s is a SafeBytes, it ends up being a
            # SafeText at the end.
            s = s.decode(encoding, errors)
    except UnicodeDecodeError:
        raise
    return s


def force_bytes(s, encoding="utf-8", strings_only=False, errors="strict"):
    """
    Similar to smart_bytes, except that lazy instances are resolved to
    strings, rather than kept as lazy objects.

    If strings_only is True, don't convert (some) non-string-like objects.
    """
    # Handle the common case first for performance reasons.
    if isinstance(s, bytes):
        if encoding == "utf-8":
            return s
        else:
            return s.decode("utf-8", errors).encode(encoding, errors)
    if strings_only and is_protected_type(s):
        return s
    if isinstance(s, memoryview):
        return bytes(s)
    else:
        return s.encode(encoding, errors)


def get_default_secret_key(generator=getpass.getuser):
    try:
        return os.environ["PY_SSH2_DEFAULT_SECRET_KEY"]
    except KeyError:
        key = generator()
        targetlen = 32
        if len(key) == 0:
            raise
        key = (key * (targetlen // len(key) + 1))[:32]
        return base64.urlsafe_b64encode(key.encode())


def b64encode(content: Union[str, bytes]):
    return force_text(base64.urlsafe_b64encode(force_bytes(content)))


def b64decode(content: Union[str, bytes]):
    return force_text(base64.urlsafe_b64decode(force_bytes(content)))


class EncryptHandler:
    def __init__(self, secret_key=get_default_secret_key()):
        self.secret_key = secret_key
        self.f = Fernet(self.secret_key)

    def encrypt(self, text: str) -> str:
        if self.Header.contain_header(text):
            return text

        text = force_bytes(text)
        return self.Header.add_header(force_text(self.f.encrypt(text)))

    def decrypt(self, encrypted: str) -> str:
        encrypted = self.Header.strip_header(encrypted)

        encrypted = force_bytes(encrypted)
        return force_text(self.f.decrypt(encrypted))

    class Header:
        HEADER = "crypt$"

        @classmethod
        def add_header(cls, text: str):
            return cls.HEADER + text

        @classmethod
        def strip_header(cls, text: str):
            # 兼容无 header 加密串
            if not cls.contain_header(text):
                return text

            return text[len(cls.HEADER) :]

        @classmethod
        def contain_header(cls, text: str) -> bool:
            return text.startswith(cls.HEADER)
