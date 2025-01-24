// AstDump.go
// program that converts markdown files into html files
// ./simpleMdCon /in=infile.md /out=outfile.html [/dbg]
// uses goldmark: github.com/yuin/goldmark
//
// author: prr, azul software
// date: 21 Nov 2024
// copyright prr, azul software
//

package main

import (
	"fmt"
	"log"
	"os"
	"io"
//	"bytes"

//    jsdom "goDemo/gmark/goldmark/renderer/jsdom"
    extDom "goDemo/gmark/goldmark/extDom"
	"goDemo/gmark/goldmark"
    "goDemo/gmark/goldmark/text"

	cliutil "github.com/prr123/utility/utilLib"
)

func main() {

//	var buf bytes.Buffer
	
	numarg := len(os.Args)
    flags:=[]string{"dbg", "in", "out"}

    useStr := " /in=infile /out=outfile [/dbg]"
    helpStr := "markdown to html conversion program"

    if numarg > len(flags) +1 {
        fmt.Println("too many arguments in cl!")
        fmt.Println("usage: %s %s\n", os.Args[0], useStr)
        os.Exit(-1)
    }

    if numarg == 1 || (numarg > 1 && os.Args[1] == "help") {
        fmt.Printf("help: %s\n", helpStr)
        fmt.Printf("usage is: %s %s\n", os.Args[0], useStr)
        os.Exit(1)
    }

    flagMap, err := cliutil.ParseFlags(os.Args, flags)
    if err != nil {log.Fatalf("util.ParseFlags: %v\n", err)}

    dbg:= false
    _, ok := flagMap["dbg"]
    if ok {dbg = true}

    inFil := ""
    inval, ok := flagMap["in"]
    if !ok {
		log.Fatalf("error -- no in flag provided!\n")
	} else {
        if inval.(string) == "none" {log.Fatalf("error -- no input file name provided!\n")}
        inFil = inval.(string)
    }

    outFil := ""
    outval, ok := flagMap["out"]
    if !ok {
		outFil = inFil
//		log.Fatalf("error -- no out flag provided!\n")
	} else {
        if outval.(string) == "none" {outFil = inFil}
//{log.Fatalf("error -- no output file name provided!\n")}
        outFil = outval.(string)
    }

	inFilnam := "md/" + inFil + ".md"
	outFilnam := "dump/" + outFil + ".txt"

	if dbg {
		fmt.Printf("input:  %s\n", inFilnam)
		fmt.Printf("output: %s\n", outFilnam)
	}

	source, err := os.ReadFile(inFilnam)
	if err != nil {log.Fatalf("error -- open in file: %v\n",err)}

//	outfil, err := os.Create(outFilnam)
//	if err != nil {log.Fatalf("error -- create out file: %v\n", err)}

	mkd := goldmark.New(goldmark.WithExtensions(extDom.Footnote))

	reader := text.NewReader(source)
	mkdParser := mkd.Parser()

    doc := mkdParser.Parse(reader)

	stout := os.Stdout
	pr, pw, err := os.Pipe()
	os.Stdout = pw
	if err != nil {log.Fatalf("error -- creating pipe: %v\n", err)}

	doc.Dump(source, 2)

	pw.Close()
	out, _ := io.ReadAll(pr)
  	os.Stdout = stout
	// save
	err = os.WriteFile(outFilnam, out, 0666)
	if err != nil {log.Fatalf("error -- write file: %v\n", err)}

	log.Println("*** success ***")
}
