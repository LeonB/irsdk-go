package irsdk

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/leonb/irsdk-go/utils"
)

type TelemetryReader struct {
	data        io.ReadSeeker
	header      *utils.Header
	subHeader   *utils.DiskSubHeader
	sessionData *SessionData
	varHeaders  []*utils.VarHeader
	dataPoints  []*TelemetryData
}

// GetHeader memoizes the ReadHeader function
func (tr *TelemetryReader) GetHeader() (*utils.Header, error) {
	var err error

	if tr.header == nil {
		tr.header, err = tr.ReadHeader()
		return tr.header, err
	}

	return tr.header, err
}

// ReadHeader tries to read the main header from the file
func (tr *TelemetryReader) ReadHeader() (*utils.Header, error) {
	header := &utils.Header{}

	// Create byte slice big enough for header data
	size := binary.Size(header)
	b := make([]byte, size)

	// Jump to right position in file
	startByte := 0
	tr.data.Seek(int64(startByte), 0) // 0 = relative to the origin of the file

	// Read data from file into byteslice
	_, err := tr.data.Read(b)
	if err != nil {
		return nil, err
	}

	// Create a pointer to the bytes
	bPtr := uintptr(unsafe.Pointer(&b[0]))

	// Point header struct to the location of the bytes
	header = (*utils.Header)(unsafe.Pointer(bPtr))
	return header, nil
}

// GetHeader memoizes the ReadSubHeader function
func (tr *TelemetryReader) GetsubHeader() (*utils.DiskSubHeader, error) {
	var err error

	if tr.subHeader == nil {
		tr.subHeader, err = tr.ReadSubHeader()
		return tr.subHeader, err
	}

	return tr.subHeader, err
}

// ReadSubHeader reads the second header specialiy for telemtry data saved to
// .ibt files
func (tr *TelemetryReader) ReadSubHeader() (*utils.DiskSubHeader, error) {
	header, err := tr.GetHeader()
	if err != nil {
		return nil, err
	}

	subHeader := &utils.DiskSubHeader{}
	headerSize := binary.Size(header)
	subHeaderSize := binary.Size(subHeader)
	startByte := headerSize

	// Create byte slice big enough for subHeader data
	b := make([]byte, subHeaderSize)

	// Jump to right position in file
	tr.data.Seek(int64(startByte), 0) // 0 = relative to the origin of the file

	// Read data from file into byteslice
	_, err = tr.data.Read(b)
	if err != nil {
		return nil, err
	}

	// Create a pointer to the bytes
	bPtr := uintptr(unsafe.Pointer(&b[0]))

	// Point subHeader struct to the location of the bytes
	subHeader = (*utils.DiskSubHeader)(unsafe.Pointer(bPtr))
	return subHeader, nil
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
	header, err := tr.GetHeader()
	if err != nil {
		return nil, err
	}

	// Copied from GetSessionInfoStr()
	startByte := header.SessionInfoOffset
	sessionSize := header.SessionInfoLen

	// Create byte slice big enough for subHeader data
	b := make([]byte, sessionSize)

	// Jump to right position in file
	tr.data.Seek(int64(startByte), 0) // 0 = relative to the origin of the file

	// Read data from file into byteslice
	_, err = tr.data.Read(b)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, nil
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
	fmt.Printf("%v\n", b)
	return NewSessionDataFromBytes(b)
}

// GetVarHeaders memoizes the ReadHeader function
func (tr *TelemetryReader) GetVarHeaders() ([]*utils.VarHeader, error) {
	var err error

	if tr.varHeaders == nil {
		tr.varHeaders, err = tr.ReadVarHeaders()
		return tr.varHeaders, err
	}

	return tr.varHeaders, err
}

// ReadVarHeaders reads the second header specialiy for telemtry data saved to
// .ibt files
func (tr *TelemetryReader) ReadVarHeaders() ([]*utils.VarHeader, error) {
	header, err := tr.GetHeader()
	if err != nil {
		return nil, err
	}

	varHeader := &utils.VarHeader{}
	startByte := int64(header.VarHeaderOffset)
	varHeaderSize := binary.Size(varHeader)
	numVars := int(header.NumVars)
	varHeaders := make([]*utils.VarHeader, numVars)

	// Jump to right position in file
	tr.data.Seek(startByte, 0) // 0 = relative to the origin of the file

	for i := 0; i < numVars; i++ {
		// Create byte slice big enough for varHeader data
		b := make([]byte, varHeaderSize)

		// Read data from file into byteslice
		_, err = tr.data.Read(b)
		if err != nil {
			return nil, err
		}

		// Create a pointer to the bytes
		bPtr := uintptr(unsafe.Pointer(&b[0]))

		// Point varHeader struct to the location of the bytes
		vh := (*utils.VarHeader)(unsafe.Pointer(bPtr))
		varHeaders[i] = vh
	}

	return varHeaders, nil
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
	header, err := tr.GetHeader()
	if err != nil {
		return nil, err
	}

	latest := header.GetLatestVarBufN()
	startByte := int(header.VarBuf[latest].BufOffset)
	varBufSize := int(header.BufLen)

	// Create byte slice big enough for telemetrydata
	b := make([]byte, varBufSize)

	// Get offset to jump to
	offset := i * varBufSize

	// Jump to right position in file
	tr.data.Seek(int64(startByte+offset), 0) // 0 = relative to the origin of the file

	// // Read data from file into byteslice
	n, err := tr.data.Read(b)
	if err != nil && err != io.EOF {
		return nil, err
	}
	if n <= 0 {
		return nil, nil
	}

	// Create new datapoint
	td := NewTelemetryData()

	// Loop through all headers and parse data/bytes
	varHeaders, _ := tr.GetVarHeaders()
	for _, varHeader := range varHeaders {
		err = td.addVarHeaderData(varHeader, b)
		if err != nil {
			return nil, err
		}
	}

	return td, nil
}

// NewTelemetryReader intializes a new TelemetryReader object based on an
// io.ReadSeeker
func NewTelemetryReader(data io.ReadSeeker) *TelemetryReader {
	return &TelemetryReader{
		data: data,
	}
}
