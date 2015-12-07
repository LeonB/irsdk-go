package irsdk

// Kan gecombineerd worden met disktelemetry
func (cw *CWrapper) getHeader() (*Header, error) {
	tr := NewTelemetryReader(cw.sharedMem)
	return tr.GetHeader()
}

func (cw *CWrapper) getVarHeaderEntry(index int) (*VarHeader, error) {
	tr := NewTelemetryReader(cw.sharedMem)
	return tr.ReadVarHeaderEntry(index)
}
