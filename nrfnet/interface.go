package nrfnet

import (
	"github.com/ziutek/nrf"
)

type Interface struct {
	vpi0, vpi1 uint32
	vci        [6]byte
	dev        nrf.Device
}

func NewInterface(dev nrf.Device) (*Interface, error) {
	if _, err := dev.SetFeature(nrf.DPL); err != nil {
		return nil, err
	}
	if _, err := dev.SetCfg(nrf.EnCRC | nrf.CRCO | nrf.PwrUp); err != nil {
		return nil, err
	}
	i := new(Interface)
	i.dev = dev
	return i, dev
}

// It can be called periodically (pulling) or called by ISR.
func (i *Interface) IRQ() {

}

// Addr selects a virtual channel available in real RF channel. Virtual channel
// can be described by two numbers: VPI - virtual path id and VCI - virtual
// channel id (base address and prefix in Nordic nRF51 nomenclature).
type Addr struct {
	VPI uint32
	VCI byte
}

// Connect establishes bidirectional connection to virtual channel (VC) in
// real RF channel. One interface supports up to 6 bidirectional connections
// simultaneously but only when all connected VCs use the same VPI. If two
// different VPI are used, only one bidirectional connection can be 
// established. In this case other connections can be rx-only. All connected 
// VCs must have unique VCIs.
func (i *Interface) Connect(addr Addr) (Conn, error) {

}

// ConnectRx establishes rx-only connection to virtual channel (VC) in real RF
// channel. One interface supports up to 6 rx-only connections simultaneously
// but there are some restrictions:
// 1. All connected VCs must have unique VCIs.
// 2. All connected VCs must share no more than 2 VPIs.
// 3. If there are two different VPIs used, only one of them can address more
//    than one VC.
func (i *Interface) ConnectRx(addr Addr) (Conn, error) {

}
