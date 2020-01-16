# 增强 SSH
ssh连接管理工具
项目背景: 方便管理 ssh 连接，降低登录服务器的成本
使用场景:
-   管理 ssh 账号、密码、秘钥等配置
-   一键登录服务器
-   基于 expect 指令进行额外操作(e.g. 通过跳板机登录运维机)

## 使用说明
### 如何安装该项目
本项目使用 [poetry](https://python-poetry.org/) 管理依赖和打包, 目前该项目未发布至任何 pypi 仓库, 因此只能从源码安装。
1. 请根据 [官方文档](https://python-poetry.org/docs/#installation) 安装 poetry
2. 下载源码
```bash
#!/usr/bin/env bash
git clone github https://github.com/shabbywu/ssh2.git
```
3. 使用 poetry 打包项目
```bash
#!/usr/bin/env bash
# 假设你刚执行完 git clone
cd ssh2
poetry build
```
4. 安装   
p.s. 推荐使用 [pipx](https://pipxproject.github.io/pipx/) 管理基于 pip 安装的命令
```bash
#!/usr/bin/env bash
# 假设你刚执行 poetry build
cd dist
## 简单使用
pip install ssh2-0.1.0.tar.gz
## 使用 pipx
pipx install ssh2-0.1.0.tar.gz

## 加载 ssh2_wrapper.sh 内置的指令
bash
source $(ssh2 get-wrapper-dot-sh)

```
5. 使用(demo)
```bash
cat > demo.yal <<< EOF
## Create with a nested object
kind: Session
spec:
    tag: str
    name: str
    plugins:
        -   kind:   SSH_LOGIN
            args:
    client:
        spec:
            user: username_whose_login_to_server
            name: unique_name_to_mark_this_client
            auth:
                spec:
                    name: unique_name_to_mark_this_auth
                    type: PASSWORD
                    content: your_password
                    expect_for_password: str
    server:
        spec:
            name: unique_name_to_mark_this_server
            host: host_of_server
            port: port_of_server
---
# Create with multi object
kind: ClientConfig
spec:
  name: unique_name_to_mark_this_client
  user: username_whose_login_to_server
  auth:
    spec:
      name: unique_name_to_mark_this_auth
      type: INTERACTIVE_PASSWORD
      content: 'a placeholder'
---
kind: ServerConfig
spec:
  name: unique_name_to_mark_this_server
  host: host_of_server
  port: port_of_server
---
kind: Session
spec:
    tag: mnet2
    name: 腾讯-跳板机
    plugins:
        -   kind:   SSH_WETERM
            args:
    client:
      ref:
        field: name
        value: unique_name_to_mark_this_client
    server:
      ref:
        field: name
        value: unique_name_to_mark_this_server
EOF
```
## 附录
### 数据结构
```yaml
---
kind: AuthMethod
spec:
    name: str | nullable
    type: str
    content: str
    expect_for_password: str
    save_private_key_in_db: bool
---
kind: ClientConfig
spec:
    user: str
    name: str | nullable
    auth:
        ref:
            field: id/name
            value: int/str
        spec:
            name: str | nullable
            type: str
            content: str
            expect_for_password: str
            save_private_key_in_db: bool
---
kind: ServerConfig
spec:
    name: str
    host: str
    port: int

---
kind: Session
spec:
    tag: str
    name: str
    plugins:
        -   kind:   SSH_LOGIN
            args:
        -   kind:   EXPECT
            args:
                expect: str
                send:   str
                raw:
                -   str
                -   str
                -   str
    client:
        ref:
            field: id/name
            value: int/str
        spec:
            user: str
            name: str | nullable
            auth:
                ref:
                    field: id/name
                    value: int/str
                spec:
                    name: str | nullable
                    type: str
                    content: password
                    expect_for_password: str
                    save_private_key_in_db: bool
    server:
        ref:
            field: id/name
            value: int/str
        spec:
            name: str
            host: str
            port: int
```
### 项目建模:
**AuthMethod**: 连接服务器时, 进行身份验证的方法(PASSWORD、PUBLISH_KEY等)
**ClientConfig**: 连接服务器时, 使用的身份信息(username), 关联着 AuthMethod
**ServerConfig**: 连接的服务器信息, 包括(host、port)
**Session**: ssh会话配置, 描述了使用哪个ClientConfig连接哪个ServerConfig的信息
项目整体结构:
```text
                                                                                             +--------------------+
                                        +-------------+       +-------------+                |  +--------------+  |
                                        | config.yaml | --+-- | config.yaml |                |  |              |  |
                                        +-------------+   |   +-------------+                |  |  AuthMethod  |  |
                                                          |                                  |  |              |  |
                                                          |                                  |  +-------+------+  |
                                                          v                                  |          ^         |
                                                    +-----+------+                           |          v         |
                                    cretae/update   |            |                           |  +-------+------+  |
                                 +----------------->+ YamlParser |                           |  |              |  |
                                 |                  |            |                           |  | ClientConfig |  |
                                 |                  +-----+------+                           |  |              |  |
                                 |                        |                                  |  +-------+------+  |
                                 |                        |                                  |          |         |
                                 |                        v                                  |          |         |
+-----------------+         +----+--------+         +-----+-------+           +----------+   |    +-----+---+     |
|                 | invoke  |             |  read   |             |           |          |   |    |         |     |
|  shell-wrapper  +-------->+  Click-cli  +-------->+ sqlalchemy  +---------->+  models  |   |    | Session |     |
|                 |         |             |         |             |           |          |   |    |         |     |
+-----+-----------+         +-------------+         +-------------+           +----------+   |    +-----+---+     |
      ^                                                                                      |          |         |
      |                                                                                      |          |         |
      |                                             +-------------+                          |  +-------+------+  |
      |    eval              +-----------+ generate |             |           bind           |  |              |  |
      +--------------------+ | expect.sh |  <----+  |   plugins   +<----------------------------+ ServerConfig |  |
                             +-----------+          |             |                          |  |              |  |
                                  ..                +-------------+                          |  +--------------+  |
                                  ..                |             |                          |                    |
                             +-----------+          | *SSH_LOGIN  |                          +--------------------+
                             | expect.sh |          |             |
                             +-----------+          | *SSH_WETERM |
                                  ..                |             |
                                  ..                | *EXPECT     |
                                  ..                |             |
                                  ..                +-------------+
                             +-----------+
                             | expect.sh |
                             +-----------+

```