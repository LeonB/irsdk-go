package irsdk

import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

type TelemetryReader struct {
	data        []byte
	header      *Header
	subHeader   *DiskSubHeader
	sessionData *SessionData

	varHeaders []*VarHeader
	dataPoints []*TelemetryData
}

// GetHeader memoizes the ReadHeader function
func (tr *TelemetryReader) GetHeader() (*Header, error) {
	var err error

	if tr.header == nil {
		tr.header, err = tr.ReadHeader()
		return tr.header, err
	}

	return tr.header, err
}

// ReadHeader tries to read the main header from the file
func (tr *TelemetryReader) ReadHeader() (*Header, error) {
	b, err := tr.GetLocationBytes()
	if err != nil {
		return nil, err
	}

	// Create a pointer to the bytes
	bPtr := uintptr(unsafe.Pointer(&b[0]))

	// Point header struct to the location of the bytes
	header := (*Header)(unsafe.Pointer(bPtr))
	return header, nil
}

func (tr *TelemetryReader) GetLocationBytes() ([]byte, error) {
	start, size, err := tr.GetHeaderLocation()
	if err != nil {
		return nil, err
	}

	end := start + size
	return tr.data[start:end], nil
}

func (tr *TelemetryReader) GetHeaderLocation() (int, int, error) {
	startByte := 0

	header := &Header{}
	// Create byte slice big enough for header data
	size := binary.Size(header)
	return startByte, size, nil
}

// GetHeader memoizes the ReadSubHeader function
func (tr *TelemetryReader) GetsubHeader() (*DiskSubHeader, error) {
	var err error

	if tr.subHeader == nil {
		tr.subHeader, err = tr.ReadSubHeader()
		return tr.subHeader, err
	}

	return tr.subHeader, err
}

// ReadSubHeader reads the second header specialiy for telemtry data saved to
// .ibt files
func (tr *TelemetryReader) ReadSubHeader() (*DiskSubHeader, error) {
	b, err := tr.GetSubHeaderBytes()
	if err != nil {
		return nil, err
	}

	// Create a pointer to the bytes
	bPtr := uintptr(unsafe.Pointer(&b[0]))

	// Point subHeader struct to the location of the bytes
	subHeader := (*DiskSubHeader)(unsafe.Pointer(bPtr))
	return subHeader, nil
}

func (tr *TelemetryReader) GetSubHeaderBytes() ([]byte, error) {
	start, size, err := tr.GetSubHeaderLocation()
	if err != nil {
		return nil, err
	}

	end := start + size
	return tr.data[start:end], nil
}

func (tr *TelemetryReader) GetSubHeaderLocation() (int, int, error) {
	header, err := tr.GetHeader()
	if err != nil {
		return 0, 0, err
	}

	subHeader := &DiskSubHeader{}
	headerSize := binary.Size(header)
	subHeaderSize := binary.Size(subHeader)
	start := headerSize
	return start, subHeaderSize, nil
}

// GetSessionData memoizes the ReadSubHeader function
func (tr *TelemetryReader) GetSessionData() (*SessionData, error) {
	var err error

	if tr.sessionData == nil {
		tr.sessionData, err = tr.ReadSessionData()
		return tr.sessionData, err
	}

	return tr.sessionData, err
}

func (tr *TelemetryReader) ReadSessionData() (*SessionData, error) {
	b, err := tr.GetSessionDataBytes()
	if err != nil {
		return nil, err
	}

	// Copied from GetRawSessionData()
	sep := []byte("\n...")
	pieces := bytes.Split(b, sep)
	if len(pieces) == 0 {
		return nil, ErrEmptySessionData
	}

	if pieces[0] == nil {
		return nil, ErrEmptySessionData
	}

	b = bytesToUtf8(pieces[0])
	return NewSessionDataFromBytes(b)
}

func (tr *TelemetryReader) GetSessionDataBytes() ([]byte, error) {
	start, size, err := tr.GetSessionDataLocation()
	if err != nil {
		return nil, err
	}

	end := start + size
	return tr.data[start:end], nil
}

func (tr *TelemetryReader) GetSessionDataLocation() (int, int, error) {
	header, err := tr.GetHeader()
	if err != nil {
		return 0, 0, err
	}

	// Copied from GetSessionInfoStr()
	start := int(header.SessionInfoOffset)
	size := int(header.SessionInfoLen)
	return start, size, nil
}

// GetVarHeaders memoizes the ReadHeader function
func (tr *TelemetryReader) GetVarHeaders() ([]*VarHeader, error) {
	var err error

	if tr.varHeaders == nil {
		tr.varHeaders, err = tr.ReadVarHeaders()
		return tr.varHeaders, err
	}

	return tr.varHeaders, err
}

// ReadVarHeaders reads the second header specialiy for telemtry data saved to
// .ibt files
func (tr *TelemetryReader) ReadVarHeaders() ([]*VarHeader, error) {
	header, err := tr.GetHeader()
	if err != nil {
		return nil, err
	}

	numVars := int(header.NumVars)
	varHeaders := make([]*VarHeader, numVars)

	for i := 0; i < numVars; i++ {
		vh, err := tr.ReadVarHeaderEntry(i)
		if err != nil {
			return nil, err
		}
		varHeaders[i] = vh
	}

	return varHeaders, nil
}

func (tr *TelemetryReader) ReadVarHeaderEntry(i int) (*VarHeader, error) {
	b, err := tr.GetVarHeaderEntryBytes(i)
	if err != nil {
		return nil, err
	}

	// Create a pointer to the bytes
	bPtr := uintptr(unsafe.Pointer(&b[0]))

	// Point varHeader struct to the location of the bytes
	varHeaderRaw := (*VarHeaderRaw)(unsafe.Pointer(bPtr))

	varHeader := &VarHeader{
		Type:   varHeaderRaw.Type,
		Offset: varHeaderRaw.Offset,
		Count:  varHeaderRaw.Count,

		Name: CToGoString(varHeaderRaw.Name[:]),
		Desc: CToGoString(varHeaderRaw.Desc[:]),
		Unit: CToGoString(varHeaderRaw.Unit[:]),
	}

	return varHeader, nil
}

func (tr *TelemetryReader) GetVarHeaderEntryBytes(i int) ([]byte, error) {
	start, size, err := tr.GetVarHeaderEntryLocation(i)
	if err != nil {
		return nil, err
	}

	end := start + size
	return tr.data[start:end], nil
}

func (tr *TelemetryReader) GetVarHeaderEntryLocation(i int) (int, int, error) {
	header, err := tr.GetHeader()
	if err != nil {
		return 0, 0, err
	}

	varHeaderRaw := &VarHeaderRaw{}
	startVarHeaders := int(header.VarHeaderOffset)
	size := int(binary.Size(varHeaderRaw))

	// Get offset to jump to
	offset := i * size
	start := startVarHeaders + offset

	return start, size, nil
}

// GetVarBufs memoizes the ReadHeader function
func (tr *TelemetryReader) GetAllDataPoints() ([]*TelemetryData, error) {
	var err error

	if tr.dataPoints == nil {
		tr.dataPoints, err = tr.ReadAllDataPoints()
		return tr.dataPoints, err
	}

	return tr.dataPoints, err
}

// ReadAllDataPoints reads all datapoints (TelemetrydDat) from the file / memmap
func (tr *TelemetryReader) ReadAllDataPoints() ([]*TelemetryData, error) {
	// diskSubHeader.sessionRecordCount should have the total number of
	// telemetrydatapoints
	//
	// David Tucker wrote: Hmm, maybe this has a clue, it looks like there is an
	// additional sessionRecordCount parameter tacked onto the YAML session
	// string when writing out the .ibt file. My code seems to blow away the sub
	// header sessionRecordCount and always uses the YAML version...
	sessionRecordCount := 0
	datapoints := make([]*TelemetryData, sessionRecordCount)

	i := 0
	for {
		td, err := tr.ReadDataPointN(i)
		if err != nil {
			return nil, err
		}
		if td == nil {
			// End of data: get out of while-loop
			break
		}

		datapoints = append(datapoints, td)
		i = i + 1
	}

	return datapoints, nil
}

// ReadDataPointN reads a specific datapoint (TelemetryData)
func (tr *TelemetryReader) ReadDataPointN(i int) (*TelemetryData, error) {
	b, err := tr.GetDataPointEntryBytes(i)
	if err != nil {
		return nil, err
	}
	// EOF probably reached
	if len(b) == 0 {
		return nil, nil
	}

	// Collect varHeaders
	varHeaders, err := tr.GetVarHeaders()
	if err != nil {
		return nil, err
	}

	// Create new datapoint
	td := NewTelemetryData()

	// Loop through all headers and parse data/bytes
	for _, varHeader := range varHeaders {
		err = td.addVarHeaderData(varHeader, b)
		if err != nil {
			return nil, err
		}
	}

	return td, nil
}

func (tr *TelemetryReader) GetDataPointEntryBytes(i int) ([]byte, error) {
	start, size, err := tr.GetLocationDataPointEntryLocation(i)
	if err != nil {
		return nil, err
	}

	end := start + size

	// check if byteslice contains enough bytes
	if end > len(tr.data) {
		return nil, nil
	}

	return tr.data[start:end], nil
}

func (tr *TelemetryReader) GetLocationDataPointEntryLocation(i int) (int, int, error) {
	header, err := tr.GetHeader()
	if err != nil {
		return 0, 0, err
	}

	latest := header.GetLatestVarBufN()
	varBufStart := int(header.VarBuf[latest].BufOffset)
	varBufSize := int(header.BufLen)

	// Get offset to jump to
	offset := i * varBufSize
	start := varBufStart + offset

	return start, varBufSize, nil
}

// NewTelemetryReader intializes a new TelemetryReader object based on an
// io.ReadSeeker
func NewTelemetryReader(data []byte) *TelemetryReader {
	return &TelemetryReader{
		data: data,
	}
}
