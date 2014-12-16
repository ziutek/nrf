// nrfbitbang shows how to use nRF24L01+ transceiver connected to PC using USB and
// FT232RL module.
//
// Connections (FT232RL -- nRF24L01+):
// TxD (DBUS0) -- MISO
// RxD (DBUS1) -- IRQ
// RTS (DBUS2) -- SCK
// CTS (DBUS3) -- MOSI
// DTR (DBUS4) -- CE
// DSR (DBUS5) -- CSN
package main

import (
	"github.com/ziutek/ftdi"
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	d, err := ftdi.OpenFirst(0x0403, 0x6001, ftdi.ChannelAny)
	checkErr(d.SetBitmode(0xff, ftdi.ModeSyncBB))
	
	// nRF24L01+ SPI clock should be <= 8 MHz
	checkErr(d.SetBaudrate(192))
}