# -*- coding: utf-8 -*-
from typing import TYPE_CHECKING

from sqlalchemy import Column, ForeignKey, Integer, Sequence, String
from sqlalchemy.orm import relationship
from ssh2.models.base import BaseModel
from ssh2.utils import uuid_str

if TYPE_CHECKING:
    from ssh2.models import AuthMethod


class ClientConfig(BaseModel):
    __tablename__ = "client_config"

    id = Column("id", Integer, Sequence("client_config_id_seq"), primary_key=True)
    name = Column("name", String(32), unique=True, default=uuid_str)

    user = Column("user", String)
    auth_method_id = Column(Integer, ForeignKey("auth_method.id"))

    auth: 'AuthMethod' = relationship("AuthMethod", back_populates="configs")
    sessions = relationship("Session", back_populates="client")

    def __str__(self):
        return f"ClientConfig<{self.id}: [{self.name}-{self.user}]>"

    def to_json(self):
        return dict(
            kind="ClientConfig",
            ref=dict(
                field="id",
                value=self.id,
            ),
            spec=dict(
                name=self.name,
                user=self.user,
                auth=self.auth.to_json(),
            ),
        )
