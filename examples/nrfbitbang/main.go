// nrfbitbang is example how to use nRF24L01+ transceiver connected to PC using USB and
// FT232RL module.
package main

import (
	"fmt"
	"os"
	"time"

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
	d, err := ftdi.OpenFirst(0x0403, 0x6001, ftdi.ChannelAny)
	checkErr(err)

	
	checkErr(d.SetBitmode(SCK|MOSI|CE|CSN, ftdi.ModeSyncBB))

	// nRF24L01+ SPI clock should be <= 8 MHz.
	// FT232R max baudrate is 3 MBaud, USB speed is 12 Mb/s.
	// Theoretical available USB speed is 12 Mb/s (sum in+out).
	// Theoretical max baudrate in one direction: 12 MBaud / 8 = 1500 kBaud.
	// Use 1500 kBaud / 2 = 750 kBaud
	checkErr(d.SetBaudrate(750e3 / 16))

	time.Sleep(time.Hour)
}
