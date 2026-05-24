package portal

import (
	"bytes"

	"esp32s3-demo/src/config"
)

func parseForm(body []byte) config.Credentials {
	return config.Credentials{
		SSID:     formValue(body, "ssid", config.MaxSSIDBytes),
		Password: formValue(body, "password", config.MaxPasswordBytes),
		APIKey:   formValue(body, "apikey", config.MaxAPIKeyBytes),
	}
}

func formValue(body []byte, key string, max int) string {
	prefix := append([]byte(key), '=')
	for len(body) > 0 {
		part := body
		if i := bytes.IndexByte(body, '&'); i >= 0 {
			part = body[:i]
			body = body[i+1:]
		} else {
			body = nil
		}
		if bytes.HasPrefix(part, prefix) {
			var out [config.MaxAPIKeyBytes]byte
			if max > len(out) {
				max = len(out)
			}
			n := urlDecode(out[:], part[len(prefix):])
			if n > max {
				n = max
			}
			return string(out[:n])
		}
	}
	return ""
}

func urlDecode(dst, src []byte) int {
	n := 0
	for i := 0; i < len(src) && n < len(dst); i++ {
		switch src[i] {
		case '+':
			dst[n] = ' '
			n++
		case '%':
			if i+2 < len(src) {
				hi, ok1 := hexVal(src[i+1])
				lo, ok2 := hexVal(src[i+2])
				if ok1 && ok2 {
					dst[n] = hi<<4 | lo
					n++
					i += 2
				}
			}
		default:
			dst[n] = src[i]
			n++
		}
	}
	return n
}

func hexVal(b byte) (byte, bool) {
	switch {
	case b >= '0' && b <= '9':
		return b - '0', true
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10, true
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10, true
	default:
		return 0, false
	}
}
