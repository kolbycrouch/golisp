// Copyright 2017 Dave Astels.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package implements a basic LISP interpretor for embedding in a go program for scripting.
// This file contains a new interpretor based on that in Principles of AI programming.

package golisp

import (
	"fmt"
)

func Eval(x *Data, env *SymbolTableFrame) (result *Data, err error) {
	if SymbolP(x) {
		fmt.Printf("Symbol: %s\n", String(x))
		result = env.ValueOf(x)
	} else if AtomP(x) {
		fmt.Printf("Atom: %s\n", String(x))
		result = x
	} else {
		head := Car(x)
		if IsEqv(head, Intern("quote")) {
			fmt.Printf("quote: %s\n", String(Cadr(x)))
			result = Cadr(x)
		} else if IsEqv(head, Intern("begin")) {
			fmt.Printf("begin\n")
			for cell := Cdr(x); NotNilP(cell); cell = Cdr(cell) {
				result, err = Eval(Car(cell), env)
				if err != nil {
					return
				}
			}
			return
		} else if IsEqv(head, Intern("set!")) {
			fmt.Printf("set!: %s\n", String(Cadr(x)))
			v, err := Eval(Caddr(x), env)
			if err != nil {
				return nil, err
			}
			result = env.BindTo(Cadr(x), v)
		} else if IsEqv(head, Intern("if")) {
			fmt.Printf("if\n")
			c, err := Eval(Second(x), env)
			if err != nil {
				return nil, err
			}
			if BooleanValue(c) {
				result, err = Eval(Third(x), env)
			} else {
				result, err = Eval(Fourth(x), env)
			}
		} else if IsEqv(head, Intern("lambda")) {
			fmt.Printf("lambda\n")
			formals := Second(x)
			if !ListP(formals) && !DottedListP(formals) {
				err = ProcessError(fmt.Sprintf("lambda requires a parameter list but recieved %s.", String(formals)), env)
				return
			}
			params := formals
			body := Cdr(x)
			return FunctionWithNameParamsDocBodyAndParent("unnamed", params, "", body, env), nil
		} else {
			fmt.Printf("expression: %s\n", String(x))

			proc, err := Eval(First(x), env)
			if err != nil {
				return nil, err
			}
			fmt.Printf("%s: %s\n", TypeName(TypeOf(proc)), String(Car(x)))
			var argList *Data
			if FunctionP(proc) || (PrimitiveP(proc) && !PrimitiveValue(proc).Special) {
				args := make([]*Data, 0, Length(x))
				for cell := Cdr(x); NotNilP(cell); cell = Cdr(cell) {
					v, err := Eval(Car(cell), env)
					if err != nil {
						return nil, err
					}
					args = append(args, v)
				}
				argList = ArrayToList(args)
			} else {
				argList = Cdr(x)
			}
			if MacroP(proc) {
				expandedMacro, err := MacroValue(proc).Expand(Cdr(x), env)
				if err != nil {
					return nil, err
				}
				result, err = Eval(expandedMacro, env)
			} else {
				result, err = ApplyWithoutEval(proc, argList, env)
			}
		}
	}
	return
}

//func maybeAdd(s string, exps *Data, ifNil *Data) *Data {
//	if NilP(exps) {
//		return ifNil
//	} else if Length(exps) == 1 {
//		return Car(exps)
//	} else {
//		return Cons(Intern(s), exps)
//	}
//}
