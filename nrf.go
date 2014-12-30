package nrf

import (
	"fmt"
)

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

func flag(s Status) byte {
	if s == 0 {
		return '-'
	}
	return '+'
}

func (s Status) String() string {
	return fmt.Sprintf(
		"TxFull%c RxPipe:%d MaxRT%c TxDS%c RxDR%c",
		flag(s&TxFull), s.RxPipe(), flag(s&MaxRT), flag(s&TxDS), flag(s&RxDR),
	)
}

// Reg invokes R_REGISTER command. It reads status byte and content of register
// r into provided buffer buf. Before call buf[0] should contain register
// address. After return buf[0] contains status byte, buf[1:] contains data
// read. This is general command. Use specific Reg* commands instead.
func (d Device) Reg(buf []byte) error {
	buf[0] &= 0x1f
	_, err := d.WriteRead(buf[:1], buf)
	return err	
}

// Config returns content of STATUS and CONFIG register.
func (d Device) Config() (Status, byte, error) {
	buf := [2]byte{}
	err := d.Reg(buf[:])
	return Status(buf[0]), buf[1], err
}

// NOP invokes NOP command.
func (d Device) NOP() (Status, error) {
	buf := []byte{0xff}
	_, err := d.WriteRead(buf, buf)
	return Status(buf[0]), err
}
