// nrfbitbang shows how to use nRF24L01+ transceiver connected to PC using USB
// and FT232RL module.
package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/ziutek/bitbang/spi"
	"github.com/ziutek/ftdi"
	"github.com/ziutek/nrf"
)

// Connections (FT232RL -- nRF24L01+):
// TxD (DBUS0) -- CSN
// RxD (DBUS1) -- CE
// RTS (DBUS2) -- MOSI
// CTS (DBUS3) -- SCK
// DTR (DBUS4) -- IRQ
// DSR (DBUS5) -- MISO
// GNF         -- GND
// 3V3         -- VCC (decoupling required, eg: 47 µF + 4.7 nF)
// You can connect VCC to USB 5V using serial LED (red or green, 20 mA) to
// decrase VCC to 3.5 V in idle state, 2.5 V in transmit/receive. Thanks to LED
// you can easly observe power conumption in every state (strong decoupling
// required between LED and VCC, eg: 100…220 µF electr. + 4…22 nF ceramic).
const (
	CSN  = 0x01
	CE   = 0x02
	MOSI = 0x04
	SCK  = 0x08
	IRQ  = 0x10
	MISO = 0x20
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
	r     *ftdi.Device
	w     *bufio.Writer

	numr, numw int
}

func (d *spiDrv) Read(b []byte) (n int, err error) {
	if d.debug {
		defer fmt.Printf("spiread: %x\n", b)
	}
	return d.r.Read(b)
}

func (d *spiDrv) Write(b []byte) (n int, err error) {
	if d.debug {
		fmt.Printf("spiwrite: %x\n", b)
	}
	return d.w.Write(b)
}

func (d *spiDrv) IRQ() (bool, error) {
	b, err := d.r.Pins()
	return b&IRQ == 0, err
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

// SetCE(0) sets CE low, SetCE(1) sets CE high, SetCE(2) sets CE high for 6
// periods and next sets CE low. Bit-bang baudrate must be < (6 sym / 10 µs) =
// 600000 Baud to satisfy nRF24L01(+) spec.
func (d *nrfDrv) SetCE(v int) error {
	base := d.Base()
	prePost, _ := d.PrePost()
	var (
		b   byte
		pre []byte
	)
	switch v {
	case 0: // Set CE low.
		base &^= CE
		prePost[0] &^= CE
		b = base | CSN
	case 1: // Set CE high.
		base |= CE
		prePost[0] |= CE
		b = base | CSN
	case 2: // Pulse CE: 111110.
		b = base | CE | CSN
		base &^= CE
		prePost[0] &^= CE
		pre = []byte{b, b, b, b, b, b}
		b &^= CE
	default:
		panic("v<0 || v>2")
	}
	d.SetPrePost(pre, nil)
	d.SetBase(b)
	_, err := d.WriteRead()
	d.SetBase(base)
	d.SetPrePost(prePost, prePost)
	return err
}

func newNrfDrv(ma *spi.Master, ce, csn byte) (*nrfDrv, error) {
	d := &nrfDrv{Master: ma, ce: ce, csn: csn}
	// Set CSN high before and after transaction
	prePost := []byte{CSN}
	d.SetPrePost(prePost, prePost)
	return d, d.SetCE(0)
}

func setup(udev *ftdi.USBDev) (nrf.Device, *spiDrv) {
	ft, err := ftdi.OpenUSBDev(udev, ftdi.ChannelAny)
	checkErr(err)
	checkErr(ft.SetBitmode(SCK|MOSI|CE|CSN, ftdi.ModeSyncBB))

	checkErr(ft.SetBaudrate(512 * 1024 / 16))
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
	for i, udev := range udevs {
		fmt.Printf("%c: %s\n", 'A'+i, udev.Serial)
	}
	if len(udevs) < 2 {
		die("Need two devices but", len(udevs), "detected.")
	}
	A, spiA := setup(udevs[0])
	B, spiB := setup(udevs[1])
	radios := []nrf.Device{A, B}

	fmt.Println("\nBefore configuration\n")
	info(radios)

	cfg := nrf.EnCRC | nrf.CRCO | nrf.PwrUp
	future := nrf.DPL
	ch := 125 // max. 125
	rf := nrf.LNAHC | nrf.DRLow | nrf.Pwr(-18)
	//rf := nrf.LNAHC | nrf.Pwr(-12)
	//rf := nrf.LNAHC | nrf.DRHigh | nrf.Pwr(-6)
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
		_, err = radio.SetFeature(future)
		checkErr(err)
		_, err = radio.SetRF(rf)
		checkErr(err)
		_, err = radio.SetCh(ch)
		checkErr(err)
		_, err = radio.SetRetr(retr, dlyus)
		checkErr(err)
		_, err = radio.SetDynPD(nrf.P0)
		checkErr(err)
		_, err = radio.SetAA(nrf.P0)
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

	fmt.Println("\nTransmission\n")

	go func() {
		//spiA.debug = true
		var (
			buf  [32]byte
			lost int
		)
		time.Sleep(5 * time.Millisecond)
		// You can kept CE high all the time. This works very realiable
		// and speedups transmission but incrases power consumption.
		// checkErr(A.SetCE(1))
		for n := 0; ; n++ {
			_, err = A.WriteTxP(buf[:])
			checkErr(err)

			// Don't use SetCE(1);sleep(10µs);SetCE(0).
			// 
			// Delay beetwen setting CE line high and next low isn't generally
			// realiable (it can be only in case of realtime OS and carefully
			// written application). Such seqeunce causes strange behavior of
			// nRF24 PTX:
			//   RxDR- TxDS- MaxRT+ FullTx+ RxPipe: 3.
			// Tx FIFO can't be flushed, MaxRT can't be cleared, data from
			// Rx FIFO can't be read or flushed.
			//
			// SetCE(2) works realiable only if write buffering guarantees
			// that it cause only single system call.
			//
			// Next source of problem is stability of FT232R baudrate. It seems
			// to work realiabe if baudrate is set to power of two.
			checkErr(A.SetCE(2))

			buf[31]++
			for i := 31; i > 0; i-- {
				if buf[i] < 10 {
					break
				}
				buf[i] = 0
				buf[i-1]++
			}

			for {
				irq, err := spiA.IRQ()
				checkErr(err)
				if irq {
					break
				}
			}
			stat, err := A.NOP()
			checkErr(err)
			if stat&nrf.MaxRT != 0 {
				isrMaxRT(A, "A")
				lost++
			}
			if stat&nrf.TxDS != 0 {
				isrTxDS(A, "A", n, lost)
			}
			if stat&nrf.RxDR != 0 {
				isrRxDR(A, "!A")
			}
		}
	}()

	//spiB.debug = true
	time.Sleep(5 * time.Millisecond)
	checkErr(B.SetCE(1))
	for {
		irq, err := spiB.IRQ()
		checkErr(err)
		if !irq {
			continue
		}
		stat, err := B.NOP()
		checkErr(err)
		if stat&nrf.MaxRT != 0 {
			isrMaxRT(B, "!B")
		}
		if stat&nrf.TxDS != 0 {
			isrTxDS(B, "!B", -1, -1)
		}
		if stat&nrf.RxDR != 0 {
			isrRxDR(B, "B")
		}
	}
}

func isrMaxRT(dev nrf.Device, name string) {
	_, err := dev.Clear(nrf.MaxRT)
	checkErr(err)
	_, err = dev.FlushTx()
	checkErr(err)
	stat, err := dev.NOP()
	checkErr(err)
	fmt.Printf("%s: MaxRT %s\n", name, stat)
}

func isrTxDS(dev nrf.Device, name string, n, lost int) {
	_, err := dev.Clear(nrf.TxDS)
	checkErr(err)
	fmt.Printf("%s: TxDS n=%d lost=%d\n", name, n, lost)
}

func isrRxDR(dev nrf.Device, name string) {
	var buf [32]byte
	for {
		plen, stat, err := dev.RxPLen()
		checkErr(err)
		if plen > 32 {
			fmt.Printf("%s: pipe=%d plen=%d>32\n", name, stat.RxPipe(), plen)
			_, err = dev.FlushRx()
			checkErr(err)
		} else {
			_, err = dev.ReadRxP(buf[:plen])
			checkErr(err)
			fmt.Printf("%s: pipe=%d %v\n", name, stat.RxPipe(), buf[:plen])
		}
		_, err = dev.Clear(nrf.RxDR)
		checkErr(err)
		fifo, _, err := dev.FIFO()
		checkErr(err)
		if fifo&nrf.RxEmpty != 0 {
			break
		}
	}
}
