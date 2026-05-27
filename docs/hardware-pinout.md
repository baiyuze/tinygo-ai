# ESP32-S3 当前接线与引脚说明

这份文档记录当前已经插好的 ESP32-S3 原型接线，方便后续排查、复刻和改代码时对照。

当前固件使用 ESP32-S3 + 2.0 寸 ST7789V2 SPI 屏幕 + WS2812 RGB 状态灯。屏幕背光直接接 3V3，不由 GPIO 控制。

## 接线速查

### ST7789V2 屏幕

| 屏幕引脚 | ESP32-S3 引脚 | 作用 | 代码配置 |
| --- | --- | --- | --- |
| VCC | 3V3 | 屏幕供电 | - |
| GND | GND | 共地 | - |
| BL | 3V3 | 背光常亮 | `BacklightPin = machine.NoPin` |
| CS | GPIO10 | SPI 片选 | `CSPin = machine.GPIO10` |
| DC | GPIO11 | 数据/命令选择 | `DCPin = machine.GPIO11` |
| RST | GPIO12 | 屏幕复位 | `ResetPin = machine.GPIO12` |
| SCL / SCK | GPIO13 | SPI 时钟 | `SCKPin = machine.GPIO13` |
| SDA / MOSI | GPIO14 | SPI 数据输出 | `MOSIPin = machine.GPIO14` |

### 其他固定引脚

| 功能 | ESP32-S3 引脚 | 说明 | 代码位置 |
| --- | --- | --- | --- |
| BOOT 强制配网 | GPIO0 | 按住 BOOT 后复位，并继续按住约 3 秒，进入配网热点 | `src/boot/boot.go` |
| WS2812 RGB 状态灯 | GPIO48 | 蓝色表示配网模式，绿色表示 WiFi 已连接，红色表示严重错误 | `src/rgb/rgb.go` |

## 当前硬件连接图

```text
ESP32-S3                    ST7789V2 2.0 inch SPI
--------                    ----------------------
3V3      -----------------> VCC
GND      -----------------> GND
3V3      -----------------> BL
GPIO10   -----------------> CS
GPIO11   -----------------> DC
GPIO12   -----------------> RST
GPIO13   -----------------> SCL / SCK
GPIO14   -----------------> SDA / MOSI

ESP32-S3
--------
GPIO0    <---------------- BOOT 按键
GPIO48   ----------------> WS2812 / NeoPixel DIN
```

## 固件里的对应配置

屏幕引脚集中在 `src/display/hardware.go`：

```go
const (
	Width        = int16(240)
	Height       = int16(320)
	RowOffset    = int16(0)
	ColumnOffset = int16(0)
	SPIFrequency = uint32(40_000_000)

	SCKPin       = machine.GPIO13
	MOSIPin      = machine.GPIO14
	ResetPin     = machine.GPIO12
	DCPin        = machine.GPIO11
	BacklightPin = machine.NoPin
	CSPin        = machine.GPIO10
)
```

SPI 总线使用 `machine.SPI0`，当前只接了写屏需要的方向：

```go
bus.Configure(machine.SPIConfig{
	Frequency: SPIFrequency,
	SCK:       SCKPin,
	SDO:       MOSIPin,
	SDI:       machine.NoPin,
	CS:        machine.NoPin,
	Mode:      machine.Mode0,
})
```

RGB 状态灯固定使用 GPIO48：

```go
pin := machine.GPIO48
```

BOOT 强制配网固定使用 GPIO0，并启用上拉输入：

```go
pin := machine.GPIO0
pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
```

## 上电后的现象

接线正确并烧录当前固件后，设备上电会出现这些现象：

1. 串口输出启动日志。
2. RGB 灯初始化。
3. ST7789V2 屏幕先显示红、绿、蓝自检画面。
4. 屏幕进入调试面板，显示 WiFi 模式、SSID、IP、日志等状态。
5. 如果没有可用 WiFi 配置，设备会开启 `tinygo-s3-setup` 热点，RGB 变蓝。
6. WiFi 连接成功后，RGB 变绿。

## 配网按键

需要强制重新配网时：

1. 按住开发板上的 BOOT 键。
2. 按一下复位键，或者重新上电。
3. 继续按住 BOOT 约 3 秒。
4. 设备跳过 Flash 里的 WiFi 缓存，进入 `tinygo-s3-setup` 热点模式。

手机连接热点后访问：

```text
http://192.168.4.1/
```

调试页面：

```text
http://192.168.4.1/debug
```

WiFi 已连接后，也可以访问：

```text
http://设备IP/debug
```

## 排查清单

### 屏幕完全不亮

- 确认 `VCC -> 3V3`、`GND -> GND`。
- 确认 `BL -> 3V3`。当前背光不是 GPIO 控制，BL 没接时屏幕可能看起来完全没亮。
- 确认 USB 线和开发板供电正常。

### 屏幕有背光但没有画面

- 检查 `SCL/SCK -> GPIO13` 和 `SDA/MOSI -> GPIO14` 是否接反。
- 检查 `CS -> GPIO10`、`DC -> GPIO11`、`RST -> GPIO12`。
- 确认固件使用的是当前 `src/display/hardware.go` 里的引脚配置。
- 如果屏幕有画面但位置偏移，再尝试调整 `RowOffset`，当前值是 `0`。

### RGB 灯不亮

- 确认 WS2812 的 DIN 接 GPIO48。
- 确认 WS2812 供电和 GND 与 ESP32-S3 共地。
- 如果开发板没有板载 GPIO48 RGB，需要外接 WS2812/NeoPixel。

### 无法进入配网模式

- 确认按的是开发板 BOOT 键，对应 GPIO0。
- 操作顺序是先按住 BOOT，再复位或重新上电，继续按住约 3 秒。
- 进入配网后应能看到 `tinygo-s3-setup` 热点。

## 备注

- `src/pinprobe/pinprobe.go` 和 `src/display/scan.go` 里保留了一些早期排针测试组合，不代表当前最终接线。
- 当前已确认的屏幕接线以 `src/display/hardware.go` 和本文档为准。
- 后续如果加入 I2S 麦克风、功放、扬声器或额外按键，需要先避开 GPIO10-14、GPIO0 和 GPIO48。
