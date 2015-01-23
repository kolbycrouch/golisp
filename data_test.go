// Copyright 2015 SteelSeries ApS.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package implements a basic LISP interpretor for embedding in a go program for scripting.
// This file tests function evaluation.

package golisp

import (
	. "launchpad.net/gocheck"
	"fmt"
	"unsafe"
)

type SizeSuite struct {
}

var _ = Suite(&SizeSuite{})

func (s *SizeSuite) SetUpSuite(c *C) {

}

func (s *SizeSuite) TestSize(c *C) {
	d := *IntegerWithValue(42)
	fmt.Printf("Data size is :%d\n", unsafe.Sizeof(d))
}