package nrf

import "strconv"

type Driver interface {
	// Perform SPI conversation.
	WriteRead(out, in []byte) (int, error)
	// Set CE line.
	Enable(bool) error
}

type Device struct {
	Driver
}

type Status byte

const (
	TxFull Status = 1 << iota // Tx FIFO full flag.
	_
	_
	_
	MaxRT // Maximum number of Tx retransmits interrupt.
	TxDS  // Data Sent Tx FIFO interrupt.
	RxDR  // Data Ready Rx FIFO interrupt.
)

// RxPipe returns data pipe number for the payload available for reading from
// RxFifo or -1 if RxFifo is empty
func (s Status) RxPipe() int {
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

func (s Status) String() string {
	return flags("RxDR+ TxDS+ MaxRT+ TxFull+ RxPipe:", 0x71, byte(s)) +
		strconv.Itoa(s.RxPipe())
}

// Reg invokes R_REGISTER command. buf[0] should contain register address.
// After return buf[0] contains status byte, buf[1:] contains data
// read.
func (d Device) Reg(buf []byte) error {
	_, err := d.WriteRead(buf[:1], buf)
	return err
}

// SetReg invokes W_REGISTER command. buf[0] should contain register address,
// buf[1:] bytes to write into register.
func (d Device) SetReg(buf []byte) (Status, error) {
	var status [1]byte
	buf[0] |= 0x20
	_, err := d.WriteRead(buf, status[:])
	return Status(status[0]), err
}

func (d Device) byteReg(addr byte) (byte, Status, error) {
	buf := [2]byte{addr, 0}
	err := d.Reg(buf[:])
	return buf[1], Status(buf[0]), err
}

type Config byte

const (
	PrimRx Config = 1 << iota
	PwrUp
	CRCO
	EnCRC
	MaskMaxRT
	MaskTxDS
	MaskRxDR
)

func (c Config) String() string {
	return flags(
		"Mask(RxDR+ TxDS+ MaxRT+) EnCRC+ CRCO+ PwrUp+ PrimRx+",
		0x7f, byte(c),
	)
}

// Config returns value of CONFIG register.
func (d Device) Config() (Config, Status, error) {
	cfg, sta, err := d.byteReg(0)
	return Config(cfg), sta, err
}

// SetConfig sets value of CONFIG register to c.
func (d Device) SetConfig(c Config) (Status, error) {
	buf := [2]byte{0, byte(c)}
	return d.SetReg(buf[:])
}

type Pipe byte

const (
	P0 Pipe = 1 << iota
	P1
	P2
	P3
	P4
	P5
)

func (p Pipe) String() string {
	return flags("P5+ P4+ P3+ P2+ P1+ P0+", 0x3f, byte(p))
}

func (d Device) AA() (Pipe, Status, error) {
	pipe, sta, err := d.byteReg(1)
	return Pipe(pipe), sta, err
}

func (d Device) SetAA(aa Pipe) (Status, error) {
	buf := [2]byte{1, byte(aa)}
	return d.SetReg(buf[:])
}

func (d Device) RxAddrEn() (Pipe, Status, error) {
	pipe, sta, err := d.byteReg(2)
	return Pipe(pipe), sta, err
}

func (d Device) SetRxAddrEn(ra Pipe) (Status, error) {
	buf := [2]byte{2, byte(ra)}
	return d.SetReg(buf[:])
}

func (d Device) AW() (int, Status, error) {
	aw, sta, err := d.byteReg(3)
	return int(aw), sta, err
}

func (d Device) SetAW(aw int) (Status, error) {
	if aw < 3 || aw > 5 {
		panic("aw<3 || aw>5")
	}
	buf := [2]byte{3, byte(aw - 2)}
	return d.SetReg(buf[:])
}

func (d Device) Retr() (cnt, dlyus int, sta Status, err error) {
	b, sta, err := d.byteReg(4)
	cnt = int(b & 0xf)
	dlyus = (int(b>>4) + 1) * 250
	return
}

func (d Device) SetRetr(cnt, dlyus int) (Status, error) {
	if cnt < 0 || cnt > 15 {
		panic("cnt<0 || cnt>15")
	}
	if dlyus < 250 || dlyus > 4000 {
		panic("dlyus<250 || dlyus>4000")
	}
	buf := [2]byte{4, byte((dlyus/250-1)<<4 | cnt)}
	return d.SetReg(buf[:])
}

// NOP invokes NOP command.
func (d Device) NOP() (Status, error) {
	buf := []byte{0xff}
	_, err := d.WriteRead(buf, buf)
	return Status(buf[0]), err
}
