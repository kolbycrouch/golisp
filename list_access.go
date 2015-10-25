// Copyright 2014 SteelSeries ApS.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package implements a basic LISP interpretor for embedding in a go program for scripting.
// This file implements data elements.

package golisp

// Cxr

func WalkList(d *Data, path string) *Data {
	c := d
	for index := len(path) - 1; index >= 0; index-- {
		if c == nil {
			return nil
		}
		if !PairP(c) && !VectorizedListP(c) {
			return nil
		}
		switch path[index] {
		case 'a':
			c = Car(c)
		case 'd':
			c = Cdr(c)
		default:
			c = nil
		}
	}
	return c
}

// Cxxr

func Caar(d *Data) *Data {
	return WalkList(d, "aa")
}

func Cadr(d *Data) *Data {
	if VectorizedListP(d) {
		vect := VectorizedListValue(d)
		if len(vect) >= 2 {
			return vect[1]
		} else {
			return nil
		}
	}
	return WalkList(d, "ad")
}

func Cdar(d *Data) *Data {
	return WalkList(d, "da")
}

func Cddr(d *Data) *Data {
	return WalkList(d, "dd")
}

// Cxxxr

func Caaar(d *Data) *Data {
	return WalkList(d, "aaa")
}

func Caadr(d *Data) *Data {
	return WalkList(d, "aad")
}

func Cadar(d *Data) *Data {
	return WalkList(d, "ada")
}

func Caddr(d *Data) *Data {
	if VectorizedListP(d) {
		vect := VectorizedListValue(d)
		if len(vect) >= 3 {
			return vect[2]
		} else {
			return nil
		}
	}
	return WalkList(d, "add")
}

func Cdaar(d *Data) *Data {
	return WalkList(d, "daa")
}

func Cdadr(d *Data) *Data {
	return WalkList(d, "dad")
}

func Cddar(d *Data) *Data {
	return WalkList(d, "dda")
}

func Cdddr(d *Data) *Data {
	return WalkList(d, "ddd")
}

// nth

func Nth(d *Data, n int) *Data {
	if d == nil || n < 0 || n >= Length(d) {
		return nil
	}

	if VectorizedListP(d) {
		return VectorizedListValue(d)[n]
	}

	var c *Data = d
	for i := n; i > 0; c, i = Cdr(c), i-1 {
	}
	return Car(c)
}

func First(d *Data) *Data {
	return Nth(d, 0)
}

func Second(d *Data) *Data {
	return Nth(d, 1)
}

func Third(d *Data) *Data {
	return Nth(d, 2)
}

func Fourth(d *Data) *Data {
	return Nth(d, 3)
}

func Fifth(d *Data) *Data {
	return Nth(d, 4)
}

func Sixth(d *Data) *Data {
	return Nth(d, 5)
}

func Seventh(d *Data) *Data {
	return Nth(d, 6)
}

func Eighth(d *Data) *Data {
	return Nth(d, 7)
}

func Ninth(d *Data) *Data {
	return Nth(d, 8)
}

func Tenth(d *Data) *Data {
	return Nth(d, 9)
}

func SetNth(list *Data, index int, value *Data) *Data {
	if VectorizedListP(list) {
		vect := VectorizedListValue(list)
		if index < len(vect) {
			vect[index] = value
		}
	} else {
		for i := index; i > 0; list, i = Cdr(list), i-1 {
		}
		if !NilP(list) {
			ConsValue(list).Car = value
		}
	}

	return value
}
