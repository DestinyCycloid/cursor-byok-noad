# 无广告版本编译指南

本文说明如何在 Windows 环境下编译不包含本地广告功能的版本。

无广告版本使用两个编译开关：

- Go 构建标签：`noads`
- 前端构建变量：`NO_ADS=true`

两者必须同时启用。只设置其中一个是不完整的：

- 只设置 Go 标签，前端仍可能包含广告入口；
- 只设置 `NO_ADS`，后端仍可能包含广告服务。

## 1. 编译结果

无广告版本是一个 Wails 二合一桌面程序：

- Vue 前端会被编译到 `frontend/dist`；
- Go 会通过 `embed` 将前端资源嵌入 exe；
- exe 同时包含桌面 UI、本地代理、嵌入式 Backend、模型路由和 Agent 功能。

无广告版本不会：

- 请求 `ads.leokun.cn`；
- 启动广告定时刷新；
- 创建或更新广告缓存；
- 挂载本地 `/ad` 广告资源路由；
- 加载主页广告 Provider。

## 2. 前置环境

需要安装：

- Windows 10/11；
- Go（版本要求见 [`go.mod`](../go.mod)）；
- Node.js；
- Yarn；
- Task；
- Wails 3 CLI，版本应与 `go.mod` 中的 Wails 依赖匹配；
- `protoc`；
- `protoc-gen-go`；
- `protoc-gen-connect-go`；
- Windows WebView2 Runtime；
- Git Bash 或 MSYS2（Taskfile 使用 `mkdir`、`rm`、`cp`、`zip` 等 Unix 命令）。

在 Git Bash 中可以先检查：

```bash
go version
node --version
yarn --version
task --version
wails3 version
protoc --version
protoc-gen-go --version
protoc-gen-connect-go --version
```

建议将 Go 工具目录加入 PATH：

```bash
export PATH="$HOME/go/bin:$PATH"
```

Windows PowerShell 中对应目录通常是：

```text
%USERPROFILE%\go\bin
```

## 3. 推荐方式：使用 Task 编译

从仓库根目录执行：

```bash
cd /c/Home/demo/CB
task build:windows:amd64:noads
```

该任务会依次完成：

1. 整理 Go 依赖；
2. 安装前端依赖；
3. 生成 Protobuf Go/Connect 代码；
4. 生成 Wails 前端 bindings；
5. 使用 `NO_ADS=true` 构建前端；
6. 生成 Windows 图标资源；
7. 使用 `production,noads` 编译 Go 程序；
8. 将 exe 打包成 ZIP。

输出文件：

```text
bin/windows-64-noads.zip
```

解压后即可运行其中的 exe。

### 其他 Windows 架构

Windows 32 位：

```bash
task build:windows:386:noads
```

当前 Windows 架构：

```bash
task build:windows:current:noads
```

对应输出通常为：

```text
bin/windows-32-noads.zip
bin/windows-64-noads.zip
```

## 4. 重要：版本号必须正确注入

不要直接使用下面这种命令作为发布构建：

```bash
go build -tags 'production,noads' .
```

因为 [`internal/buildinfo/buildinfo.go`](../internal/buildinfo/buildinfo.go) 中的默认版本是：

```go
var Version = "0.0.0"
```

如果没有使用 linker flag 覆盖它，程序会以 `0.0.0` 运行。更新器会把这个版本判断为非常旧，可能触发更新提示或更新流程。

Taskfile 会从 [`build/config.yml`](../build/config.yml) 读取版本，例如：

```yaml
info:
  version: "0.0.42"
```

并自动注入：

```text
-X cursor/internal/buildinfo.Version=0.0.42
```

因此，**正式编译必须优先使用 `task build:windows:amd64:noads`**，不要绕过 Taskfile 手动执行不完整的 `go build`。

## 5. 手动编译方式

如果没有 Task，也可以按以下步骤手动构建，但必须自己完成版本注入。

### 5.1 生成协议代码

从仓库根目录执行：

```bash
protoc \
  -I ./proto \
  --go_out=. \
  --go_opt=module=cursor \
  --connect-go_out=. \
  --connect-go_opt=module=cursor \
  ./proto/agent_v1.proto \
  ./proto/aiserver_v1.proto
```

生成结果位于被 Git 忽略的 `gen/` 目录。

### 5.2 生成 Wails bindings

```bash
wails3 generate bindings -f '-tags production,noads' -clean=true
```

无广告绑定中应保留核心服务绑定，例如：

```text
frontend/bindings/cursor/internal/bridge/metricsservice.js
frontend/bindings/cursor/internal/bridge/proxyservice.js
frontend/bindings/cursor/internal/bridge/windowservice.js
```

不应生成广告服务绑定：

```text
frontend/bindings/cursor/internal/bridge/adservice.js
```

### 5.3 构建无广告前端

```bash
NO_ADS=true yarn --cwd frontend run build
```

确认构建成功后，`frontend/dist` 才能作为 Go 的嵌入资源。

可以检查前端产物是否包含广告标记：

```bash
if grep -R -n -E 'ads\.leokun\.cn|AdModelProvider|ad:updated|cursor:open-ad' frontend/dist; then
  echo "发现广告内容，停止构建"
  exit 1
else
  echo "前端产物未发现广告标记"
fi
```

### 5.4 注入版本并编译 exe

假设 `build/config.yml` 当前版本是 `0.0.42`：

```bash
mkdir -p bin
GOOS=windows \
CGO_ENABLED=0 \
GOARCH=amd64 \
go build \
  -tags 'production,noads' \
  -trimpath \
  -ldflags '-w -s -H windowsgui -X cursor/internal/buildinfo.Version=0.0.42' \
  -o bin/cursor-byok-windows-64-noads.exe .
```

编译结束后，可以删除 `frontend/dist`，因为它是中间产物；exe 已经将其嵌入。

> Windows 原生 shell 下环境变量写法不同。推荐在 Git Bash 执行上面的命令，或直接使用仓库的 Taskfile。

## 6. 验证编译结果

### 6.1 验证 exe 文件存在

```bash
ls -lh bin/cursor-byok-windows-64-noads.exe
```

### 6.2 验证版本号

确保编译命令包含 linker flag：

```text
-X cursor/internal/buildinfo.Version=<当前版本>
```

不要让最终程序保留默认的 `0.0.0` 版本。

### 6.3 检查 exe 中的广告标记

可以使用 Python 检查二进制：

```bash
python - <<'PY'
from pathlib import Path

path = Path("bin/cursor-byok-windows-64-noads.exe")
data = path.read_bytes()
markers = [
    b"ads.leokun.cn",
    b"ad:updated",
    b"AdModelProvider",
    b"cursor:open-ad",
    b"/ad/",
]

for marker in markers:
    if marker in data:
        raise SystemExit(f"发现广告标记: {marker.decode()}")
    print(f"未发现: {marker.decode()}")
PY
```

### 6.4 编译检查

使用无广告标签运行项目已有测试/编译检查：

```bash
go test -tags 'production,noads' ./...
```

这一步至少可以确认无广告条件下所有 Go 包能够编译并通过现有测试。

## 7. 运行程序

直接运行项目目录下的 exe：

```bash
./bin/cursor-byok-windows-64-noads.exe
```

PowerShell：

```powershell
& "C:\Home\demo\CB\bin\cursor-byok-windows-64-noads.exe"
```

程序启动后，前端和后端已经在同一个 exe 中，不需要单独启动前端开发服务器或 Backend 进程。

默认本地地址：

```text
MITM 代理：127.0.0.1:18080
Backend：127.0.0.1:18090
```

## 8. 构建产物和 Git 忽略

项目的 `.gitignore` 已忽略：

```gitignore
bin
```

因此 exe 和 ZIP 应放在：

```text
bin/
```

检查是否被忽略：

```bash
git status --short --ignored bin
```

如果看到：

```text
!! bin/
```

说明构建产物不会进入 Git 提交。

## 9. 常见错误

### `frontend/dist` 不存在

先构建前端：

```bash
NO_ADS=true yarn --cwd frontend run build
```

### 缺少 `adservice.js`

通常是 Wails bindings 和前端构建模式不一致。重新生成无广告 bindings：

```bash
wails3 generate bindings -f '-tags production,noads' -clean=true
```

### exe 显示版本 `0.0.0`

说明漏了版本 linker flag。使用 Taskfile，或手动添加：

```text
-X cursor/internal/buildinfo.Version=<版本号>
```

### 无广告前端仍出现广告标记

确认构建时使用了：

```bash
NO_ADS=true
```

并删除旧的 `frontend/dist` 后重新构建：

```bash
rm -rf frontend/dist
NO_ADS=true yarn --cwd frontend run build
```

### Windows 无法启动

检查：

- WebView2 Runtime 是否已安装；
- 是否使用正确的 Windows 架构；
- 是否被杀毒软件拦截；
- 代理端口 `18080` 和 Backend 端口 `18090` 是否被占用。
