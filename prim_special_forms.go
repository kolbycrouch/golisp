// Copyright 2014 SteelSeries ApS.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This package implements a basic LISP interpretor for embedding in a go program for scripting.
// This file contains the built-in primitive functions.

package golisp

import (
	"fmt"
	"strings"
)

var TypeMap map[string]uint32

func initTypeMap() {
	TypeMap = make(map[string]uint32, 20)
	TypeMap["list"] = 0x00000002
	TypeMap["vector"] = 0x00000004
	TypeMap["sequence"] = 0x00000006
	TypeMap["integer"] = 0x00000008
	TypeMap["float"] = 0x00000010
	TypeMap["number"] = 0x00000018
	TypeMap["boolean"] = 0x00000020
	TypeMap["string"] = 0x00000040
	TypeMap["character"] = 0x00000080
	TypeMap["symbol"] = 0x00000100
	TypeMap["stringy"] = 0x00000140
	TypeMap["function"] = 0x00000200
	TypeMap["macro"] = 0x00000400
	TypeMap["primitive"] = 0x00000800
	TypeMap["procedure"] = 0x00000A00
	TypeMap["boxedobject"] = 0x00001000
	TypeMap["frame"] = 0x00002000
	TypeMap["environment"] = 0x00004000
	TypeMap["port"] = 0x00008000
	TypeMap["anytype"] = 0xFFFFFFFF
}

func RegisterSpecialFormPrimitives() {
	MakeSpecialForm("define", ">=1", defineImpl)
	MakeSpecialForm("typedef", ">=1", typeDefImpl)
	MakeSpecialForm("defmacro", ">=1", defMacroImpl)
	MakeSpecialForm("define-macro", ">=1", defMacroImpl)
	MakeSpecialForm("define-compiler-macro", ">=1", defCompilerMacroImpl)
	MakeSpecialForm("do", ">=2", doImpl)
	MakeSpecialForm("apply", ">=1", applyImpl)
	MakeSpecialForm("->", ">=1", chainImpl)
	MakeSpecialForm("=>", ">=1", tapImpl)
	MakeSpecialForm("definition-of", "1", definitionOfImpl)
	MakeSpecialForm("doc", "1", docImpl)
	MakeSpecialForm("type", "1", typeImpl)
	MakeSpecialForm("let", ">=1", letImpl)
	MakeSpecialForm("let*", ">=1", letStarImpl)
	MakeSpecialForm("letrec", ">=1", letRecImpl)

	initTypeMap()
}

func defineImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	var value *Data
	thing := Car(args)
	if SymbolP(thing) {
		//		fmt.Printf("Defining var: %s\n", String(thing))
		value, err = Eval(Cadr(args), env)
		if err != nil {
			return
		}
	} else if ListP(thing) || DottedListP(thing) {
		name := Car(thing)
		params := Cdr(thing)
		thing = name
		//		fmt.Printf("Defining function: %s\n", String(name))
		if !SymbolP(name) {
			err = ProcessError("Function name has to be a symbol", env)
			return
		}
		existingValueOrNil := env.ValueOf(name)
		if PrimitiveP(existingValueOrNil) {
			err = ProcessError(fmt.Sprintf("Primitive function %s can not be redefined", StringValue(name)), env)
			return
		}
		var body *Data = Cdr(args)
		var doc *Data = nil
		if StringP(Car(body)) {
			doc = Car(body)
			body = Cdr(body)
		}
		value = FunctionWithNameParamsDocBodyAndParent(StringValue(name), params, StringValue(doc), body, env)
	} else {
		err = ProcessError(fmt.Sprintf("define expected a symbol or formals list as its first argument but received %s", String(thing)), env)
		return
	}
	_, err = env.BindLocallyTo(thing, value)
	return value, err
}

func typeSpecToTypeMask(typeSpec string, env *SymbolTableFrame) (mask uint32, err error) {
	mask = uint32(0)
	for _, spec := range strings.Split(typeSpec, "|") {
		t := TypeMap[spec]
		if t == 0 {
			err = ProcessError(fmt.Sprintf("typedef specified an invalid type: '%s'", spec), env)
			return
		} else {
			mask = mask | t
		}
	}
	return
}

func typeDefImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	thing := First(args)
	if !SymbolP(thing) {
		err = ProcessError(fmt.Sprintf("typeDef expected a symbol name but received %s", String(Car(thing))), env)
		return
	} else {
		// function
		functionName := StringValue(thing)
		argTypes := make([]uint32, 0, Length(Cdr(thing)))
		var argType uint32
		var returnType uint32 = AnyType
		for cell := Cdr(args); NotNilP(cell); cell = Cdr(cell) {
			if SymbolP(Car(cell)) && StringValue(Car(cell)) == "->" {
				if NilP(Cdr(cell)) {
					err = ProcessError("typeDef expected a symbol name to follow ->", env)
					return
				}
				returnType, err = typeSpecToTypeMask(StringValue(Cadr(cell)), env)
				if err != nil {
					return
				}
				break
			}
			argType, err = typeSpecToTypeMask(StringValue(Car(cell)), env)
			if err != nil {
				return
			}
			argTypes = append(argTypes, argType)
		}
		AddTypesForFunction(functionName, argTypes, returnType)
	}
	return
}

func defMacroImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	var value *Data
	thing := Car(args)
	if ListP(thing) || DottedListP(thing) {
		name := Car(thing)
		params := Cdr(thing)
		thing = name
		if !SymbolP(name) {
			err = ProcessError("Macro name has to be a symbol", env)
			return
		}
		body := Cadr(args)
		value = MacroWithNameParamsBodyAndParent(StringValue(name), params, body, env)
	} else {
		err = ProcessError("Invalid macro definition", env)
		return
	}
	_, err = env.BindLocallyTo(thing, value)
	return value, err
}

func defCompilerMacroImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	// if BooleanValue(Global.ValueOf(CompilingSymbol)) {
	var value *Data
	thing := Car(args)
	if ListP(thing) || DottedListP(thing) {
		name := Car(thing)
		params := Cdr(thing)
		thing = name
		if !SymbolP(name) {
			err = ProcessError("Compiler macro name has to be a symbol", env)
			return
		}
		body := Cadr(args)
		value = CompilerMacroWithNameParamsBodyAndParent(StringValue(name), params, body, env)
	} else {
		err = ProcessError("Invalid compiler macro definition", env)
		return
	}
	env.BindLocallyTo(thing, value)
	return value, nil
	// }
	// fmt.Printf("Attempt to define compiled macro '%s' when not compiling\n", String(First(args)))
	// return
}

func bindLetLocals(bindingForms *Data, rec bool, localEnv *SymbolTableFrame, evalEnv *SymbolTableFrame) (err error) {
	var name *Data
	var value *Data

	for cell := bindingForms; NotNilP(cell); cell = Cdr(cell) {
		bindingPair := Car(cell)
		if !ListP(bindingPair) {
			err = ProcessError("Let requires a list of bindings (with are pairs) as it's first argument", evalEnv)
			return
		}
		name = Car(bindingPair)
		if !SymbolP(name) {
			err = ProcessError("First part of a let binding pair must be a symbol", evalEnv)
			return
		}

		if rec {
			_, err = localEnv.BindLocallyTo(name, nil)
			if err != nil {
				return
			}
		}
	}

	for cell := bindingForms; NotNilP(cell); cell = Cdr(cell) {
		bindingPair := Car(cell)
		name = First(bindingPair)
		if Length(bindingPair) != 1 {
			value, err = Eval(Second(bindingPair), evalEnv)
			if err != nil {
				return
			}
		}
		_, err = localEnv.BindLocallyTo(name, value)
		if err != nil {
			return
		}
	}
	return
}

func letCommon(args *Data, env *SymbolTableFrame, star bool, rec bool) (result *Data, err error) {
	if !ListP(Car(args)) {
		err = ProcessError("Let requires a list of bindings as it's first argument", env)
		return
	}

	localEnv := NewSymbolTableFrameBelow(env, "let")
	localEnv.Previous = env
	var evalEnv *SymbolTableFrame
	if star || rec {
		evalEnv = localEnv
	} else {
		evalEnv = env
	}
	err = bindLetLocals(Car(args), rec, localEnv, evalEnv)
	if err != nil {
		return
	}

	for cell := Cdr(args); NotNilP(cell); cell = Cdr(cell) {
		sexpr := Car(cell)
		result, err = Eval(sexpr, localEnv)
		if err != nil {
			return
		}
	}

	return
}

func namedLetImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	name := First(args)
	bindings := Second(args)
	if !ListP(bindings) {
		err = ProcessError("A named let requires a list of bindings as it's second argument", env)
		return
	}
	body := Cddr(args)

	vars := make([]*Data, 0, Length(bindings))
	initials := make([]*Data, 0, Length(bindings))
	for remainingBindings := bindings; NotNilP(remainingBindings); remainingBindings = Cdr(remainingBindings) {
		binding := Car(remainingBindings)
		if !SymbolP(Car(binding)) {
			err = ProcessError("The first element of a binding must be a symbol", env)
			return
		}
		vars = append(vars, Car(binding))
		initials = append(initials, Cadr(binding))
	}
	varsList := ArrayToList(vars)
	initialsList := ArrayToList(initials)
	localEnv := NewSymbolTableFrameBelow(env, StringValue(name))
	localEnv.Previous = env
	localEnv.BindLocallyTo(name, nil)
	namedLetProc := FunctionWithNameParamsDocBodyAndParent(StringValue(name), varsList, "", body, localEnv)
	localEnv.BindLocallyTo(name, namedLetProc)
	return FunctionValue(namedLetProc).Apply(initialsList, env)
}

func letImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	if SymbolP(Car(args)) {
		return namedLetImpl(args, env)
	} else {
		return letCommon(args, env, false, false)
	}
}

func letStarImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	return letCommon(args, env, true, false)
}

func letRecImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	return letCommon(args, env, false, true)
}

func rebindDoLocals(bindingForms *Data, env *SymbolTableFrame) (err error) {
	var names []*Data
	var values []*Data
	var value *Data
	var name *Data

	for cell := bindingForms; NotNilP(cell); cell = Cdr(cell) {
		bindingTuple := Car(cell)
		name = First(bindingTuple)
		names = append(names, name)
		if NotNilP(Third(bindingTuple)) {
			value, err = Eval(Third(bindingTuple), env)
			if err != nil {
				return
			}
		} else {
			value = env.ValueOf(name)
		}
		values = append(values, value)
	}

	for i := 0; i < len(names); i++ {
		_, err = env.BindLocallyTo(names[i], values[i])
		if err != nil {
			return
		}
	}
	return
}

func doImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	bindings := Car(args)
	if !ListP(bindings) {
		err = ProcessError("Do requires a list of bindings as it's first argument", env)
		return
	}

	testClause := Cadr(args)
	if !ListP(testClause) {
		err = ProcessError("Do requires a list as it's second argument", env)
		return
	}

	localEnv := NewSymbolTableFrameBelow(env, "do")
	localEnv.Previous = env
	err = bindLetLocals(bindings, false, localEnv, env)
	if err != nil {
		return
	}

	body := Cddr(args)

	var shouldExit *Data

	for true {
		shouldExit, err = Eval(Car(testClause), localEnv)
		if err != nil {
			return
		}

		if BooleanValue(shouldExit) {
			for cell := Cdr(testClause); NotNilP(cell); cell = Cdr(cell) {
				sexpr := Car(cell)
				result, err = Eval(sexpr, localEnv)
				if err != nil {
					return
				}
			}
			return
		}

		for cell := body; NotNilP(cell); cell = Cdr(cell) {
			sexpr := Car(cell)
			result, err = Eval(sexpr, localEnv)
			if err != nil {
				return
			}
		}

		if rebindDoLocals(bindings, localEnv) != nil {
			return
		}
	}
	return
}

func applyImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	f, err := Eval(Car(args), env)
	if err != nil {
		return
	}

	if !FunctionOrPrimitiveP(f) {
		err = ProcessError(fmt.Sprintf("apply requires a function as it's first argument, but got %s.", String(f)), env)
		return
	}

	ary := make([]*Data, 0, Length(args)-1)

	var v *Data
	for c := Cdr(args); NotNilP(c); c = Cdr(c) {
		v, err = Eval(Car(c), env)
		if err != nil {
			return
		}
		ary = append(ary, v)
	}

	var argList *Data
	if ListP(ary[len(ary)-1]) {
		if len(ary) > 1 {
			argList = ArrayToListWithTail(ary[0:len(ary)-1], ary[len(ary)-1])
		} else {
			argList = ary[0]
		}
	} else {
		err = ProcessError(fmt.Sprintf("apply requires the last arg to be a list, but got %s", String(ary[len(ary)-1])), env)
		return
	}

	return ApplyWithoutEval(f, argList, env)
}

func chainImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	var value *Data

	value, err = Eval(Car(args), env)
	if err != nil {
		return
	}

	for cell := Cdr(args); NotNilP(cell); cell = Cdr(cell) {
		sexpr := Car(cell)
		var newExpr *Data
		if ListP(sexpr) {
			newExpr = Cons(Car(sexpr), Append(Cdr(sexpr), QuoteIt(value)))
		} else {
			newExpr = Cons(sexpr, Cons(QuoteIt(value), nil))
		}
		value, err = Eval(newExpr, env)
		if err != nil {
			return
		}
	}
	result = value
	return
}

func tapImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	var value *Data

	value, err = Eval(Car(args), env)
	if err != nil {
		return
	}
	result = value

	for cell := Cdr(args); NotNilP(cell); cell = Cdr(cell) {
		sexpr := Car(cell)
		var newExpr *Data
		if ListP(sexpr) {
			newExpr = Cons(Car(sexpr), Append(Cdr(sexpr), QuoteIt(value)))
		} else {
			newExpr = Cons(sexpr, Cons(QuoteIt(value), nil))
		}
		_, err = Eval(newExpr, env)
		if err != nil {
			return
		}
	}
	return
}

func definitionOfImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	var name *Data = nil
	if SymbolP(Car(args)) {
		name = Car(args)
	} else {
		name = Intern("anonymous")
	}

	f, err := Eval(Car(args), env)
	if err != nil {
		return
	}
	if !FunctionP(f) {
		err = ProcessError(fmt.Sprintf("definition-of requires a function argument, but received a %s.", TypeName(TypeOf(f))), env)
		return
	}

	function := FunctionValue(f)
	if function.Name == "unnamed" {
		return Cons(Intern("define"), Cons(name, Cons(Cons(Intern("lambda"), Cons(function.Params, function.Body)), nil))), nil
	} else {
		return Cons(Intern("define"), Cons(Cons(Intern(function.Name), function.Params), function.Body)), nil
	}
}

func docImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	var name *Data = First(args)
	if !SymbolP(name) {
		err = ProcessError(fmt.Sprintf("doc requires a symbol naming a function, but received %s.", String(name)), env)
		return
	}

	f, err := Eval(name, env)
	if err != nil {
		return
	}
	if !FunctionP(f) {
		err = ProcessError(fmt.Sprintf("doc requires a function argument, but received %s.", String(f)), env)
		return
	}

	function := FunctionValue(f)
	if function.DocString == "" {
		return StringWithValue(fmt.Sprintf("%s has no documentation string.", StringValue(name))), nil
	} else {
		return StringWithValue(function.DocString), nil
	}
}

func typeImpl(args *Data, env *SymbolTableFrame) (result *Data, err error) {
	var name *Data = First(args)
	if !SymbolP(name) {
		err = ProcessError(fmt.Sprintf("type requires a symbol naming a function, but received %s.", String(name)), env)
		return
	}

	f, err := Eval(name, env)
	if err != nil {
		return
	}
	if !FunctionP(f) {
		err = ProcessError(fmt.Sprintf("type requires a function argument, but received %s.", String(f)), env)
		return
	}

	function := FunctionValue(f)
	if function.TypeSignature != nil {
		result = function.MakeTypeSpec()
	}
	return
}
