# MiraiGo-module-weather

ID: `com.aimerneige.weather`

Module for [MiraiGo-Template](https://github.com/Logiase/MiraiGo-Template)

## 功能

- 在群聊或私聊接受到「实时天气」时查询实时天气情况

## 使用方法

在适当位置引用本包

```go
package example

imports (
    // ...

    _ "github.com/yukichan-bot-module/MiraiGo-module-weather"

    // ...
)

// ...
```

在你的 `application.yaml` 里填入配置：

```yaml
aimerneige:
  weather:
    path: "./weather.yaml" # 配置文件路径，未设置默认为 `./weather.yaml`
```

编辑你的配置文件：

```yaml
key: TAkhjf8d1nlSlspN # api key
limit: 10 # 每人每天访问次数上限
allowed: # 群白名单，在允许列表里才会提供服务
  - 1149558764
  - 857066811
blacklist: # 黑名单用户（不提供服务）
  - 1781924496
whitelist: # 白名单用户（不受调用限制）
  - 1227427929
db:
  type: sqlite # mysql | sqlite
  mysql:
    username: root
    password: password
    host: localhost
    port: 3306
    database: example
    charset: utf8mb4
  sqlite:
    path: "./db/weather.db"
```

## LICENSE

<a href="https://www.gnu.org/licenses/agpl-3.0.en.html">
<img src="https://www.gnu.org/graphics/agplv3-155x51.png">
</a>

本项目使用 `AGPLv3` 协议开源，您可以在 [GitHub](https://github.com/yukichan-bot-module/MiraiGo-module-weather) 获取本项目源代码。为了整个社区的良性发展，我们强烈建议您做到以下几点：

- **间接接触（包括但不限于使用 `Http API` 或 跨进程技术）到本项目的软件使用 `AGPLv3` 开源**
- **不鼓励，不支持一切商业使用**
