# TinyGo AI ESP32-S3 配网终端

这是一个基于 ESP32-S3 + TinyGo 的嵌入式 AI 设备底座项目。目前项目重点是把设备的基础能力跑通：WiFi 配网、凭据保存、状态显示、调试页面和 RGB 状态灯。AI API Key 已经可以通过网页录入并缓存，为后续接入云端 AI 服务、语音/屏幕交互或传感器工作流做准备。

## 项目目标

目标是做一个低成本、可复制、可离线维护的 TinyGo AI 硬件终端：

- 设备首次启动或联网失败时自动进入热点配网模式。
- 用户用手机连接设备热点，在网页中填写 WiFi 和 AI API Key。
- 设备保存配置后自动尝试连接家庭/办公室 WiFi。
- 屏幕、RGB 灯和串口日志同步展示当前状态，方便现场调试。
- 后续在这个底座上扩展 AI 对话、传感器采集、远程控制、语音输出等能力。

## 当前效果

当前固件烧录后可以达到这些效果：

- 开机初始化串口、WiFi Radio、RGB 灯和 ST7789 屏幕。
- 屏幕先显示红/绿/蓝自检，再进入调试面板。
- 优先读取 Flash 中缓存的 WiFi 配置并连接。
- 如果缓存不存在、缓存连接失败，或按住 BOOT 约 3 秒强制配网，则开启 `tinygo-s3-setup` 热点。
- 手机连接热点后访问 `http://192.168.4.1/`，可提交 WiFi 名称、密码和 AI API Key。
- 配置会写入 Flash 最后一个 4KB 扇区，下次启动自动使用。
- 连接成功后 RGB 变绿；配置热点模式 RGB 变蓝；严重错误 RGB 变红。
- `http://设备IP/debug` 或 `http://192.168.4.1/debug` 可查看运行状态和最近日志。

## 架构说明

```text
main.go
  -> app.Run()
      -> boot        BOOT 键长按检测，决定是否强制进入配网
      -> rgb         GPIO48 WS2812 状态灯
      -> display     ST7789V2 屏幕初始化、自检、调试面板
      -> wifi        STA/AP 模式切换、DHCP、网络栈轮询
      -> portal      HTTP 配网页面和 debug 页面
      -> store       Flash 扇区读写，保存 WiFi 与 API Key
      -> diag        全局状态、串口日志、屏幕/debug 页面数据源
```

核心流程：

1. `app.Run()` 初始化串口和 `espradio`。
2. `display.Setup()` 初始化 ST7789 屏幕并显示自检画面。
3. `boot.ForceSetupRequested()` 检查 BOOT 键是否被长按。
4. `wifi.Run()` 优先尝试 Flash 缓存配置，其次尝试内置默认 WiFi。
5. 连接失败时进入 AP 配网模式，启动 DHCP 和 HTTP Portal。
6. 表单提交后 `store.Save()` 写入 Flash，随后尝试 STA 连接。
7. 诊断状态通过串口、屏幕和 `/debug` 页面输出。

## 硬件配置

推荐硬件：

| 硬件 | 建议规格 | 说明 |
| --- | --- | --- |
| 主控 | ESP32-S3 开发板，Flash >= 2MB | 项目依赖 ESP32-S3 WiFi 和 TinyGo `espradio` |
| 屏幕 | 2.0 寸 ST7789V2 SPI，240x320 | 当前代码按 GMT020-02-8P 8 针模块配置 |
| RGB 灯 | 板载或外接 WS2812/NeoPixel | 默认使用 GPIO48 |
| 供电 | USB 5V 或稳定 3.3V | 屏幕 VCC/BL 接 3V3 |
| 数据线 | 可传数据 USB-C/Micro USB | 用于烧录和串口日志 |

当前屏幕接线：

| 屏幕引脚 | ESP32-S3 引脚 |
| --- | --- |
| VCC | 3V3 |
| GND | GND |
| BL | 3V3 |
| CS | GPIO10 |
| DC | GPIO11 |
| RST | GPIO12 |
| SCL/SCK | GPIO13 |
| SDA/MOSI | GPIO14 |

其他固定引脚：

| 功能 | 引脚 |
| --- | --- |
| BOOT 强制配网 | GPIO0 |
| RGB 状态灯 | GPIO48 |

## 所需内容

开发环境：

- Go 1.24.x 或兼容版本。
- TinyGo，需支持 ESP32-S3 目标。
- ESP32-S3 USB 串口驱动。
- 可用的 2.4GHz WiFi。
- 后续 AI 功能需要对应服务商的 API Key，目前只负责录入和保存。

Go 依赖已经记录在 `go.mod`：

- `tinygo.org/x/espradio`
- `tinygo.org/x/drivers`
- `tinygo.org/x/tinyfont`
- `github.com/soypat/lneto`

## 构建与烧录

本机需要先安装 TinyGo。当前环境未检测到 `tinygo` 命令，因此下面是当前硬件的建议命令，需要按你的开发板 target 和串口名调整：

```bash
go mod download
tinygo flash -target=ESP32-S3-generetor -port=/dev/tty.usbmodemXXXX .
```

查看串口日志：

```bash
screen /dev/tty.usbmodemXXXX 115200
```

`-target` 必须使用当前硬件对应的 TinyGo target；本项目当前使用 `ESP32-S3-generetor`。如果你的串口名不同，请用实际设备路径替换 `/dev/tty.usbmodemXXXX`。

## 使用方式

1. 烧录固件并重启设备。
2. 观察屏幕自检和 RGB 状态。
3. 如果设备进入蓝灯配网模式，手机连接 WiFi：`tinygo-s3-setup`。
4. 浏览器打开 `http://192.168.4.1/`。
5. 填写 WiFi 名称、密码和 AI API Key，点击保存。
6. 设备连接成功后 RGB 变绿，屏幕显示 STA 模式和 IP。
7. 访问 `http://设备IP/debug` 查看运行日志。

强制重新配网：

1. 按住开发板 BOOT 键。
2. 复位或重新上电。
3. 继续按住约 3 秒，设备会跳过缓存配置并进入热点配网。

## 成本估算

价格会随采购渠道和开发板型号波动，下面按常见电商散件估算：

| 项目 | 估算单价 |
| --- | --- |
| ESP32-S3 开发板 | 25-60 元 |
| 2.0 寸 ST7789V2 SPI 屏幕 | 18-45 元 |
| 杜邦线/排针/小面包板 | 5-20 元 |
| 外接 WS2812 RGB 灯，若开发板无板载灯 | 1-5 元 |
| 外壳、螺丝、亚克力或 3D 打印件 | 10-50 元 |

最小原型成本约 50-130 元；带外壳和更好开发板的版本约 100-200 元。云端 AI 的调用费用另计，目前项目还没有实际发起 AI 请求。

## 当前进度

已完成：

- TinyGo 项目骨架和 ESP32-S3 启动入口。
- `espradio` WiFi STA/AP 初始化。
- 配网热点 `tinygo-s3-setup`。
- AP 模式 DHCP 服务。
- 轻量 HTTP 配网页面。
- WiFi、密码、AI API Key 表单解析。
- Flash 最后 4KB 扇区配置缓存，带 magic、版本和 checksum。
- BOOT 长按强制重新配网。
- ST7789V2 240x320 SPI 屏幕驱动、自检和调试面板。
- 中文日志标题和基础 CJK 字体子集。
- GPIO48 WS2812 RGB 状态灯。
- 串口日志和 `/debug` 诊断页。

待完成：

- 真正接入 AI 服务调用。
- API Key 的加密保存或更安全的凭据管理。
- WiFi 扫描列表和更完整的配置页交互。
- 断线自动重连、失败退避和更细的错误状态。
- OTA 升级或串口外的固件更新方案。
- 传感器、麦克风、扬声器或按键交互。
- 自动化测试和硬件兼容性矩阵。

## 目录说明

```text
.
├── main.go                         # 固件入口
├── src/app/app.go                  # 应用启动编排
├── src/boot/boot.go                # BOOT 强制配网检测
├── src/config/config.go            # 默认配置和尺寸限制
├── src/diag/diag.go                # 状态、日志、debug HTML
├── src/display/                    # ST7789 屏幕、字体、图标、调试 UI
├── src/portal/                     # 配网页面、HTTP 服务、表单解析
├── src/rgb/rgb.go                  # WS2812 RGB 状态灯
├── src/store/store.go              # Flash 配置缓存
├── src/wifi/wifi.go                # WiFi STA/AP、DHCP、网络栈
└── tools/fontsubset/               # CJK 字体子集生成工具
```

## 注意事项

- `src/config/config.go` 中目前包含默认 WiFi 名称和密码，公开仓库前建议删除或改为空。
- API Key 当前是明文保存在 Flash 中，适合原型验证，不适合直接用于量产。
- `store` 模块使用 ESP32 ROM 函数地址读写 Flash，换芯片、换 TinyGo 版本或换启动布局前需要重新验证。
- 屏幕模块默认 `BL -> 3V3`，如果你的屏幕背光需要 GPIO 控制，需要修改 `BacklightPin`。
- 如果屏幕画面偏移，可按代码注释尝试调整 `RowOffset` 为 `0` 或 `80`。
