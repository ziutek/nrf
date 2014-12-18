// nrfbitbang is example how to use nRF24L01+ transceiver connected to PC using
// USB and FT232RL module.
package main

import (
	"fmt"
	"os"

	"github.com/ziutek/bitbang/spi"
	"github.com/ziutek/ftdi"
)

// Connections (FT232RL -- nRF24L01+):
// TxD (DBUS0) -- MISO
// RxD (DBUS1) -- IRQ
// RTS (DBUS2) -- SCK
// CTS (DBUS3) -- MOSI
// DTR (DBUS4) -- CE
// DSR (DBUS5) -- CSN
const (
	MISO = 1 << iota
	IRQ
	SCK
	MOSI
	CE
	CSN
)

func checkErr(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	drv, err := ftdi.OpenFirst(0x0403, 0x6001, ftdi.ChannelAny)
	checkErr(err)

	checkErr(drv.SetBitmode(SCK|MOSI|CE|CSN, ftdi.ModeSyncBB))

	// nRF24L01+ SPI clock should be <= 8 MHz.
	// FT232R max baudrate is 3 MBaud, USB speed is 12 Mb/s = 1500 kB/s..
	// In best case: 1308 kB/s fdata + 192 kB/s overhead.
	// Theoretical max baudrate in one direction: 1308 kBaud
	// Use 750 kBaud (48 MHz clock divided by 64).
	checkErr(drv.SetBaudrate(750e3 / 16))

	checkErr(drv.WriteByte(SCK))
	var buf [64]byte
	n, err := drv.Read(buf[:])
	checkErr(err)
	fmt.Println(buf)
	return

	ma := spi.NewMaster(drv, SCK, MOSI, MISO)
	cfg := spi.Config{
		Mode:     spi.MSBF | spi.CPOL0 | spi.CPHA0,
		FrameLen: 1,
		Delay:    0,
	}
	ma.Configure(cfg)

	// CSN is always zero (slave selected).

	checkErr(ma.Begin(nil))
	_, err = ma.WriteN(0xff, 1) // NOP
	checkErr(err)
	n, err = ma.Read(buf[:])
	fmt.Println(buf[:n])
	checkErr(err)
	checkErr(ma.End(nil))

}
