package boot

import (
	"time"

	"esp32s3-demo/src/diag"
	"machine"
)

func ForceSetupRequested(hold time.Duration) bool {
	pin := machine.GPIO0
	pin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	if pin.Get() {
		return false
	}

	diag.Log("BOOT held, keep holding to enter setup AP")
	time.Sleep(hold)
	if pin.Get() {
		diag.Log("BOOT released, continue normal boot")
		return false
	}

	diag.Log("force setup AP requested")
	return true
}
