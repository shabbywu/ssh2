# -*- coding: utf-8 -*-
from sqlalchemy import Column, Integer, String, Sequence
from sqlalchemy.orm import relationship

from ssh2.utils import uuid_str
from ssh2.models import BaseModel


class ServerConfig(BaseModel):
    __tablename__ = 'server_config'

    id = Column("id", Integer, Sequence("server_config_id_seq"), primary_key=True)
    name = Column("name", String(32), unique=True, default=uuid_str)
    host = Column("host", String)
    port = Column("port", Integer)

    sessions = relationship("Session", back_populates="server")

    def __str__(self):
        return f"ServerConfig<{self.id}: [{self.name}-{self.host}:{self.port}]>"

    def to_json(self):
        return dict(
            kind="ServerConfig",
            filter_by="id",
            filter_value=self.id,
            spec=dict(name=self.name,
                      host=self.host,
                      port=self.port,))
