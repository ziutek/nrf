package nrf

import "strconv"

func (d Device) byteReg(addr byte) (byte, Stat, error) {
	var buf [1]byte
	stat, err := d.Reg(addr, buf[:])
	return buf[0], stat, err
}

type Stat byte

const (
	FullTx Stat = 1 << iota // Tx FIFO full flag.
	_
	_
	_
	MaxRT // Maximum number of Tx retransmits interrupt.
	TxDS  // Data Sent Tx FIFO interrupt.
	RxDR  // Data Ready Rx FIFO interrupt.
)

// RxPipe returns data pipe number for the payload available for reading from
// RxFifo or -1 if RxFifo is empty
func (s Stat) RxPipe() int {
	n := int(s) & 0x0e
	if n == 0x0e {
		return -1
	}
	return n >> 1
}

func flags(f string, mask, b byte) string {
	buf := make([]byte, len(f))
	m := byte(0x80)
	for i := range buf {
		if f[i] == '+' {
			for mask&m == 0 {
				m >>= 1
			}
			if b&m == 0 {
				buf[i] = '-'
			} else {
				buf[i] = '+'
			}
			m >>= 1
		} else {
			buf[i] = f[i]
		}
	}
	return string(buf)
}

func (s Stat) String() string {
	return flags("RxDR+ TxDS+ MaxRT+ FullTx+ RxPipe:", 0x71, byte(s)) +
		strconv.Itoa(s.RxPipe())
}

type Cfg byte

const (
	PrimRx    Cfg = 1 << iota //  Rx/Tx control 1: PRX, 0: PTX.
	PwrUp                     // 1: power up, 0: power down.
	CRCO                      // CRC encoding scheme 0: one byte, 1: two bytes.
	EnCRC                     // Enable CRC. Force 1 if one of bits in AA is 1.
	MaskMaxRT                 // If 1 then mask interrupt caused by MaxRT.
	MaskTxDS                  // If 1 then mask interrupt caused by TxDS.
	MaskRxDR                  // If 1 then mask interrupt caused by RxDR.
)

func (c Cfg) String() string {
	return flags(
		"Mask(RxDR+ TxDS+ MaxRT+) EnCRC+ CRCO+ PwrUp+ PrimRx+",
		0x7f, byte(c),
	)
}

// Cfg returns value of CONFIG register.
func (d Device) Cfg() (Cfg, Stat, error) {
	cfg, stat, err := d.byteReg(0)
	return Cfg(cfg), stat, err
}

// SetCfg sets value of CONFIG register.
func (d Device) SetCfg(c Cfg) (Stat, error) {
	return d.SetReg(0, byte(c))
}

// Pipe is a bitfield that represents nRF24L01+ Rx data pipes.
type Pipe byte

const (
	P0 Pipe = 1 << iota
	P1
	P2
	P3
	P4
	P5
	PAll = P0 | P1 | P2 | P3 | P4 | P5
)

func (p Pipe) String() string {
	return flags("P5+ P4+ P3+ P2+ P1+ P0+", 0x3f, byte(p))
}

// AA returns value of EN_AA (Enable ‘Auto Acknowledgment’ Function) register.
func (d Device) AA() (Pipe, Stat, error) {
	p, stat, err := d.byteReg(1)
	return Pipe(p), stat, err
}

// SetAA sets value of EN_AA (Enable ‘Auto Acknowledgment’ Function) register.
func (d Device) SetAA(p Pipe) (Stat, error) {
	return d.SetReg(1, byte(p))
}

// RxAE returns value of EN_RXADDR (Enabled RX Addresses) register.
func (d Device) RxAE() (Pipe, Stat, error) {
	p, stat, err := d.byteReg(2)
	return Pipe(p), stat, err
}

// SetRxAE sets value of EN_RXADDR (Enabled RX Addresses) register.
func (d Device) SetRxAE(p Pipe) (Stat, error) {
	return d.SetReg(2, byte(p))
}

// AW returns value of SETUP_AW (Setup of Address Widths) register.
func (d Device) AW() (int, Stat, error) {
	aw, stat, err := d.byteReg(3)
	return int(aw) + 2, stat, err
}

// SetAW sets value of SETUP_AW (Setup of Address Widths) register.
func (d Device) SetAW(aw int) (Stat, error) {
	if aw < 3 || aw > 5 {
		panic("aw<3 || aw>5")
	}
	return d.SetReg(3, byte(aw-2))
}

// Retr returns value of SETUP_RETR (Setup of Automatic Retransmission) reg.
func (d Device) Retr() (cnt, dlyus int, stat Stat, err error) {
	b, stat, err := d.byteReg(4)
	cnt = int(b & 0xf)
	dlyus = (int(b>>4) + 1) * 250
	return
}

// SetRetr sets value of SETUP_RETR (Setup of Automatic Retransmission) reg.
func (d Device) SetRetr(cnt, dlyus int) (Stat, error) {
	if uint(cnt) > 15 {
		panic("cnt<0 || cnt>15")
	}
	if dlyus < 250 || dlyus > 4000 {
		panic("dlyus<250 || dlyus>4000")
	}
	return d.SetReg(4, byte((dlyus/250-1)<<4|cnt))
}

// Ch returns value of RF_CH (RF Channel) register.
func (d Device) Ch() (int, Stat, error) {
	ch, stat, err := d.byteReg(5)
	return int(ch), stat, err
}

// SetCh sets value of RF_CH (RF Channel) register.
func (d Device) SetCh(ch int) (Stat, error) {
	if uint(ch) > 127 {
		panic("ch<0 || ch>127")
	}
	return d.SetReg(5, byte(ch))
}

type RF byte

const (
	LNAHC RF = 1 << iota // (nRF24L01.LNA_HCURR) Rx LNA gain 0: -1.5dB,-0.8mA.
	_
	_
	DRHigh // (RF_DR_HIGH) Select high speed data rate 0: 1Mbps, 1: 2Mbps.
	Lock   // (PLL_LOCK) Force PLL lock signal. Only used in test.
	DRLow  // (RF_DR_LOW) Set RF Data Rate to 250kbps.
	_
	Wave // (CONT_WAVE) Enables continuous carrier transmit when 1.
)

// Pwr returns RF output power in Tx mode [dBm].
func (rf RF) Pwr() int {
	return 3*int(rf&6) - 18
}

func Pwr(dbm int) RF {
	switch {
	case dbm < -18:
		dbm = -18
	case dbm > 0:
		dbm = 0
	}
	return RF((18+dbm)/3) & 6
}

func (rf RF) String() string {
	return flags("Wave+ DRLow+ Lock+ DRHigh+ LNAHC+ Pwr:", 0xb9, byte(rf)) +
		strconv.Itoa(rf.Pwr()) + "dBm"
}

// RF returns value of RF_SETUP register.
func (d Device) RF() (RF, Stat, error) {
	rf, stat, err := d.byteReg(6)
	return RF(rf), stat, err
}

// RF sets value of RF_SETUP register.
func (d Device) SetRF(rf RF) (Stat, error) {
	return d.SetReg(6, byte(rf))
}

// Clear clears specified bits in STATUS register.
func (d Device) Clear(stat Stat) (Stat, error) {
	return d.SetReg(7, byte(stat))
}

// TxCnt returns values of PLOS and ARC counters from OBSERVE_TX register.
func (d Device) TxCnt() (plos, arc int, stat Stat, err error) {
	b, stat, err := d.byteReg(8)
	arc = int(b & 0xf)
	plos = int(b >> 4)
	return
}

// RPD returns value of RPD (Received Power Detector) register (is RP > -64dBm).
// In case of nRF24L01 it returns value of.CD (Carrier Detect) register.
func (d Device) RPD() (bool, Stat, error) {
	b, stat, err := d.byteReg(9)
	return b&1 != 0, stat, err
}

func checkPN(pn int) {
	if uint(pn) > 5 {
		panic("pn<0 || pn>5")
	}
}

func checkAddr(addr []byte) {
	if len(addr) > 5 {
		panic("len(addr)>5")
	}
}

func checkPNA(pn int, addr []byte) {
	checkPN(pn)
	checkAddr(addr)
	if pn > 1 && len(addr) > 1 {
		panic("pn>1 && len(addr)>1")
	}
}

// RxAddr reads address assigned to Rx pipe pn into addr.
func (d Device) RxAddr(pn int, addr []byte) (stat Stat, err error) {
	checkPNA(pn, addr)
	return d.Reg(byte(0xa+pn), addr)
}

// SetRxAddr sets address assigned to Rx pipe pn to addr.
func (d Device) SetRxAddr(pn int, addr []byte) (Stat, error) {
	checkPNA(pn, addr)
	return d.SetReg(byte(0xa+pn), addr...)
}

// RxAddr0 returns least significant byte of address assigned to Rx pipe pn.
func (d Device) RxAddr0(pn int) (byte, Stat, error) {
	checkPN(pn)
	return d.byteReg(byte(0xa + pn))
}

// SetRxAddr0 sets least significant byte of address assigned to Rx pipe pn.
func (d Device) SetRxAddr0(pn int, a0 byte) (Stat, error) {
	checkPN(pn)
	return d.SetReg(byte(0xa+pn), a0)
}

// TxAddr returns value of TX_ADDR (Transmit address).
func (d Device) TxAddr(addr []byte) (Stat, error) {
	checkAddr(addr)
	return d.Reg(0x10, addr)
}

// SetTxAddr sets value of TX_ADDR (Transmit address).
func (d Device) SetTxAddr(addr []byte) (Stat, error) {
	checkAddr(addr)
	return d.SetReg(0x10, addr...)
}

// RxPW returns Rx payload width set for pipe pn.
func (d Device) RxPW(pn int) (int, Stat, error) {
	checkPN(pn)
	pw, stat, err := d.byteReg(byte(0x11 + pn))
	return int(pw & 0x3f), stat, err
}

// SetRxPW sets Rx payload width for pipe pn.
func (d Device) SetRxPW(pn, pw int) (Stat, error) {
	checkPN(pn)
	if uint(pw) > 32 {
		panic("pw<0 || pw>32")
	}
	return d.SetReg(byte(0x11+pn), byte(pw))
}

type FIFO byte

const (
	RxEmpty FIFO = 1 << iota // 1: Rx FIFO empty, 0: Data in Rx FIFO.
	RxFull                   // 1: Rx FIFO full, 0: Avail.locations in Rx FIFO.
	_
	_
	TxEmpty // 1: Tx FIFO empty, 0: Data in Tx FIFO.
	TxFull  // 1: Tx FIFO full, 0: Available locations in Tx FIFO.
	TxReuse // 1: Reuse last transmitted payload.
)

func (f FIFO) String() string {
	return flags("TxReuse+ TxFull+ TxEmpty+ RxFull+ RxEmpty+", 0x73, byte(f))
}

// FIFO returns value of FIFO_STATUS register.
func (d Device) FIFO() (FIFO, Stat, error) {
	fifo, stat, err := d.byteReg(0x17)
	return FIFO(fifo), stat, err
}

// DynPD returns value of DYNPD (Enable dynamic payload length) register.
func (d Device) DynPD() (Pipe, Stat, error) {
	p, stat, err := d.byteReg(0x1c)
	return Pipe(p), stat, err
}

// SetDynPD sets value of DYNPD (Enable dynamic payload length) register.
func (d Device) SetDynPD(p Pipe) (Stat, error) {
	return d.SetReg(0x1c, byte(p))
}

type Feature byte

const (
	DynAck Feature = 1 << iota // 1: Enables the W_TX_PAYLOAD_NOACK command.
	AckPay                     // 1: Enables payload with ACK
	DPL                        // 1: Enables dynamic payload length
)

func (f Feature) String() string {
	return flags("DPL+ AckPay+ DynAck+", 7, byte(f))
}

// Feature returns value of FEATURE register.
func (d Device) Feature() (Feature, Stat, error) {
	f, stat, err := d.byteReg(0x1d)
	return Feature(f), stat, err
}

// SetFeature sets value of FEATURE register.
func (d Device) SetFeature(f Feature) (Stat, error) {
	return d.SetReg(0x1d, byte(f))
}
