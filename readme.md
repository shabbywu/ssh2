# 增强 SSH
ssh连接管理工具
项目背景: 方便管理 ssh 连接，降低登录服务器的成本
使用场景:
-   管理 ssh 账号、密码、秘钥等配置
-   一键登录服务器
-   基于 expect 指令进行额外操作(e.g. 通过跳板机登录运维机)

## 使用说明
### 1. 安装 ssh2 与 go2s
本项目基于 golang 1.22 开发。先安装 `ssh2`，并确保 Go 的 bin 目录已经加入 `PATH`：

```bash
go install github.com/shabbywu/ssh2@latest
ssh2 --help
```

`go2s` 是调用 `ssh2 login` 的 shell wrapper，支持无参数列出 session tag 以及 `-d`/`--direct`，并在 Zsh 和 PowerShell 中提供 tag 补全。wrapper 默认安装到 `~/.ssh/ssh2`；设置了 `SSH2_HOME` 时则使用该目录。

`ssh2 install-ssh2-auto-complete` 可以安装当前平台的默认 wrapper：Windows 安装 PowerShell 脚本，其他平台安装 Bash/Zsh 脚本。该命令只写入脚本，不会修改 shell profile。也可以按下面的方式显式安装并启用。

#### Linux / macOS：Bash

仅在当前终端启用：

```bash
source "$(ssh2 get-wrapper-dot-sh)"
```

永久启用时，将实际脚本路径写入 `~/.bashrc`；macOS 使用 Bash 登录 shell 时可改为 `~/.bash_profile`：

```bash
wrapper="$(ssh2 get-wrapper-dot-sh)"
printf '\nsource "%s"\n' "$wrapper" >> ~/.bashrc  # 只需执行一次
source ~/.bashrc
```

#### Linux / macOS：Zsh

```zsh
# 当前终端
source "$(ssh2 get-wrapper-dot-sh)"

# 永久启用，只需执行一次
wrapper="$(ssh2 get-wrapper-dot-sh)"
printf '\nsource "%s"\n' "$wrapper" >> ~/.zshrc
source ~/.zshrc
```

#### Windows：Windows PowerShell 5.1 / PowerShell 7

仅在当前终端启用：

```powershell
$wrapper = ssh2 get-wrapper-dot-ps1
. $wrapper
```

永久启用时，将 dot-source 命令写入当前 PowerShell 的 `$PROFILE`。Windows PowerShell 和 PowerShell 7 使用不同的 profile；如果两者都使用，需要分别执行：

```powershell
$wrapper = ssh2 get-wrapper-dot-ps1
$profileDirectory = Split-Path -Parent $PROFILE
if (-not (Test-Path -LiteralPath $profileDirectory)) {
    New-Item -ItemType Directory -Path $profileDirectory -Force | Out-Null
}
if (-not (Test-Path -LiteralPath $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force | Out-Null
}
$escapedWrapper = $wrapper.Replace("'", "''")
$loadLine = ". '$escapedWrapper'"
if (-not (Select-String -LiteralPath $PROFILE -SimpleMatch $loadLine -Quiet)) {
    Add-Content -LiteralPath $PROFILE -Value $loadLine
}
. $PROFILE
```

如果 PowerShell 提示脚本执行被禁用，可以在确认本机安全策略后为当前用户启用本地脚本，再重新加载 profile：

```powershell
Set-ExecutionPolicy -Scope CurrentUser RemoteSigned
. $PROFILE
```

Windows `cmd.exe` 不提供 `go2s` wrapper，可以直接使用 `ssh2 login <tag>`。Git Bash 用户可使用 Unix 脚本，并用 `cygpath` 转换 Windows 路径：

```bash
wrapper="$(ssh2 get-wrapper-dot-sh)"
source "$(cygpath -u "$wrapper")"
```

安装完成后可以用以下命令验证：

```text
go2s
go2s session-1
go2s -d session-1
go2s --direct session-1
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
        -   kind:   SSH_LOGIN
            args:
        -   kind:   EXPECT
            args:
                steps:
                    -   expect: jump-host$
                        send: "ssh target\r"
                    -   expect: "password:"
                        send: "target-password\r"
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
# 使用快捷 wrapper 登录，或无参数列出所有 session tag
go2s session-1
go2s
```

`SSH_WETERM` 是历史文档里的外部插件示例，当前 Go 版本未内置；需要类似能力时请使用 `EXPECT.steps` 或后续扩展插件。

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
                steps:
                -   expect: str
                    send: str
                -   expect: str
                    send: str
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
                                    create/update   |            |                           |  +-------+------+  |
                                 +----------------->+ YamlParser |                           |  |              |  |
                                 |                  |            |                           |  | ClientConfig |  |
                                 |                  +-----+------+                           |  |              |  |
                                 |                        |                                  |  +-------+------+  |
                                 |                        |                                  |          |         |
                                 |                        v                                  |          |         |
+-----------------+         +----+--------+         +-----+-------+           +----------+   |    +-----+---+     |
|                 | invoke  |             |  read   |             |           |          |   |    |         |     |
|  shell-wrapper  +-------->+  urfave-cli +-------->+  buntdb     +---------->+  models  |   |    | Session |     |
|                 |         |             |         |             |           |          |   |    |         |     |
+-----+-----------+         +-------------+         +-------------+           +----------+   |    +-----+---+     |
      ^                                                                                      |          |         |
      |                                                                                      |          |         |
      |                                             +-------------+                          |  +-------+------+  |
      |    go2s             +-----------+ execute  |             |           bind           |  |              |  |
      +--------------------+ | ssh2 login|  <----+  |   plugins   +<----------------------------+ ServerConfig |  |
                             +-----------+          |             |                          |  |              |  |
                                  ..                +-------------+                          |  +--------------+  |
                                  ..                |             |                          |                    |
                             +-----------+          | *SSH_LOGIN  |                          +--------------------+
                             |    pty    |          |             |
                             +-----------+          | *EXPECT     |
                                  ..                |             |
                                  ..                +-------------+
                             +-----------+
                             | terminal  |
                             +-----------+

```
