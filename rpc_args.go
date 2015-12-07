package irsdk

type OpenFileMappingArgs struct {
	LpName string
}

type MapViewOfFileArgs struct {
	HMemMapFile          uintptr
	DwNumberOfBytesToMap int
}

type OpenEventArgs struct {
	LpName string
}

type WaitForSingleObjectArgs struct {
	HDataValidEvent uintptr
	TimeOut         int
}

type PtrToHeaderArgs struct {
	SharedMemPtr uintptr
}

type PtrToSharedMemArgs struct {
	SharedMemPtr uintptr
}

type PtrToVarHeaderArgs struct {
	VarHeaderPtr uintptr
}

type GetMemoryArgs struct {
	SharedMemPtr uintptr
	StartByte    int
	EndByte      int
}
