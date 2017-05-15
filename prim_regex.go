// Copyright 2016 Dave AStels.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package implements a basic LISP interpretor for embedding in a go program for scripting.
// This file contains the primitive functions for regular expression support.

package golisp

import (
	"regexp"
)

func RegisterRegexPrimitives() {
	MakeTypedPrimitiveFunction("re-string-match-go", "2", ReStringMatchImpl, []uint32{StringType, StringType})
}

func ReStringMatchImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	regexPattern := StringValue(First(args))
	stringToMatch := StringValue(Second(args))

	re := regexp.MustCompile(regexPattern)
	capturedStrings := re.FindStringSubmatch(stringToMatch)

	if capturedStrings == nil {
		result = LispFalse
		return
	} else {
		strs := make([]*Data, 0, len(capturedStrings))
		for _, s := range capturedStrings {
			strs = append(strs, StringWithValue(s))
		}
		result = ArrayToList(strs)
		return
	}
}