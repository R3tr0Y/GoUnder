# GoUnder

**GoUnder** 是一款由 Go 语言开发的自动化信息搜集与 CDN 绕过工具，适用于网络安全实战中的目标溯源与指纹识别。工具支持命令行与 Web UI 双模式操作，适用于渗透测试、红队演练与执法取证等场景。

---

## 🧠 项目功能

- 🎯 **绕过 CDN**：通过 FOFA 数据接口，结合 `host`、`title`、 `favicon hash` 与 `cert` 策略定位目标真实 IP。
- 🔍 **网站指纹识别**：集成 Wappalyzer 本地库与 WhatCMS API，实现网站技术栈探测。
- 🌐 **图形界面支持**：提供 Web UI 页面，零门槛查询 CDN 与指纹信息。
- 🧩 **模块化架构**：可扩展性强，支持更多信息源、指纹引擎与展示方式。

---

## ⚙️ 使用方法

### 🔧 编译与运行

```bash
git clone https://github.com/yourname/GoUnder.git
cd GoUnder
go run main.go
```

或构建可执行文件：

```
go build -o gounder main.go
./gounder --help
```

### 🔌 CDN 绕过命令示例

```
go run main.go cdn -u http://hscks.com/ -p icon
go run main.go cdn -u 911blw.com -p title
```

支持参数：

| 参数 | 说明                                |
| ---- | ----------------------------------- |
| `-u` | 目标网站 URL                        |
| `-p` | 查询策略：`host` / `title` / `icon` / `cert` |
| `--log` | 记录查询日志: `false`               |
------

### 🧬 指纹识别命令示例

```
go run main.go fingerprint -u http://example.com -e wappalyzer
go run main.go fingerprint -u http://example.com -e whatcms
```

支持引擎：

- `wappalyzer`：本地识别，速度快，依赖少
- `whatcms`：云端识别，支持更多 CMS 数据

------

### 🌐 启动 Web UI（默认端口 8080）

```
go run main.go webui -p 8080 -a 0.0.0.0
```

访问地址：

```
http://localhost:8080/
```

可在页面中进行：

- CDN 溯源信息查询
- 技术栈指纹识别
- JSON 数据查看与复制

------

## 📂 配置说明

请先配置以下 JSON 文件于 `configs/` 目录：

### FOFA 配置（`configs/fofa.json`）

```
{
  "email": "your_email@example.com",
  "key": "your_fofa_api_key"
}
```

### WhatCMS 配置（`configs/whatcms.json`）

```
{
  "key": "your_whatcms_api_key"
}
```
**如果编译二进制文件运行，则需要设置全局配置文件，请运行程序并根据程序提供的文件路径配置，默认路径：**

```
linux: $HOME/.config/GoUnder

windows: %APPDATA%/GoUnder

mac: $HOME/Library/Application\ Support/GoUnder
```

------

## 🛠 依赖环境

- Go 1.18+
- 第三方依赖：

```
go install github.com/projectdiscovery/wappalyzergo
go get github.com/go-resty/resty/v2
go get github.com/spf13/cobra
go get github.com/gin-gonic/gin
```

---

## 📈 项目结构

```
GoUnder/
├── cmd/
│   ├── cdn.go             # CDN绕过模块
│   ├── fingerprint.go     # 指纹识别模块
│   ├── webui.go           # Web UI模块
│   ├── utils_cmd.go       # 公共函数
│   ├── webui/static/      # 前端资源（静态页面）
├── configs/               # 配置文件目录
├── utils/                 # 工具函数（如icon hash计算）
└── main.go                # 项目入口
```

------

## 🧭 下一步计划

-  增加 Shodan、ZoomEye 支持
-  PDF 报告生成功能
- 批量检测和批量导出

------

## 🔐 法律与合规声明

- 本工具仅用于教育与授权的渗透测试环境；
- 禁止使用本项目从事任何非法活动；
- 使用者需自行遵守所在地区的相关法律法规。

------

## 👨‍💻 开发者

> 本项目由 @RetroYoung 开发。
>
> 如需项目合作或内部部署支持，请联系：Email: retro@stu.ppsuc.edu.cn