package diag

import (
	"strconv"
	"sync"
)

const maxLogs = 40

var (
	mu             sync.Mutex
	mode           = "boot"
	ssid           string
	ip             string
	lastError      string
	lastRequest    string
	forceSetup     bool
	apiKeyReceived bool
	cachePresent   bool
	cacheError     string
	logs           [maxLogs]string
	nextLog        int
	logCount       int
)

type State struct {
	Mode           string
	SSID           string
	IP             string
	LastError      string
	LastRequest    string
	ForceSetup     bool
	APIKeyReceived bool
	CachePresent   bool
	CacheError     string
}

func Log(parts ...string) {
	line := "[tinygo-setup]"
	for _, part := range parts {
		line += " "
		line += part
	}
	println(line)

	mu.Lock()
	logs[nextLog] = line
	nextLog = (nextLog + 1) % len(logs)
	if logCount < len(logs) {
		logCount++
	}
	mu.Unlock()
}

func Snapshot(logDst []string) (State, int) {
	mu.Lock()
	defer mu.Unlock()

	state := State{
		Mode:           mode,
		SSID:           ssid,
		IP:             ip,
		LastError:      lastError,
		LastRequest:    lastRequest,
		ForceSetup:     forceSetup,
		APIKeyReceived: apiKeyReceived,
		CachePresent:   cachePresent,
		CacheError:     cacheError,
	}

	count := logCount
	if count > len(logDst) {
		count = len(logDst)
	}
	start := nextLog - count
	if start < 0 {
		start += len(logs)
	}
	for i := 0; i < count; i++ {
		logDst[i] = logs[(start+i)%len(logs)]
	}
	return state, count
}

func Mode(value string) {
	mu.Lock()
	mode = value
	mu.Unlock()
}

func SSID(value string) {
	mu.Lock()
	ssid = value
	mu.Unlock()
}

func IP(value string) {
	mu.Lock()
	ip = value
	mu.Unlock()
}

func Error(value string) {
	mu.Lock()
	lastError = value
	mu.Unlock()
	Log("error:", value)
}

func Request(method, uri string) {
	mu.Lock()
	lastRequest = method + " " + uri
	mu.Unlock()
	Log("HTTP:", method, uri)
}

func ForceSetup(value bool) {
	mu.Lock()
	forceSetup = value
	mu.Unlock()
}

func APIKeyReceived(value bool) {
	mu.Lock()
	apiKeyReceived = value
	mu.Unlock()
}

func Cache(present bool, err string) {
	mu.Lock()
	cachePresent = present
	cacheError = err
	mu.Unlock()
}

func AppendDebugHTML(dst []byte) []byte {
	mu.Lock()
	defer mu.Unlock()

	dst = append(dst, "<!doctype html><html lang=\"zh-CN\"><head><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><title>ESP32-S3 Debug</title><style>body{font-family:system-ui;margin:20px;line-height:1.45;background:#f7f7f4;color:#151515}table{border-collapse:collapse;width:100%;background:#fff}td{border:1px solid #ddd;padding:8px}pre{white-space:pre-wrap;background:#111;color:#eee;padding:12px;border-radius:6px;font-size:13px}</style></head><body><h2>ESP32-S3 Debug</h2><table>"...)
	dst = appendRow(dst, "mode", mode)
	dst = appendRow(dst, "ssid", ssid)
	dst = appendRow(dst, "ip", ip)
	dst = appendRow(dst, "force setup", boolText(forceSetup))
	dst = appendRow(dst, "api key received", boolText(apiKeyReceived))
	dst = appendRow(dst, "device cache", boolText(cachePresent))
	dst = appendRow(dst, "cache error", cacheError)
	dst = appendRow(dst, "last request", lastRequest)
	dst = appendRow(dst, "last error", lastError)
	dst = append(dst, "</table><h3>Recent Logs</h3><pre>"...)

	start := nextLog - logCount
	if start < 0 {
		start += len(logs)
	}
	for i := 0; i < logCount; i++ {
		idx := (start + i) % len(logs)
		dst = appendEscaped(dst, logs[idx])
		dst = append(dst, '\n')
	}

	dst = append(dst, "</pre><p><a href=\"/\">返回配置页</a></p></body></html>"...)
	return dst
}

func Int(value int) string {
	return strconv.Itoa(value)
}

func appendRow(dst []byte, key, value string) []byte {
	dst = append(dst, "<tr><td>"...)
	dst = appendEscaped(dst, key)
	dst = append(dst, "</td><td>"...)
	dst = appendEscaped(dst, value)
	dst = append(dst, "</td></tr>"...)
	return dst
}

func boolText(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func appendEscaped(dst []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			dst = append(dst, "&amp;"...)
		case '<':
			dst = append(dst, "&lt;"...)
		case '>':
			dst = append(dst, "&gt;"...)
		case '"':
			dst = append(dst, "&quot;"...)
		default:
			dst = append(dst, s[i])
		}
	}
	return dst
}
