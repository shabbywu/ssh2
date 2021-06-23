# -*- coding: utf-8 -*-
import json
from textwrap import dedent
from typing import TYPE_CHECKING, List

from sqlalchemy import Column, ForeignKey, Integer, Sequence, String, Text
from sqlalchemy.orm import relationship
from ssh2.models.base import BaseModel
from ssh2.utils import uuid_str

if TYPE_CHECKING:
    from ssh2.models import ClientConfig, ServerConfig


class Session(BaseModel):
    __tablename__ = "session"

    id = Column("id", Integer, Sequence("session_id_seq"), primary_key=True)
    tag = Column("tag", String(16), index=True, unique=True)
    name = Column("name", String(32), unique=True, default=uuid_str)
    plugins = Column("plugings", Text)

    client_config_id = Column(Integer, ForeignKey("client_config.id"))
    server_config_id = Column(Integer, ForeignKey("server_config.id"))

    client: 'ClientConfig' = relationship("ClientConfig", back_populates="sessions")
    server: 'ServerConfig' = relationship("ServerConfig", back_populates="sessions")

    def to_expect_cmds(self) -> str:
        from ssh2.plugins import BasePlugin

        plugins: List[dict] = json.loads(self.plugins)
        cmds = [
            "set timeout 20",
            dedent(
                """
                trap {
                    set rows [stty rows]
                    set cols [stty columns]
                    stty rows $rows columns $cols < $spawn_out(slave,name)
                } WINCH"""
            ),
        ]

        for plugin_definition in plugins:
            plugin = BasePlugin.from_dict(plugin_definition)
            cmds.extend(plugin.to_expect_cmds(self))

        if "interact" not in cmds:
            cmds.append("interact")

        return "\n".join(cmds)

    def __str__(self):
        return f"Session<{self.id}: [{self.tag}-{self.name}]>"

    def to_json(self):
        plugins: List[dict] = json.loads(self.plugins)

        return dict(
            kind="Session",
            ref=dict(
                field="id",
                value=self.id,
            ),
            spec=dict(
                name=self.name,
                tag=self.tag,
                client=self.client.to_json(),
                server=self.server.to_json(),
                plugins=plugins,
            ),
        )
