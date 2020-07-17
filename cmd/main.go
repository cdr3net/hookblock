package main

import (
	"os"
	"reflect"

	"github.com/dbolotin/deadmanswitch/blocks"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func main() {
	parser := hclparse.NewParser()
	writer := hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 80, true)

	var allDiag hcl.Diagnostics

	f, diag := parser.ParseHCLFile("./example.hcl")
	allDiag = append(allDiag, diag...)

	if allDiag.HasErrors() {
		_ = writer.WriteDiagnostics(diag)
		return
	}

	// conf := &Config{}
	// context := &hcl.EvalContext{
	// 	Variables: map[string]cty.Value{
	// 		"env": cty.MapVal(map[string]cty.Value{
	// 			"MY_SECRET_1": cty.StringVal("uggugug"),
	// 		}),
	// 	},
	// 	Functions: nil,
	// }
	// diag = gohcl.DecodeBody(f.Body, context, conf)
	// allDiag = append(allDiag, diag...)

	// registry := blocks.HBlockRegistry()
	// schema := hcl.BodySchema{}
	//
	// for _, f := range registry {
	// 	b := f()
	// 	bodySchema, _ := gohcl.ImpliedBodySchema(b)
	// 	schema.Blocks = append(schema.Blocks, bodySchema.Blocks...)
	// }
	//
	// fmt.Println()
	//
	// schema, _ := gohcl.ImpliedBodySchema(&blocks.HConfig{})
	//
	// content, diag := f.Body.Content(schema)
	// allDiag = append(allDiag, diag...)
	//
	// if allDiag.HasErrors() {
	// 	_ = writer.WriteDiagnostics(diag)
	// 	return
	// }
	//
	// log.Println(content)

	h := &blocks.HttpServer{}

	body := f.Body.(*hclsyntax.Body)

	cType := cty.Capsule("channel", reflect.ChanOf(reflect.BothDir, reflect.TypeOf("")))
	cc := make(chan string)
	ccc := cty.CapsuleVal(cType, &cc)

	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"dms0": ccc,
		},
		Functions: nil,
	}

	diag = gohcl.DecodeBody(body.Blocks[0].Body, ctx, h)

	allDiag = append(allDiag, diag...)

	if allDiag.HasErrors() {
		_ = writer.WriteDiagnostics(diag)
		return
	}

	// f.Body.Content()
}
