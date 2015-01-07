// nrfbitbang shows how to use nRF24L01+ transceiver connected to PC using USB
// and FT232RL module.
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"

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
// GNF         -- GND
// 3V3         -- VCC (decoupling required)
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
	debug bool
	r     io.Reader
	w     *bufio.Writer
}

func (d *spiDrv) Read(b []byte) (int, error) {
	if d.debug {
		defer fmt.Printf("spiread: %x\n", b)
	}
	return d.r.Read(b)
}

func (d *spiDrv) Write(b []byte) (int, error) {
	if d.debug {
		fmt.Printf("spiwrite: %x\n", b)
	}
	return d.w.Write(b)
}

func (d *spiDrv) Flush() error {
	if d.debug {
		fmt.Println("spiflush")
	}
	return d.w.Flush()
}

type nrfDrv struct {
	*spi.Master
	ce, csn byte
}

// Enable(true);Enable(false) sets CE bit for period of 3 bit-bang symbols.
// If you need such sequence set bit-bang baudrate < 3 sym / 10 us = 300000 Baud
// to satisfy nRF24L01(+) timing or wait enough between calls.
func (d *nrfDrv) SetCE(up bool) error {
	var ce byte
	if up {
		ce = d.ce
	}
	base := d.csn | ce
	prePost := []byte{base}
	// Setup CSN, CE lines before and after conversation.
	d.SetPrePost(prePost, prePost)
	// Setup CSN, CE lines during conversation.
	d.SetBase(base)

	// This bit-bangs []byte{base, base, base}.
	_, err := d.WriteRead()
	if err == nil {
		err = d.Flush()
	}

	// Clear CSN in base.
	d.SetBase(ce)

	return err
}

func newNrfDrv(ma *spi.Master, ce, csn byte) (*nrfDrv, error) {
	d := &nrfDrv{Master: ma, ce: ce, csn: csn}
	return d, d.SetCE(false)
}

func setup(udev *ftdi.USBDev) (nrf.Device, *spiDrv) {
	ft, err := ftdi.OpenUSBDev(udev, ftdi.ChannelAny)
	checkErr(err)
	checkErr(ft.SetBitmode(SCK|MOSI|CE|CSN, ftdi.ModeSyncBB))

	checkErr(ft.SetBaudrate(8 * 1024 / 16))
	const cs = 4096
	checkErr(ft.SetReadChunkSize(cs))
	checkErr(ft.SetWriteChunkSize(cs))
	checkErr(ft.SetLatencyTimer(2))
	checkErr(ft.PurgeBuffers())

	spid := &spiDrv{r: ft, w: bufio.NewWriterSize(ft, cs)}
	ma := spi.NewMaster(
		spid,
		SCK, MOSI, MISO,
	)
	cfg := spi.Config{
		Mode:     spi.MSBF | spi.CPOL0 | spi.CPHA0,
		FrameLen: 1,
		Delay:    0,
	}
	ma.Configure(cfg)

	nrfd, err := newNrfDrv(ma, CE, CSN)
	checkErr(err)

	return nrf.Device{nrfd}, spid
}

func info(radios []nrf.Device) {
	for i, radio := range radios {
		cfg, stat, err := radio.Cfg()
		checkErr(err)
		aa, _, err := radio.AA()
		checkErr(err)
		rxae, _, err := radio.RxAE()
		checkErr(err)
		aw, _, err := radio.AW()
		checkErr(err)
		cnt, dlyus, _, err := radio.Retr()
		checkErr(err)
		ch, _, err := radio.Ch()
		checkErr(err)
		rf, _, err := radio.RF()
		checkErr(err)
		plos, arc, _, err := radio.TxCnt()
		checkErr(err)
		rpd, _, err := radio.RPD()
		checkErr(err)
		rpds := "< -64dBm"
		if rpd {
			rpds = "> -64dBm"
		}
		var a0, a1, txa [5]byte
		_, err = radio.RxAddr(0, a0[:])
		checkErr(err)
		_, err = radio.RxAddr(1, a1[:])
		checkErr(err)
		a2, _, err := radio.RxAddr0(2)
		checkErr(err)
		a3, _, err := radio.RxAddr0(3)
		checkErr(err)
		a4, _, err := radio.RxAddr0(4)
		checkErr(err)
		a5, _, err := radio.RxAddr0(5)
		checkErr(err)
		_, err = radio.TxAddr(txa[:])
		checkErr(err)
		var pw [6]int
		for i := range pw {
			pw[i], _, err = radio.RxPW(i)
			checkErr(err)
		}
		fifo, _, err := radio.FIFO()
		checkErr(err)
		dynpd, _, err := radio.DynPD()
		checkErr(err)
		feature, _, err := radio.Feature()
		checkErr(err)

		fmt.Printf(
			"Radio %c registers:\n"+
				" Cfg:   %s\n"+
				" AA:    %s\n"+
				" RxAE:  %s\n"+
				" AW:    %d\n"+
				" Retr:  %d times, %d us\n"+
				" Ch:    %d\n"+
				" RF:    %s\n"+
				" Stat:  %s\n"+
				" TxCnt: %d pkt lost, %d retr\n"+
				" RPD:   %t (%s)\n"+
				" Addr0: %x\n"+
				" Addr1: %x\n"+
				" Addr2: %x\n"+
				" Addr3: %x\n"+
				" Addr4: %x\n"+
				" Addr5: %x\n"+
				" TxAddr:%x\n",
			'A'+i,
			cfg, aa, rxae, aw,
			cnt, dlyus,
			ch, rf, stat,
			plos, arc,
			rpd, rpds,
			a0, a1, a2, a3, a4, a5, txa,
		)
		for i, pw := range pw {
			fmt.Printf(" PW%d:   %d\n", i, pw)
		}
		fmt.Printf(
			" FIFO:  %s\n"+
				" DynPD: %s\n"+
				" Fature:%s\n",
			fifo, dynpd, feature,
		)
	}
}

func main() {
	udevs, err := ftdi.FindAll(0x0403, 0x6001)
	checkErr(err)

	if len(udevs) < 2 {
		die("Need two devices but", len(udevs), "detected.")
	}
	A, _ := setup(udevs[0])
	B, _ := setup(udevs[1])
	radios := []nrf.Device{A, B}

	fmt.Println("\nBefore configuration\n")
	info(radios)

	cfg := nrf.EnCRC | nrf.CRCO | nrf.PwrUp
	future := nrf.DPL
	ch := 125 // max. 125
	rf := nrf.LNAHC | nrf.DRLow | nrf.Pwr(-18)
	//rf := nrf.LNAHC | nrf.Pwr(-18)
	retr := 15
	var dlyus int
	if future&nrf.AckPay != 0 {
		if rf&nrf.DRLow != 0 {
			dlyus = 1500
		} else {
			dlyus = 500
		}
	} else {
		if rf&nrf.DRLow != 0 {
			dlyus = 500
		} else {
			dlyus = 250
		}
	}

	for _, radio := range radios {
		_, err = radio.SetRF(rf)
		checkErr(err)
		_, err = radio.SetCh(ch)
		checkErr(err)
		_, err = radio.SetRetr(retr, dlyus)
		checkErr(err)
		_, err = radio.SetFeature(future)
		checkErr(err)
		_, err = radio.SetDynPD(nrf.PAll)
		checkErr(err)
		_, err = radio.SetRxAE(nrf.P0)
		checkErr(err)
	}
	_, err = A.SetCfg(cfg)
	checkErr(err)
	_, err = B.SetCfg(cfg | nrf.PrimRx)
	checkErr(err)

	fmt.Println("\nAfter configuration\n")
	info(radios)

	fmt.Println("\nTransmision\n")

	go func() {
		time.Sleep(time.Second)
		var (
			buf  [32]byte
			lost int
		)
		for k := 0; ; k++ {
			_, err := A.WriteTxP(buf[:])
			checkErr(err)
			checkErr(A.SetCE(true))
			checkErr(A.SetCE(false))

			buf[31]++
			for i := 31; i > 0; i-- {
				if buf[i] < 10 {
					break
				}
				buf[i] = 0
				buf[i-1]++
			}

			for i := 0; ; i++ {
				fifo, stat, err := A.FIFO()
				checkErr(err)
				if i&0xff == 0 {
					fmt.Println("A:", lost, "/", k, stat, fifo)
				}
				if stat&nrf.MaxRT != 0 {
					_, err := A.Clear(nrf.MaxRT)
					checkErr(err)
					_, err = A.FlushTx()
					checkErr(err)
					lost++
					break
				}
				if stat&nrf.FullTx == 0 {
					break
				}
				_, err = A.FlushRx()
				checkErr(err)
			}
		}

		/*
			_, err = A.ReuseTxP()
			checkErr(err)
			checkErr(A.SetCE(true))
			checkErr(A.SetCE(false))
			time.Sleep(time.Second)
			checkErr(A.SetCE(true))
			checkErr(A.SetCE(false))
		*/
	}()

	checkErr(B.SetCE(true))
	var buf [32]byte
	for i := 0; ; i++ {
		plen, stat, err := B.RxPLen()
		checkErr(err)
		if plen > 32 {
			fmt.Println("B: ", plen, "> 32")
			_, err = B.FlushRx()
			checkErr(err)
			continue
		}
		fifo, stat, err := B.FIFO()
		checkErr(err)
		if fifo&nrf.RxEmpty == 0 {
			fmt.Println("B: ", stat, "Plen:", plen)
			checkErr(err)
			_, err = B.ReadRxP(buf[:plen])
			checkErr(err)
			_, err = B.Clear(nrf.RxDR)
			checkErr(err)
			fmt.Println("B: ", buf[:plen])
		} else if i&0xff == 0 {
			fmt.Println("B: ", stat, fifo)
		}
	}
}
