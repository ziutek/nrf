package nrf

// Reg invokes R_REGISTER command.
func (d *Device) Reg(addr byte, val []byte) {
	if d.Err != nil {
		return
	}
	var stat [1]byte
	_, d.Err = d.WriteRead([]byte{addr}, stat[:], nil, val)
	d.Status = Status(stat[0])
}

// SetReg invokes W_REGISTER command.
func (d *Device) SetReg(addr byte, val ...byte) {
	if d.Err != nil {
		return
	}
	var stat [1]byte
	_, d.Err = d.WriteRead([]byte{addr | 0x20}, stat[:], val)
	d.Status = Status(stat[0])
}

// ReadRxP invokes R_RX_PAYLOAD command.
func (d *Device) ReadRxP(pay []byte) {
	if d.Err != nil {
		return
	}
	var stat [1]byte
	_, d.Err = d.WriteRead([]byte{0x61}, stat[:], nil, pay)
	d.Status = Status(stat[0])
}

func checkPlen(plen int) {
	if plen > 32 {
		panic("plen>32")
	}
}

// WriteTxP invokes W_TX_PAYLOAD command.
func (d *Device) WriteTxP(pay []byte) {
	checkPlen(len(pay))
	if d.Err != nil {
		return
	}
	var stat [1]byte
	_, d.Err = d.WriteRead([]byte{0xa0}, stat[:], pay)
	d.Status = Status(stat[0])
}

// FlushTx invokes FLUSH_TX command.
func (d *Device) FlushTx() {
	if d.Err != nil {
		return
	}
	var stat [1]byte
	_, d.Err = d.WriteRead([]byte{0xe1}, stat[:])
	d.Status = Status(stat[0])
}

// FlushRx invokes FLUSH_RX command.
func (d *Device) FlushRx() {
	if d.Err != nil {
		return
	}
	var stat [1]byte
	_, d.Err = d.WriteRead([]byte{0xe2}, stat[:])
	d.Status = Status(stat[0])
}

// ReuseTxP invokes REUSE_TX_PL command.
func (d *Device) ReuseTxP() {
	if d.Err != nil {
		return
	}
	var stat [1]byte
	_, d.Err = d.WriteRead([]byte{0xe3}, stat[:])
	d.Status = Status(stat[0])
}

// Activate invokes nRF24L01 ACTIVATE command.
func (d *Device) Activate(b byte) {
	if d.Err != nil {
		return
	}
	var stat [1]byte
	_, d.Err = d.WriteRead([]byte{0x50, b}, stat[:])
	d.Status = Status(stat[0])
}

// RxPLen invokes R_RX_PL_WID command.
func (d *Device) RxPLen() int {
	if d.Err != nil {
		return 0
	}
	var ret [2]byte
	_, d.Err = d.WriteRead([]byte{0x60}, ret[:])
	d.Status = Status(ret[0])
	return int(ret[1])
}

// WriteAckP invokes W_ACK_PAYLOAD command.
func (d *Device) WriteAckP(pn int, pay []byte) {
	checkPN(pn)
	checkPlen(len(pay))
	if d.Err != nil {
		return
	}
	var stat [1]byte
	_, d.Err = d.WriteRead([]byte{byte(0xa8 | pn)}, stat[:], pay)
	d.Status = Status(stat[0])
}

// WriteTxPNoAck invokes W_TX_PAYLOAD_NOACK command.
func (d *Device) WriteTxPNoAck(pay []byte) {
	checkPlen(len(pay))
	if d.Err != nil {
		return
	}
	var stat [1]byte
	_, d.Err = d.WriteRead([]byte{0xa0}, stat[:], pay)
	d.Status = Status(stat[0])
}

// NOP invokes NOP command.
func (d *Device) NOP() {
	if d.Err != nil {
		return
	}
	var stat [1]byte
	_, d.Err = d.WriteRead([]byte{0xff}, stat[:])
	d.Status = Status(stat[0])
}
