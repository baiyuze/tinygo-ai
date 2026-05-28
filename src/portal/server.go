package portal

import (
	"bytes"
	_ "embed"
	"net/netip"
	"strconv"
	"time"

	"esp32s3-demo/src/config"
	"esp32s3-demo/src/diag"
	"esp32s3-demo/src/store"
	"github.com/soypat/lneto/http/httpraw"
	"github.com/soypat/lneto/tcp"
	"github.com/soypat/lneto/x/xnet"
)

//go:embed resources/setup.html
var setupHTML string

func Serve(stack *xnet.StackAsync, listenIP netip.Addr, onSave func(config.Credentials)) {
	tcpPool, err := xnet.NewTCPPool(xnet.TCPPoolConfig{
		PoolSize:           2,
		QueueSize:          2,
		TxBufSize:          4096,
		RxBufSize:          2048,
		EstablishedTimeout: 10 * time.Second,
		ClosingTimeout:     5 * time.Second,
		NewUserData: func() any {
			return new(connState)
		},
	})
	if err != nil {
		diag.Error("tcp pool: " + err.Error())
		return
	}

	var listener tcp.Listener
	if err := listener.Reset(config.HTTPPort, tcpPool); err != nil {
		diag.Error("listener reset: " + err.Error())
		return
	}
	if err := stack.RegisterListener(&listener); err != nil {
		diag.Error("listener register: " + err.Error())
		return
	}
	diag.Log("listening on http://" + listenIP.String() + "/")

	for {
		if listener.NumberOfReadyToAccept() == 0 {
			tcpPool.CheckTimeouts()
			time.Sleep(config.PollTime)
			continue
		}
		conn, userData, err := listener.TryAccept()
		if err != nil {
			diag.Error("accept: " + err.Error())
			continue
		}
		handleConn(conn, userData.(*connState), onSave)
	}
}

type connState struct {
	req  [12288]byte
	resp [12288]byte
	body [4096]byte
	hdr  httpraw.Header
}

func handleConn(conn *tcp.Conn, cs *connState, onSave func(config.Credentials)) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(8 * time.Second))

	req := readRequest(conn, cs)
	cs.hdr.Reset(cs.req[:0])
	cs.hdr.ReadFromBytes(req)
	_, _ = cs.hdr.TryParse(false)
	method := cs.hdr.Method()
	uri := cs.hdr.RequestURI()
	body, _ := cs.hdr.Body()

	diag.Request(string(method), string(uri))
	if bytes.HasPrefix(uri, []byte("/debug")) {
		debugBody := diag.AppendDebugHTML(cs.body[:0])
		writeHTMLBytes(conn, cs.resp[:0], debugBody)
		return
	}
	if bytes.HasPrefix(uri, []byte("/config.json")) {
		writeJSONBytes(conn, cs.resp[:0], appendConfigJSON(cs.body[:0]))
		return
	}

	if bytes.Equal(method, []byte("POST")) && bytes.HasPrefix(uri, []byte("/save")) {
		cfg := parseForm(body)
		if cfg.SSID == "" {
			writeRedirect(conn, cs.resp[:0], "/?status=error")
			return
		}
		onSave(cfg)
		writeRedirect(conn, cs.resp[:0], "/?status=saved")
		return
	}

	writeHTML(conn, cs.resp[:0], setupHTML)
}

func appendConfigJSON(dst []byte) []byte {
	cfg, ok, _ := store.Load()
	dst = append(dst, "{\"hasConfig\":"...)
	if ok {
		dst = append(dst, "true"...)
	} else {
		dst = append(dst, "false"...)
	}
	dst = append(dst, ",\"apiKey\":"...)
	dst = appendJSONString(dst, cfg.APIKey)
	dst = append(dst, ",\"systemPrompt\":"...)
	dst = appendJSONString(dst, cfg.SystemPrompt)
	dst = append(dst, "}"...)
	return dst
}

func appendJSONString(dst []byte, s string) []byte {
	dst = append(dst, '"')
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\', '"':
			dst = append(dst, '\\', s[i])
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			if s[i] >= 0x20 {
				dst = append(dst, s[i])
			}
		}
	}
	dst = append(dst, '"')
	return dst
}

func readRequest(conn *tcp.Conn, cs *connState) []byte {
	req := cs.req[:0]
	headerEnd := -1
	contentLen := 0
	for len(req) < cap(cs.req) {
		n, err := conn.Read(cs.req[len(req):cap(cs.req)])
		if n > 0 {
			req = cs.req[:len(req)+n]
			if headerEnd < 0 {
				headerEnd = bytes.Index(req, []byte("\r\n\r\n"))
				if headerEnd >= 0 {
					contentLen = parseContentLength(req[:headerEnd])
				}
			}
			if headerEnd >= 0 && len(req) >= headerEnd+4+contentLen {
				break
			}
		}
		if err != nil {
			break
		}
	}
	return req
}

func parseContentLength(header []byte) int {
	const key = "Content-Length:"
	idx := bytes.Index(header, []byte(key))
	if idx < 0 {
		return 0
	}
	line := header[idx+len(key):]
	if end := bytes.IndexByte(line, '\r'); end >= 0 {
		line = line[:end]
	}
	n := 0
	for _, b := range line {
		if b >= '0' && b <= '9' {
			n = n*10 + int(b-'0')
		}
	}
	return n
}

func writeHTML(conn *tcp.Conn, buf []byte, body string) {
	writeHTMLBytes(conn, buf, []byte(body))
}

func writeHTMLBytes(conn *tcp.Conn, buf []byte, body []byte) {
	resp := append(buf[:0], "HTTP/1.1 200 OK\r\n"...)
	resp = append(resp, "Content-Type: text/html; charset=utf-8\r\n"...)
	resp = append(resp, "Connection: close\r\n"...)
	resp = append(resp, "Content-Length: "...)
	resp = strconv.AppendInt(resp, int64(len(body)), 10)
	resp = append(resp, "\r\n\r\n"...)
	resp = append(resp, body...)
	conn.Write(resp)
	conn.Flush()
}

func writeJSONBytes(conn *tcp.Conn, buf []byte, body []byte) {
	resp := append(buf[:0], "HTTP/1.1 200 OK\r\n"...)
	resp = append(resp, "Content-Type: application/json; charset=utf-8\r\n"...)
	resp = append(resp, "Cache-Control: no-store\r\n"...)
	resp = append(resp, "Connection: close\r\n"...)
	resp = append(resp, "Content-Length: "...)
	resp = strconv.AppendInt(resp, int64(len(body)), 10)
	resp = append(resp, "\r\n\r\n"...)
	resp = append(resp, body...)
	conn.Write(resp)
	conn.Flush()
}

func writeRedirect(conn *tcp.Conn, buf []byte, location string) {
	resp := append(buf[:0], "HTTP/1.1 303 See Other\r\n"...)
	resp = append(resp, "Location: "...)
	resp = append(resp, location...)
	resp = append(resp, "\r\nConnection: close\r\nContent-Length: 0\r\n\r\n"...)
	conn.Write(resp)
	conn.Flush()
}
