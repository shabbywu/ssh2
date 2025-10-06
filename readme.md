# 增强 SSH
ssh连接管理工具
项目背景: 方便管理 ssh 连接，降低登录服务器的成本
使用场景:
-   管理 ssh 账号、密码、秘钥等配置
-   一键登录服务器
-   基于 expect 指令进行额外操作(e.g. 通过跳板机登录运维机)

## 使用说明
### 1. 如何安装该项目
本项目基于 golang 1.22 开发, 可以直接通过以下命令安装

```bash
> go install github.comshabbywu/ssh2
```

### 2. 使用示例
```yml
# file `demo.yml`
## Create with a nested object
kind: Session
spec:
    tag: session-1
    name: unique_name_to_mark_this_session_1
    plugins:
        -   kind:   SSH_LOGIN
            args:
    client:
        spec:
            user: username_whose_login_to_server
            name: unique_name_to_mark_this_client_1
            auth:
                spec:
                    name: unique_name_to_mark_this_auth_1
                    type: PASSWORD
                    content: your_password
                    expect_for_password: str
    server:
        spec:
            name: unique_name_to_mark_this_server_1
            host: host_of_server
            port: port_of_server
---
# Create with multi object
kind: ClientConfig
spec:
  name: unique_name_to_mark_this_client_2
  user: username_whose_login_to_server
  auth:
    spec:
      name: unique_name_to_mark_this_auth_2
      type: INTERACTIVE_PASSWORD
      content: 'a placeholder'
---
kind: ServerConfig
spec:
  name: unique_name_to_mark_this_server_2
  host: host_of_server
  port: port_of_server
---
kind: Session
spec:
    tag: session-2
    name: unique_name_to_mark_this_session_2
    plugins:
        -   kind:   SSH_WETERM
            args:
    client:
      ref:
        field: name
        value: unique_name_to_mark_this_client_2
    server:
      ref:
        field: name
        value: unique_name_to_mark_this_server_2
```
```bash
# 录入登录配置
ssh2 create -f demo.yml
# 登录 session-1 服务器
ssh2 login session-1
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
