package main

import (
	"fmt"

	. "github.com/dave/jennifer/jen"
)

func main() {
	f := NewFile("diag")
	f.HeaderComment("Code generated by ./cmd/generate-zerolog; DO NOT EDIT.")

	f.Comment("MsgData fields functions").Line()

	type fieldFunctionsDef struct {
		name      string
		valueType *Statement
	}

	fieldFunctions := []fieldFunctionsDef{
		{name: "Str", valueType: String()},
		{name: "Strs", valueType: Index().String()},
		{name: "Stringer", valueType: Qual("fmt", "Stringer")},
		{name: "Bytes", valueType: Index().Byte()},
	}

	for _, fieldFunction := range fieldFunctions {
		f.Func().Params(
			Id("d").Op("*").Id("zerologLogData"),
		).Id(fieldFunction.name).Params(
			Id("key").String(),
			Id("value").Add(fieldFunction.valueType),
		).Id("MsgData").Block(
			Return(Op("&").Id("zerologLogData").Values(
				Dict{
					Id("Event"): Id("d").Dot("Event").Dot(fieldFunction.name).Call(Id("key"), Id("value")),
				},
			)),
		).Line()
	}
	fmt.Printf("%#v", f)
}
