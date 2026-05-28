package app

import (
	"machine"
	"time"

	"esp32s3-demo/src/boot"
	"esp32s3-demo/src/diag"
	"esp32s3-demo/src/display"
	"esp32s3-demo/src/micverify"
	"esp32s3-demo/src/rgb"
	"esp32s3-demo/src/wifi"
	"tinygo.org/x/espradio"
)

func Run() {
	machine.Serial.Configure(machine.UARTConfig{BaudRate: 115200})
	time.Sleep(time.Second)

	diag.Log("boot")
	if micverify.Enabled {
		rgb.Setup()
		display.Setup()
		micverify.Run()
		return
	}

	diag.Log("enabling radio")
	if err := espradio.Enable(espradio.Config{Logging: espradio.LogLevelError}); err != nil {
		diag.Error("radio enable failed: " + err.Error())
		return
	}
	diag.Log("radio enabled")

	rgb.Setup()
	display.Setup()
	display.StartDebugger()
	forceSetup := boot.ForceSetupRequested(3 * time.Second)
	diag.ForceSetup(forceSetup)
	wifi.Run(forceSetup)
}
