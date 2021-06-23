# -*- coding: utf-8 -*-
import time
from operator import attrgetter, itemgetter
from pathlib import Path

import click
import yaml
from ssh2.cli.parser import YamlParser
from ssh2.models import create_dababases, get_scoped_session, session_scope
from ssh2.models.auth_method import AuthMethod
from ssh2.models.client_config import ClientConfig
from ssh2.models.server_config import ServerConfig
from ssh2.models.session import Session
from ssh2.utils.tempfile import clean_temp_file_core, generate_temp_file

RESOURCE_CLS_MAPPER = {
    "ClientConfig": ClientConfig,
    "AuthMethod": AuthMethod,
    "ServerConfig": ServerConfig,
    "Session": Session,
    "auth": AuthMethod,
    "client": ClientConfig,
    "server": ServerConfig,
    "session": Session,
}


def all_tag():
    try:
        return list(map(itemgetter(0), get_scoped_session().query(Session.tag)))
    except Exception:
        return []


@click.group()
def cli():
    """ssh2 helpers"""
    pass


@cli.command()
@click.argument("resource_type", type=click.Choice(RESOURCE_CLS_MAPPER.keys()))
@click.option("--format", "format", required=False)
def get(resource_type, **options):
    with session_scope() as s:
        resources = s.query(RESOURCE_CLS_MAPPER[resource_type]).all()

    f = str if not options["format"] else attrgetter(options["format"].lstrip("."))
    for instance in resources:
        click.echo(f(instance))


@cli.command()
@click.option("-f", "file", type=click.File("r"), required=True)
def create(file):
    doc = "".join(file.readlines())
    YamlParser(doc).parse()


@cli.command()
@click.argument("resource_type", type=click.Choice(RESOURCE_CLS_MAPPER.keys()))
@click.option("-id", "--id", type=int, default=None)
@click.option("-name", "--name", type=str, default=None)
def edit(resource_type, id, name):
    q = dict(id=id, name=name)
    for key, value in list(q.items()):
        if value is None:
            q.pop(key)

    if not q:
        raise Exception("Must input either one of the name and id")

    with session_scope() as s:
        query = s.query(RESOURCE_CLS_MAPPER[resource_type]).filter_by(**q)
        if query.count() != 1:
            raise Exception("can not edit multi resource as once, please check the filter condition")
        resource = query.scalar()
        json = resource.to_json()

    edited_yaml = click.edit(yaml.dump(json, allow_unicode=True), extension="yaml")

    if edited_yaml:
        click.echo("update by: ")
        click.echo(edited_yaml)
        click.echo("updating...")
        click.echo(YamlParser(edited_yaml).parse()[0].name)
    else:
        click.echo("not modify")


@cli.command()
@click.argument("resource_type", type=click.Choice(RESOURCE_CLS_MAPPER.keys()))
@click.option("-id", "--id", type=int, default=None)
@click.option("-name", "--name", type=str, default=None)
@click.option("-f", "--force", type=bool, default=False)
def delete(resource_type, id, name, force):
    q = dict(id=id, name=name)
    for key, value in list(q.items()):
        if value is None:
            q.pop(key)

    if not q:
        raise Exception("Must input either one of the name and id")

    with session_scope() as s:
        query = s.query(RESOURCE_CLS_MAPPER[resource_type]).filter_by(**q)
        if query.count() != 1 and not force:
            raise Exception(
                "can not delete multi resource as once, " "please check the filter condition or add `--force` option"
            )
        query.delete()


@cli.command()
@click.argument("tag", type=click.Choice(all_tag()))
def quick_login_command(tag):
    session = get_scoped_session().query(Session).filter_by(tag=tag).scalar()
    path = generate_temp_file()
    cmds = session.to_expect_cmds()
    with open(path, "w") as fh:
        fh.write("#!/usr/bin/expect")
        fh.write("\nspawn ssh2 clean-temp-file\n")
        fh.write(cmds)
    click.echo(f"expect -f {path}")


@cli.command()
def clean_temp_file():
    time.sleep(2)
    clean_temp_file_core()


@cli.command()
def init_db():
    create_dababases()


@cli.command()
def get_wrapper_dot_sh():
    current = Path(__file__)
    ssh2_wrapper_dot_sh = current.parent.parent.parent / "ssh2_wrapper.sh"
    if ssh2_wrapper_dot_sh.exists():
        print(ssh2_wrapper_dot_sh.absolute())
    else:
        raise Exception("ssh2_wrapper.sh does not found!")


@cli.command()
def ui():
    pass
