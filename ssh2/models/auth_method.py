# -*- coding: utf-8 -*-
from sqlalchemy import Column, Integer, Sequence, String, UnicodeText
from sqlalchemy.orm import relationship
from ssh2.constants import AuthMethodType
from ssh2.models.base import BaseModel
from ssh2.utils import uuid_str
from ssh2.utils.crypto import EncryptHandler, b64decode, b64encode

encrypt_handler = EncryptHandler()


class AuthMethod(BaseModel):
    __tablename__ = "auth_method"

    id = Column("id", Integer, Sequence("auth_method_id_seq"), primary_key=True)
    name = Column("name", String(32), unique=True, default=uuid_str)
    type = Column("auth_method_type", String(32))
    content = Column("content", UnicodeText, nullable=True)
    expect_for_password = Column("expect_for_password", String(32), nullable=True)

    configs = relationship("ClientConfig", back_populates="auth")

    def to_json(self):
        return dict(
            kind="AuthMethod",
            ref=dict(
                field="id",
                value=self.id,
            ),
            spec=dict(
                name=self.name,
                type=self.type,
                # 增加一层 b64encode
                content=b64encode(self.content_decrypted)
                if self.type == AuthMethodType.PUBLISH_KEY_CONTENT
                else self.content_decrypted,
                expect_for_password=self.expect_for_password,
            ),
        )

    @classmethod
    def from_publishkey_file(cls, file_path: str, save_private_key_in_db: bool, name=None):
        if save_private_key_in_db:
            with open(file_path, mode="r") as fh:
                content = "".join(fh.readlines())
                type = AuthMethodType.PUBLISH_KEY_CONTENT.value
        else:
            content = file_path
            type = AuthMethodType.PUBLISH_KEY_PATH.value
        return cls(type=type, content=encrypt_handler.encrypt(b64encode(content)), name=name)

    @classmethod
    def from_publishkey_content(cls, content, name=None):
        return cls(
            type=AuthMethodType.PUBLISH_KEY_CONTENT.value,
            content=encrypt_handler.encrypt(content),
            name=name,
        )

    @classmethod
    def from_password(cls, password: str, expect_for_password="", name=None):
        return cls(
            type=AuthMethodType.PASSWORD.value,
            content=encrypt_handler.encrypt(b64encode(password)),
            expect_for_password=expect_for_password,
            name=name,
        )

    @classmethod
    def interactive(cls, expect_for_password="", name=None):
        return cls(
            type=AuthMethodType.INTERACTIVE_PASSWORD.value,
            content=encrypt_handler.encrypt(""),
            expect_for_password=expect_for_password,
            name=name,
        )

    @property
    def content_decrypted(self) -> str:
        return b64decode(encrypt_handler.decrypt(self.content))

    def __str__(self):
        return f"AuthMethod<{self.id}: [{self.name}-{self.type}]>"
