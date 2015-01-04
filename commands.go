package nrf

// Reg invokes R_REGISTER command.
func (d Device) Reg(addr byte, val []byte) (Stat, error) {
	cmd := []byte{addr}
	_, err := d.WriteRead(cmd, cmd, nil, val)
	return Stat(cmd[0]), err
}

// SetReg invokes W_REGISTER command.
func (d Device) SetReg(addr byte, val ...byte) (Stat, error) {
	cmd := []byte{addr | 0x20}
	_, err := d.WriteRead(cmd, cmd, val)
	return Stat(cmd[0]), err
}

func checkPlen(plen int) {
	if plen > 32 {
		panic("plen>32")
	}
}

// ReadRxP invokes R_RX_PAYLOAD command.
func (d Device) ReadRxP(pay []byte) (Stat, error) {
	cmd := []byte{0x61}
	_, err := d.WriteRead(cmd, cmd, nil, pay)
	return Stat(cmd[0]), err
}

// WriteTxP invokes W_TX_PAYLOAD command.
func (d Device) WriteTxP(pay []byte) (Stat, error) {
	checkPlen(len(pay))
	cmd := []byte{0xa0}
	_, err := d.WriteRead(cmd, cmd, pay)
	return Stat(cmd[0]), err
}

// FlushTx invokes FLUSH_TX command.
func (d Device) FlushTx() (Stat, error) {
	cmd := []byte{0xe1}
	_, err := d.WriteRead(cmd, cmd)
	return Stat(cmd[0]), err
}

// FlushRx invokes FLUSH_RX command.
func (d Device) FlushRx() (Stat, error) {
	cmd := []byte{0xe2}
	_, err := d.WriteRead(cmd, cmd)
	return Stat(cmd[0]), err
}

// ReuseTxP invokes REUSE_TX_PL command.
func (d Device) ReuseTxP() (Stat, error) {
	cmd := []byte{0xe3}
	_, err := d.WriteRead(cmd, cmd)
	return Stat(cmd[0]), err
}

// Activate invokes nRF24L01 ACTIVATE command.
func (d Device) Activate(b byte) (Stat, error) {
	cmd := []byte{0x50, b}
	_, err := d.WriteRead(cmd, cmd[:1])
	return Stat(cmd[0]), err
}

// RxPLen invokes R_RX_PL_WID command.
func (d Device) RxPLen() (int, Stat, error) {
	cmd := []byte{0x60, 0}
	_, err := d.WriteRead(cmd[:1], cmd)
	return int(cmd[1]), Stat(cmd[0]), err
}

// WriteAckP invokes W_ACK_PAYLOAD command.
func (d Device) WriteAckP(pn int, pay []byte) (Stat, error) {
	checkPN(pn)
	checkPlen(len(pay))
	cmd := []byte{byte(0xa8 | pn)}
	_, err := d.WriteRead(cmd, cmd, pay)
	return Stat(cmd[0]), err
}

// WriteTxPNoAck invokes W_TX_PAYLOAD_NOACK command.
func (d Device) WriteTxPNoAck(pay []byte) (Stat, error) {
	checkPlen(len(pay))
	cmd := []byte{0xa0}
	_, err := d.WriteRead(cmd, cmd, pay)
	return Stat(cmd[0]), err
}

// NOP invokes NOP command.
func (d Device) NOP() (Stat, error) {
	cmd := []byte{0xff}
	_, err := d.WriteRead(cmd, cmd)
	return Stat(cmd[0]), err
}
