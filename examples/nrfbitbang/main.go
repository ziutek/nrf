// nrfbitbang is example how to use nRF24L01+ transceiver connected to PC using
// USB and FT232RL module.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/ziutek/bitbang/spi"
	"github.com/ziutek/ftdi"
	"github.com/ziutek/nrf"
)

// Connections (FT232RL -- nRF24L01+):
// TxD (DBUS0) -- MISO
// RxD (DBUS1) -- IRQ
// RTS (DBUS2) -- SCK
// CTS (DBUS3) -- MOSI
// DTR (DBUS4) -- CE
// DSR (DBUS5) -- CSN
const (
	MISO = 0x01
	IRQ  = 0x02
	SCK  = 0x04
	MOSI = 0x08
	CE   = 0x10
	CSN  = 0x20
)

func die(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func checkErr(err error) {
	if err == nil {
		return
	}
	die(err)
}

type spiDrv struct {
	io.Reader
	*bufio.Writer
}

type nrfDrv struct {
	*spi.Master
	ce, csn byte
}

func (d *nrfDrv) Enable(en bool) error {
	var base byte
	prePost := []byte{d.csn}
	if en {
		prePost[0] |= d.ce
		base = d.ce
	}
	// Setup CSN, CE lines before and after conversation.
	d.SetPrePost(prePost, prePost)
	// Setup CSN, CE line during conversation.
	d.SetBase(base)
	_, err := d.WriteRead(prePost, nil)
	return err
}

func newNrfDrv(ma *spi.Master, ce, csn byte) (*nrfDrv, error) {
	d := &nrfDrv{Master: ma, ce: ce, csn: csn}
	return d, d.Enable(false)
}

func setup(udev *ftdi.USBDev) nrf.Device {
	ft, err := ftdi.OpenUSBDev(udev, ftdi.ChannelAny)
	checkErr(err)
	checkErr(ft.SetBitmode(SCK|MOSI|CE|CSN, ftdi.ModeSyncBB))

	checkErr(ft.SetBaudrate(512 * 1024 / 16))
	const cs = 4096
	checkErr(ft.SetReadChunkSize(cs))
	checkErr(ft.SetWriteChunkSize(cs))
	checkErr(ft.SetLatencyTimer(2))
	checkErr(ft.PurgeBuffers())

	ma := spi.NewMaster(
		&spiDrv{ft, bufio.NewWriterSize(ft, cs)},
		SCK, MOSI, MISO,
	)
	cfg := spi.Config{
		Mode:     spi.MSBF | spi.CPOL0 | spi.CPHA0,
		FrameLen: 1,
		Delay:    0,
	}
	ma.Configure(cfg)

	drv, err := newNrfDrv(ma, CE, CSN)
	checkErr(err)

	return nrf.Device{drv}
}

func info(radios []nrf.Device) {
	for i, radio := range radios {
		cfg, sta, err := radio.Config()
		checkErr(err)
		aa, _, err := radio.AA()
		checkErr(err)
		rae, _, err := radio.RxAddrEn()
		checkErr(err)
		aw, _, err := radio.AW()
		checkErr(err)
		cnt, dlyus, _, err := radio.Retr()
		checkErr(err)
		fmt.Printf(
			"%c:\n"+
				" Cfg: %s\n"+
				" EnAA: %s\n"+
				" EnRxAddr: %s\n"+
				" AW: %d\n"+
				" Retr: %d times, %d us\n"+
				" Status: %s\n",
			'A'+i, cfg, aa, rae, aw, cnt, dlyus, sta,
		)
	}
}

func main() {
	udevs, err := ftdi.FindAll(0x0403, 0x6001)
	checkErr(err)

	if len(udevs) < 2 {
		die("Need two devices but", len(udevs), "detected.")
	}
	A := setup(udevs[0])
	B := setup(udevs[1])
	radios := []nrf.Device{A, B}

	info(radios)
	_, err = A.SetRetr(15, 500)
	checkErr(err)
	_, err = A.SetConfig(nrf.EnCRC | nrf.CRCO | nrf.PwrUp)
	checkErr(err)
	_, err = B.SetRetr(15, 500)
	checkErr(err)
	_, err = B.SetConfig(nrf.EnCRC | nrf.CRCO | nrf.PwrUp | nrf.PrimRx)
	checkErr(err)
	info(radios)
}
