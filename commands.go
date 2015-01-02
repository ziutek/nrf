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
	_, err := d.WriteRead(cmd, cmd, val, nil)
	return Stat(cmd[0]), err
}

// ReadRxP invokes R_RX_PAYLOAD comand.
/*func (d Device) ReadRxP(pay []byte) (Stat, error) {
	buf := [6]byte{0: 0x61}
}*/

// NOP invokes NOP command.
func (d Device) NOP() (Stat, error) {
	cmd := []byte{0xff}
	_, err := d.WriteRead(cmd, cmd)
	return Stat(cmd[0]), err
}
