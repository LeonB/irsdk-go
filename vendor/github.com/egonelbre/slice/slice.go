// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package slice provides a slice sorting function.
//
// It uses gross, low-level operations to make it easy to sort
// arbitrary slices with only a less function, without defining a new
// type with Len and Swap operations.
package slice

import (
	"fmt"
	"reflect"
	"sort"
	"unsafe"
)

const useReflectSwap = false

const ptrSize = unsafe.Sizeof((*int)(nil))

// Sort sorts the provided slice using the function less.
// If slice is not a slice, Sort panics.
func Sort(slice interface{}, less func(i, j int) bool) {
	sort.Sort(SortInterface(slice, less))
}

// SortInterface returns a sort.Interface to sort the provided slice
// using the function less.
func SortInterface(slice interface{}, less func(i, j int) bool) sort.Interface {
	sv := reflect.ValueOf(slice)
	if sv.Kind() != reflect.Slice {
		panic(fmt.Sprintf("slice.Sort called with non-slice value of type %T", slice))
	}

	slen := sv.Len()

	if useReflectSwap {
		return &swapReflect{
			temp:  reflect.New(sv.Type().Elem()).Elem(),
			slice: sv,
			less:  less,
			len:   slen,
		}
	}

	size := sv.Type().Elem().Size()

	var baseMem unsafe.Pointer
	if slen > 0 {
		baseMem = unsafe.Pointer(sv.Index(0).Addr().Pointer())
	}

	header := unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(baseMem),
		Len:  slen,
		Cap:  slen,
	})

	// Check whether there is a specialized struct swapper
	if sorter, ok := swapStruct(size, less, header); ok {
		return sorter
	}

	// Make a properly-typed (for GC) chunk of memory for swap
	// operations.
	temp := reflect.New(sv.Type().Elem()).Elem()
	tempMem := unsafe.Pointer(temp.Addr().Pointer())
	ms := newMemSwap(size, baseMem, tempMem)
	ms.less = less
	ms.len = slen
	return ms
}

func newMemSwap(size uintptr, baseMem, tempMem unsafe.Pointer) *swapMem {
	tempSlice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: uintptr(tempMem),
		Len:  int(size),
		Cap:  int(size),
	}))
	ms := &swapMem{
		imem: *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(baseMem), Len: int(size), Cap: int(size)})),
		jmem: *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{Data: uintptr(baseMem), Len: int(size), Cap: int(size)})),
		temp: tempSlice,
		size: size,
		base: baseMem,
	}
	ms.ibase = (*uintptr)(unsafe.Pointer(&ms.imem))
	ms.jbase = (*uintptr)(unsafe.Pointer(&ms.jmem))
	return ms
}

// swapReflect is the pure reflect-based swap. It's compiled out by
// default because it's ridiculously slow. But it's kept here in case
// you want to see for yourself.
type swapReflect struct {
	temp  reflect.Value
	slice reflect.Value
	less  func(i, j int) bool
	len   int
}

func (s *swapReflect) Len() int           { return s.len }
func (s *swapReflect) Less(i, j int) bool { return s.less(i, j) }

func (s *swapReflect) Swap(i, j int) {
	s.temp.Set(s.slice.Index(i))
	s.slice.Index(i).Set(s.slice.Index(j))
	s.slice.Index(j).Set(s.temp)
}

// swapMem swaps regions of memory
type swapMem struct {
	imem  []byte
	jmem  []byte
	temp  []byte   // properly typed slice of memory to use as temp space
	ibase *uintptr // ibase points to the Data word of imem
	jbase *uintptr // jbase points to the Data word of jmem
	size  uintptr
	base  unsafe.Pointer
	less  func(i, j int) bool
	len   int
}

func (s *swapMem) Len() int           { return s.len }
func (s *swapMem) Less(i, j int) bool { return s.less(i, j) }

func (s *swapMem) Swap(i, j int) {
	imem, jmem, temp := s.imem, s.jmem, s.temp
	base, size := s.base, s.size
	*(*uintptr)(unsafe.Pointer(&imem)) = uintptr(base) + size*uintptr(i)
	*(*uintptr)(unsafe.Pointer(&jmem)) = uintptr(base) + size*uintptr(j)
	copy(temp, imem)
	copy(imem, jmem)
	copy(jmem, temp)
}

// swapPtr swaps pointers.
type swapPtr struct {
	slice []uintptr
	less  func(i, j int) bool
}

func (s *swapPtr) Len() int           { return len(s.slice) }
func (s *swapPtr) Less(i, j int) bool { return s.less(i, j) }
func (s *swapPtr) Swap(i, j int)      { s.slice[i], s.slice[j] = s.slice[j], s.slice[i] }
