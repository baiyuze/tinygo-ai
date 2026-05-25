# ESP32-S3 TinyGo AI Voice Plan

## Verdict

The original all-on-device TinyGo plan is not executable as written, but an ESP32-S3 can call DeepSeek/OpenAI-compatible APIs directly if the firmware stack provides real HTTPS/TLS.

- `tinygo.org/x/drivers/i2s` is not a package in the current TinyGo drivers module.
- The installed TinyGo ESP32-S3 target does not expose `machine.I2S`, so INMP441 capture cannot be built with the normal TinyGo I2S API yet.
- `github.com/go-kws/s3kws` cannot be fetched from GitHub, so it is not a usable dependency.
- The current TinyGo `espradio`/`lneto` path has Go `net/http` examples, but the TLS path is still device/netdev dependent. In the current dependency set, `_IPPROTO_TLS` is accepted but is routed through the same TCP socket path, so DeepSeek HTTPS is not a reliable production path from this TinyGo app.

## Direct-Cloud Architecture

For a true "ESP32-S3 calls AI itself" build, use ESP-IDF or Arduino-ESP32 for the cloud-call layer:

1. ESP32-S3 runs TinyGo firmware.
2. WiFi connects in STA mode.
3. Firmware creates an HTTPS client with certificate validation.
4. Firmware POSTs JSON to `https://api.deepseek.com/chat/completions` or `https://api.deepseek.com/v1/chat/completions`.
5. Firmware parses `choices[0].message.content`.
6. ESP32-S3 displays the answer on the ST7789 screen.

This is feasible with Arduino-ESP32 `WiFiClientSecure` + `HTTPClient`, or ESP-IDF `esp_http_client`. It is not just prompt concatenation: the mandatory part is a working TLS client.

## TinyGo Architecture

The TinyGo path in this repo remains useful for the display, WiFi setup portal, cached config, and local HTTP endpoints.

1. ESP32-S3 runs TinyGo firmware.
2. ESP32-S3 handles WiFi setup, credential cache, display, and a small HTTP endpoint.
3. Local text or a nearby gateway can post text to `POST /ai`.
4. ESP32-S3 displays the AI answer on the ST7789 screen.

This keeps the current TinyGo codebase working while the HTTPS/TLS layer is either replaced or added through an ESP-IDF/TLS shim.

## Current Wiring

Display wiring already used by the project:

- CS -> GPIO10
- DC -> GPIO11
- RST -> GPIO12
- SCL -> GPIO13
- SDA -> GPIO14
- VCC -> 3.3V
- GND -> GND
- BL -> 3.3V

Proposed INMP441 wiring can stay reserved for the later audio stage:

- VDD -> 3.3V
- GND -> GND
- SD -> GPIO20
- SCK -> GPIO21
- WS -> GPIO47
- LR -> GPIO48

## Run

Flash the firmware first:

```powershell
tinygo flash -target=esp32s3-generic -port=COMx .
```

After the board is reachable on WiFi, run the gateway from this repo:

```powershell
$env:DEEPSEEK_API_KEY="your-rotated-key"
go run .\cmd\deepseek-gateway -device http://<esp32-ip> -prompt "用一句话介绍 TinyGo"
```

For setup AP mode, the default device URL is:

```powershell
go run .\cmd\deepseek-gateway -prompt "hello"
```

The gateway defaults to `https://api.deepseek.com/v1` and model `deepseek-chat`.

## Direct DeepSeek Request Shape

The actual API call is ordinary JSON over HTTPS:

```http
POST /chat/completions HTTP/1.1
Host: api.deepseek.com
Authorization: Bearer <api-key>
Content-Type: application/json

{"model":"deepseek-chat","messages":[{"role":"user","content":"hello"}],"stream":false}
```

On ESP32 this should be implemented with:

- Arduino-ESP32: `WiFiClientSecure`, `client.setCACert(...)`, `HTTPClient`, JSON request body.
- ESP-IDF: `esp_http_client`, CA bundle or pinned CA certificate, POST body, response callback.
- TinyGo future path: add a real TLS-capable netdev or an ESP-IDF `esp_http_client` C shim that TinyGo calls.

## Next Step For Voice

The next voice milestone should be one of these:

- Add an ESP-IDF/C shim for ESP32-S3 I2S RX, then feed PCM frames into a KWS engine that is confirmed to build under TinyGo.
- Move speech-to-text and wake-word detection to the gateway, with the ESP32-S3 staying as display/control hardware.
- Switch firmware stack for the voice/cloud path to ESP-IDF/Arduino if INMP441 + offline KWS + direct DeepSeek must run fully on the MCU today.

Do not store API keys in source code or flash unless there is no alternative.
