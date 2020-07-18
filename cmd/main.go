package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/dbolotin/deadmanswitch/bctx"
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
	allDiag = allDiag.Extend(diag)

	if allDiag.HasErrors() {
		_ = writer.WriteDiagnostics(diag)
		os.Exit(1)
	}

	bCtx := bctx.NewCtx(writer)
	variables := make(map[string]cty.Value)

	// Adding predefined variables
	// TODO

	registry := blocks.BlockRegistry()
	blocksById := make(map[string]blocks.Block)
	var blocksByIndex []blocks.Block

	body := f.Body.(*hclsyntax.Body)
	for _, b := range body.Blocks {
		rng := b.Range()

		factory := registry[b.Type]
		if factory == nil {
			allDiag = allDiag.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Unknown block type",
				Detail:   fmt.Sprintf("Unknown block: %s", b.Type),
				Subject:  &b.TypeRange,
				Context:  &rng,
			})
			break
		}

		block := factory()
		blocksByIndex = append(blocksByIndex, block)

		if len(b.Labels) == 0 {
			block.SetId(b.Type + "_" + strconv.Itoa(len(blocksByIndex)-1))
			break
		}

		id := b.Labels[0]
		block.SetId(id)

		if _, exist := blocksById[id]; exist {
			allDiag = allDiag.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Duplicate block identifier",
				Detail:   fmt.Sprintf("Duplicate block identifier: %s", id),
				Subject:  &b.LabelRanges[0],
				Context:  &rng,
			})
			break
		}

		if _, exist := variables[id]; exist {
			allDiag = allDiag.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Reserved variable name",
				Detail:   fmt.Sprintf("Reserved variable name: %s", id),
				Subject:  &b.LabelRanges[0],
				Context:  &rng,
			})
			break
		}

		blocksById[id] = block
		variables[id] = block.GetValue(bCtx)
	}

	if allDiag.HasErrors() {
		_ = writer.WriteDiagnostics(allDiag)
		os.Exit(1)
	}

	// Second round of parsing

	ctx := &hcl.EvalContext{
		Variables: variables,
	}

	for i, b := range body.Blocks {
		//noinspection GoNilness
		block := blocksByIndex[i]
		diag = gohcl.DecodeBody(b.Body, ctx, block)
		allDiag = allDiag.Extend(diag)
		if diag.HasErrors() {
			break
		}
	}

	if allDiag.HasErrors() {
		_ = writer.WriteDiagnostics(allDiag)
		os.Exit(1)
	}

	// Starting the process
	for _, block := range blocksByIndex {
		err := block.Start(bCtx)
		if err != nil {
			log.Fatalln(err)
		}
	}

	// Block forever
	select {}

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

	// registry := blocks.BlockRegistry()
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
	//
	// cType := cty.Capsule("channel", reflect.ChanOf(reflect.BothDir, reflect.TypeOf("")))
	// cc := make(chan string)
	// ccc := cty.CapsuleVal(cType, &cc)
	//
	// ctx := &hcl.EvalContext{
	// 	Variables: map[string]cty.Value{
	// 		"dms0": ccc,
	// 	},
	// 	Functions: nil,
	// }
	//
	// diag = gohcl.DecodeBody(body.Blocks[0].Body, ctx, h)
	//
	// allDiag = append(allDiag, diag...)

	// f.Body.Content()
}
