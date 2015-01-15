package nrf

// Reg invokes R_REGISTER command.
func (d Device) Reg(addr byte, val []byte) (Stat, error) {
	var stat [1]byte
	_, err := d.WriteRead([]byte{addr}, stat[:], nil, val)
	return Stat(stat[0]), err
}

// SetReg invokes W_REGISTER command.
func (d Device) SetReg(addr byte, val ...byte) (Stat, error) {
	var stat [1]byte
	_, err := d.WriteRead([]byte{addr | 0x20}, stat[:], val)
	return Stat(stat[0]), err
}

func checkPlen(plen int) {
	if plen > 32 {
		panic("plen>32")
	}
}

// ReadRxP invokes R_RX_PAYLOAD command.
func (d Device) ReadRxP(pay []byte) (Stat, error) {
	var stat [1]byte
	_, err := d.WriteRead([]byte{0x61}, stat[:], nil, pay)
	return Stat(stat[0]), err
}

// WriteTxP invokes W_TX_PAYLOAD command.
func (d Device) WriteTxP(pay []byte) (Stat, error) {
	checkPlen(len(pay))
	var stat [1]byte
	_, err := d.WriteRead([]byte{0xa0}, stat[:], pay)
	return Stat(stat[0]), err
}

// FlushTx invokes FLUSH_TX command.
func (d Device) FlushTx() (Stat, error) {
	var stat [1]byte
	_, err := d.WriteRead([]byte{0xe1}, stat[:])
	return Stat(stat[0]), err
}

// FlushRx invokes FLUSH_RX command.
func (d Device) FlushRx() (Stat, error) {
	var stat [1]byte
	_, err := d.WriteRead([]byte{0xe2}, stat[:])
	return Stat(stat[0]), err
}

// ReuseTxP invokes REUSE_TX_PL command.
func (d Device) ReuseTxP() (Stat, error) {
	var stat [1]byte
	_, err := d.WriteRead([]byte{0xe3}, stat[:])
	return Stat(stat[0]), err
}

// Activate invokes nRF24L01 ACTIVATE command.
func (d Device) Activate(b byte) (Stat, error) {
	var stat [1]byte
	_, err := d.WriteRead([]byte{0x50, b}, stat[:])
	return Stat(stat[0]), err
}

// RxPLen invokes R_RX_PL_WID command.
func (d Device) RxPLen() (int, Stat, error) {
	var ret [2]byte
	_, err := d.WriteRead([]byte{0x60}, ret[:])
	return int(ret[1]), Stat(ret[0]), err
}

// WriteAckP invokes W_ACK_PAYLOAD command.
func (d Device) WriteAckP(pn int, pay []byte) (Stat, error) {
	checkPN(pn)
	checkPlen(len(pay))
	var stat [1]byte
	_, err := d.WriteRead([]byte{byte(0xa8 | pn)}, stat[:], pay)
	return Stat(stat[0]), err
}

// WriteTxPNoAck invokes W_TX_PAYLOAD_NOACK command.
func (d Device) WriteTxPNoAck(pay []byte) (Stat, error) {
	checkPlen(len(pay))
	var stat [1]byte
	_, err := d.WriteRead([]byte{0xa0}, stat[:], pay)
	return Stat(stat[0]), err
}

// NOP invokes NOP command.
func (d Device) NOP() (Stat, error) {
	var stat [1]byte
	_, err := d.WriteRead([]byte{0xff}, stat[:])
	return Stat(stat[0]), err
}
