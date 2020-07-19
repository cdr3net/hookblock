package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/dbolotin/deadmanswitch/bctx"
	"github.com/dbolotin/deadmanswitch/blocks"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func main() {
	// HCL Parser
	parser := hclparse.NewParser()

	// Diagnostics writer
	writer := hcl.NewDiagnosticTextWriter(os.Stdout, parser.Files(), 80, true)
	allDiag := hcl.Diagnostics{}

	if len(os.Args) != 2 {
		fmt.Println("Wrong number of arguments.")
		os.Exit(1)
	}

	// Parsing config file
	f, diag := parser.ParseHCLFile(os.Args[1])
	allDiag = allDiag.Extend(diag)
	if allDiag.HasErrors() {
		_ = writer.WriteDiagnostics(diag)
		os.Exit(1)
	}

	// Creating main context
	bCtx := bctx.NewCtx(writer)

	// Setting context variables

	// Environment variables
	envs := make(map[string]cty.Value)
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		envs[pair[0]] = cty.StringVal(pair[1])
	}
	bCtx.DefaultVariables["env"] = cty.MapVal(envs)

	// First round of config interpretation

	blockVariables := make(map[string]cty.Value)
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

		if _, exist := blockVariables[id]; exist {
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
		blockVariables[id] = block.GetValue(bCtx)
	}

	if allDiag.HasErrors() {
		_ = writer.WriteDiagnostics(allDiag)
		os.Exit(1)
	}

	// Second round of config interpretation

	// Add all default variables to block variables
	for k, v := range bCtx.DefaultVariables {
		blockVariables[k] = v
	}
	ctx := &hcl.EvalContext{
		Variables: blockVariables,
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
}
