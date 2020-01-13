# -*- coding: utf-8 -*-
from sqlalchemy import Column, Integer, String, ForeignKey, Sequence
from sqlalchemy.orm import relationship

from ssh2.utils import uuid_str
from ssh2.models import BaseModel


class ClientConfig(BaseModel):
    __tablename__ = 'client_config'

    id = Column("id", Integer, Sequence("client_config_id_seq"),  primary_key=True)
    name = Column("name", String(32), unique=True, default=uuid_str)

    user = Column("user", String)
    auth_method_id = Column(Integer, ForeignKey('auth_method.id'))

    auth = relationship("AuthMethod", back_populates="configs")
    sessions = relationship("Session", back_populates="client")

    def __str__(self):
        return f"ClientConfig<{self.id}: [{self.name}-{self.user}]>"

    def to_json(self):
        return dict(
            kind="ClientConfig",
            filter_by="id",
            filter_value=self.id,
            spec=dict(name=self.name,
                      user=self.user,
                      auth=self.auth.to_json(),))
