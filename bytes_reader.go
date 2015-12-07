package irsdk

import (
	"bytes"
	"encoding/binary"
	"unsafe"
)

type BytesReader struct {
	data        []byte
	header      *Header
	subHeader   *DiskSubHeader
	sessionData *SessionData

	varHeaders []*VarHeader
	dataPoints []*TelemetryData
}

// GetHeader memoizes the ReadHeader function
func (br *BytesReader) GetHeader() (*Header, error) {
	var err error

	if br.header == nil {
		br.header, err = br.ReadHeader()
		return br.header, err
	}

	return br.header, err
}

// ReadHeader tries to read the main header from the file
func (br *BytesReader) ReadHeader() (*Header, error) {
	b, err := br.GetLocationBytes()
	if err != nil {
		return nil, err
	}

	// Create a pointer to the bytes
	bPtr := uintptr(unsafe.Pointer(&b[0]))

	// Point header struct to the location of the bytes
	header := (*Header)(unsafe.Pointer(bPtr))
	return header, nil
}

func (br *BytesReader) GetLocationBytes() ([]byte, error) {
	start, size, err := br.GetHeaderLocation()
	if err != nil {
		return nil, err
	}

	end := start + size
	return br.data[start:end], nil
}

func (br *BytesReader) GetHeaderLocation() (int, int, error) {
	startByte := 0

	header := &Header{}
	// Create byte slice big enough for header data
	size := binary.Size(header)
	return startByte, size, nil
}

// GetHeader memoizes the ReadSubHeader function
func (br *BytesReader) GetsubHeader() (*DiskSubHeader, error) {
	var err error

	if br.subHeader == nil {
		br.subHeader, err = br.ReadSubHeader()
		return br.subHeader, err
	}

	return br.subHeader, err
}

// ReadSubHeader reads the second header specialiy for telemtry data saved to
// .ibt files
func (br *BytesReader) ReadSubHeader() (*DiskSubHeader, error) {
	b, err := br.GetSubHeaderBytes()
	if err != nil {
		return nil, err
	}

	// Create a pointer to the bytes
	bPtr := uintptr(unsafe.Pointer(&b[0]))

	// Point subHeader struct to the location of the bytes
	subHeader := (*DiskSubHeader)(unsafe.Pointer(bPtr))
	return subHeader, nil
}

func (br *BytesReader) GetSubHeaderBytes() ([]byte, error) {
	start, size, err := br.GetSubHeaderLocation()
	if err != nil {
		return nil, err
	}

	end := start + size
	return br.data[start:end], nil
}

func (br *BytesReader) GetSubHeaderLocation() (int, int, error) {
	header, err := br.GetHeader()
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
func (br *BytesReader) GetSessionData() (*SessionData, error) {
	var err error

	if br.sessionData == nil {
		br.sessionData, err = br.ReadSessionData()
		return br.sessionData, err
	}

	return br.sessionData, err
}

func (br *BytesReader) ReadSessionData() (*SessionData, error) {
	b, err := br.GetSessionDataBytes()
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

func (br *BytesReader) GetSessionDataBytes() ([]byte, error) {
	start, size, err := br.GetSessionDataLocation()
	if err != nil {
		return nil, err
	}

	end := start + size
	return br.data[start:end], nil
}

func (br *BytesReader) GetSessionDataLocation() (int, int, error) {
	header, err := br.GetHeader()
	if err != nil {
		return 0, 0, err
	}

	// Copied from GetSessionInfoStr()
	start := int(header.SessionInfoOffset)
	size := int(header.SessionInfoLen)
	return start, size, nil
}

// GetVarHeaders memoizes the ReadHeader function
func (br *BytesReader) GetVarHeaders() ([]*VarHeader, error) {
	var err error

	if br.varHeaders == nil {
		br.varHeaders, err = br.ReadVarHeaders()
		return br.varHeaders, err
	}

	return br.varHeaders, err
}

// ReadVarHeaders reads the second header specialiy for telemtry data saved to
// .ibt files
func (br *BytesReader) ReadVarHeaders() ([]*VarHeader, error) {
	header, err := br.GetHeader()
	if err != nil {
		return nil, err
	}

	numVars := int(header.NumVars)
	varHeaders := make([]*VarHeader, numVars)

	for i := 0; i < numVars; i++ {
		vh, err := br.ReadVarHeaderEntry(i)
		if err != nil {
			return nil, err
		}
		varHeaders[i] = vh
	}

	return varHeaders, nil
}

func (br *BytesReader) ReadVarHeaderEntry(i int) (*VarHeader, error) {
	b, err := br.GetVarHeaderEntryBytes(i)
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

func (br *BytesReader) GetVarHeaderEntryBytes(i int) ([]byte, error) {
	start, size, err := br.GetVarHeaderEntryLocation(i)
	if err != nil {
		return nil, err
	}

	end := start + size
	return br.data[start:end], nil
}

func (br *BytesReader) GetVarHeaderEntryLocation(i int) (int, int, error) {
	header, err := br.GetHeader()
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
func (br *BytesReader) GetAllDataPoints() ([]*TelemetryData, error) {
	var err error

	if br.dataPoints == nil {
		br.dataPoints, err = br.ReadAllDataPoints()
		return br.dataPoints, err
	}

	return br.dataPoints, err
}

// ReadAllDataPoints reads all datapoints (TelemetrydDat) from the file / memmap
func (br *BytesReader) ReadAllDataPoints() ([]*TelemetryData, error) {
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
		td, err := br.ReadDataPointN(i)
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
func (br *BytesReader) ReadDataPointN(i int) (*TelemetryData, error) {
	b, err := br.GetDataPointEntryBytes(i)
	if err != nil {
		return nil, err
	}
	// EOF probably reached
	if len(b) == 0 {
		return nil, nil
	}

	// Collect varHeaders
	varHeaders, err := br.GetVarHeaders()
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

func (br *BytesReader) GetDataPointEntryBytes(i int) ([]byte, error) {
	start, size, err := br.GetLocationDataPointEntryLocation(i)
	if err != nil {
		return nil, err
	}

	end := start + size

	// check if byteslice contains enough bytes
	if end > len(br.data) {
		return nil, nil
	}

	return br.data[start:end], nil
}

func (br *BytesReader) GetLocationDataPointEntryLocation(i int) (int, int, error) {
	header, err := br.GetHeader()
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
func NewTelemetryReader(data []byte) *BytesReader {
	return &BytesReader{
		data: data,
	}
}
