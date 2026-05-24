package wifi

/*
extern int esp_wifi_set_mode(int mode);

typedef void (*software_reset_fn)(void);
static void tinygo_software_reset(void) {
	((software_reset_fn)0x400006d8)();
}
*/
import "C"

import (
	"net/netip"
	"time"

	"esp32s3-demo/src/config"
	"esp32s3-demo/src/diag"
	"esp32s3-demo/src/portal"
	"esp32s3-demo/src/rgb"
	"esp32s3-demo/src/store"
	"github.com/soypat/lneto"
	"github.com/soypat/lneto/dhcpv4"
	"tinygo.org/x/espradio"
)

var (
	dhcpServer dhcpv4.Server
	pendingCfg config.Credentials
	hasPending bool
	started    bool
)

func Run(forceSetup bool) {
	if !forceSetup {
		if cached, ok, err := store.Load(); ok {
			diag.Cache(true, "")
			diag.Log("trying cached WiFi:", cached.SSID)
			if connectSTA(cached) {
				return
			}
		} else if err != nil {
			diag.Cache(false, err.Error())
			diag.Log("cached WiFi unavailable:", err.Error())
		} else {
			diag.Cache(false, "")
			diag.Log("no cached WiFi")
		}
	} else {
		diag.Log("force setup requested, skipping cached WiFi")
	}

	if !forceSetup && config.DefaultSSID != "" {
		diag.Log("trying built-in WiFi:", config.DefaultSSID)
		if connectSTA(config.Credentials{SSID: config.DefaultSSID, Password: config.DefaultPassword}) {
			return
		}
	}

	startSetupAP()
}

func connectSTA(cfg config.Credentials) bool {
	diag.Mode("STA")
	diag.SSID(cfg.SSID)
	diag.Log("starting station mode")
	if started {
		if code := C.esp_wifi_set_mode(1); code != 0 {
			diag.Error("station mode switch failed: " + diag.Int(int(code)))
			return false
		}
	} else {
		if err := espradio.Start(); err != nil {
			diag.Error("station start failed: " + err.Error())
			return false
		}
		started = true
	}

	diag.Log("connecting WiFi:", cfg.SSID)
	if err := espradio.Connect(espradio.STAConfig{SSID: cfg.SSID, Password: cfg.Password}); err != nil {
		diag.Error("WiFi connect failed: " + err.Error())
		return false
	}

	diag.Log("WiFi connected:", cfg.SSID)
	if cfg.APIKey != "" {
		diag.APIKeyReceived(true)
		diag.Log("API key received")
	}

	diag.Log("starting station netdev")
	nd, err := espradio.StartNetDev()
	if err != nil {
		diag.Error("station netdev failed: " + err.Error())
		return false
	}

	diag.Log("creating station stack")
	stack, err := espradio.NewStack(nd, espradio.StackConfig{
		Hostname:    "tinygo-s3",
		MaxUDPPorts: 2,
		MaxTCPPorts: 1,
	})
	if err != nil {
		diag.Error("station stack failed: " + err.Error())
		return false
	}

	go pollStack(stack)

	diag.Log("starting DHCP client")
	dhcp, err := stack.SetupWithDHCP(espradio.DHCPConfig{})
	if err != nil {
		diag.Error("DHCP client failed: " + err.Error())
		return false
	}
	diag.IP(dhcp.AssignedAddr.String())
	diag.Log("got IP:", dhcp.AssignedAddr.String())

	lstack := stack.LnetoStack()
	rstack := lstack.StackRetrying(lneto.BackoffStrategy(func(_ uint) time.Duration {
		return config.PollTime
	}))
	gatewayHW, err := rstack.DoResolveHardwareAddress6(dhcp.Router, 500*time.Millisecond, 4)
	if err != nil {
		diag.Error("gateway ARP failed: " + err.Error())
		return false
	}
	lstack.SetGateway6(gatewayHW)

	go portal.Serve(lstack, dhcp.AssignedAddr, savePending)
	rgb.WiFiConnected()

	for {
		if hasPending {
			hasPending = false
			diag.Log("new config saved in STA mode, rebooting")
			time.Sleep(300 * time.Millisecond)
			C.tinygo_software_reset()
		}
		time.Sleep(5 * time.Second)
		diag.Log("WiFi OK:", cfg.SSID)
	}
}

func startSetupAP() {
	diag.Mode("AP")
	diag.SSID(config.SetupSSID)
	diag.IP(config.APIP)
	rgb.SetupMode()
	diag.Log("starting setup AP:", config.SetupSSID)
	if err := espradio.StartAP(espradio.APConfig{
		SSID:     config.SetupSSID,
		Channel:  6,
		AuthOpen: true,
	}); err != nil {
		diag.Error("AP start failed: " + err.Error())
		rgb.Error()
		return
	}
	started = true

	diag.Log("starting AP netdev")
	nd, err := espradio.StartNetDevAP()
	if err != nil {
		diag.Error("AP netdev failed: " + err.Error())
		rgb.Error()
		return
	}

	addr := netip.MustParseAddr(config.APIP)
	diag.Log("creating AP stack")
	stack, err := espradio.NewStack(nd, espradio.StackConfig{
		Hostname:      config.SetupSSID,
		StaticAddress: addr,
		MaxUDPPorts:   2,
		MaxTCPPorts:   1,
	})
	if err != nil {
		diag.Error("stack failed: " + err.Error())
		rgb.Error()
		return
	}

	diag.Log("configuring DHCP")
	if err := dhcpServer.Configure(dhcpv4.ServerConfig{
		ServerAddr: addr.As4(),
		Gateway:    addr.As4(),
		DNS:        addr.As4(),
		Subnet:     netip.MustParsePrefix("192.168.4.0/24"),
	}); err != nil {
		diag.Error("DHCP configure failed: " + err.Error())
		return
	}
	if err := stack.LnetoStack().RegisterUDP(&dhcpServer, nil, dhcpv4.DefaultClientPort); err != nil {
		diag.Error("DHCP register failed: " + err.Error())
		return
	}

	go pollStack(stack)
	go portal.Serve(stack.LnetoStack(), addr, savePending)

	diag.Log("setup portal ready")
	diag.Log("connect phone to:", config.SetupSSID)
	diag.Log("open: http://" + config.APIP + "/")

	for {
		if hasPending {
			cfg := pendingCfg
			hasPending = false
			diag.Log("form submitted, trying WiFi:", cfg.SSID)
			if connectSTA(cfg) {
				return
			}
			diag.Log("returned to AP setup mode")
			rgb.SetupMode()
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func savePending(cfg config.Credentials) {
	if err := store.Save(cfg); err != nil {
		diag.Cache(true, err.Error())
		diag.Error("cache save failed: " + err.Error())
	} else {
		diag.Cache(true, "")
		diag.Log("credentials cached on device")
	}
	pendingCfg = cfg
	hasPending = true
	if cfg.APIKey != "" {
		diag.APIKeyReceived(true)
	}
}

func pollStack(stack *espradio.Stack) {
	for {
		send, recv, err := stack.RecvAndSend()
		if err != nil {
			diag.Error("stack: " + err.Error())
		}
		if send == 0 && recv == 0 {
			time.Sleep(config.PollTime)
		}
	}
}
