// simpleMd2JsConvV3.go
// program that converts markdown files into html files
// ./simpleMdCon /in=infile.md /out=outfile.html [/dbg]
// uses goldmark: github.com/yuin/goldmark
//
// author: prr, azul software
// date: 22 Nov 2024
// copyright prr, azul software
//
// v3: aggregate child nodes of paragraphs
//	   create funcs for element and style creation

package main

import (
	"fmt"
	"log"
	"os"
	"bytes"

	md2js "goDemo/gmark/goldmark/renderer/jsdom"

	"goDemo/gmark/goldmark"
	util "github.com/prr123/utility/utilLib"
)

func main() {

	var buf bytes.Buffer

	numarg := len(os.Args)
    flags:=[]string{"dbg", "in", "out", "style", "site"}

    useStr := " /in=infile /out=outfile [/style=styleFile] [/site=siteFile] [/dbg]"
    helpStr := "markdown to html conversion program V2"

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

    flagMap, err := util.ParseFlags(os.Args, flags)
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
//log.Fatalf("error -- no output file name provided!\n")}
        outFil = outval.(string)
    }

    stylFil := "mdStyle"
    stylval, ok := flagMap["style"]
    if ok {
        if stylval.(string) == "none" {log.Fatalf("error -- no style file name provided!\n")}
        stylFil = stylval.(string)
    }

    siteFil := "mdSite"
    siteval, ok := flagMap["site"]
    if ok {
        if siteval.(string) == "none" {log.Fatalf("error -- no site file name provided!\n")}
        siteFil = siteval.(string)
    }

	inFilnam := "md/" + inFil + ".md"
	metaFilnam := "md/" + inFil + ".meta"
	outFilnam := "script/" + outFil + ".js"
	stylFilnam := "style/" + stylFil + ".js"
	siteFilnam := "site/" + siteFil + ".js"

	if dbg {
		fmt.Printf("input:  %s\n", inFilnam)
		fmt.Printf("output: %s\n", outFilnam)
		fmt.Printf("style:  %s\n", stylFilnam)
		fmt.Printf("site:   %s\n", siteFilnam)
		fmt.Printf("meta:   %s\n", metaFilnam)
	}

	mdData, err := os.ReadFile(inFilnam)
	if err != nil {log.Fatalf("error -- open file: %v\n", err)}

	metaData, err := os.ReadFile(metaFilnam)
	if err != nil {log.Printf("info -- no meta file: %v\n", err)}

	stylData, err := os.ReadFile(stylFilnam)
	if err != nil {log.Printf("info -- no style file: %v\n", err)}

	siteData, err := os.ReadFile(siteFilnam)
	if err != nil {log.Printf("info -- no style file: %v\n", err)}

	oFil, err := os.Create(outFilnam)
	if err != nil {log.Fatalf("error -- create out File: %v\n", err)}
	defer oFil.Close()

	if len(metaData) > 0 {
		mData, err := md2js.GetMeta(metaData)
		if err !=nil {log.Fatal("error -- converting meta: %v\n", err)}
		md2js.PrintMeta(mData)
	}

	startMdStr := md2js.JSRenderStartFunc()
	_, err = oFil.Write(startMdStr)
	if err != nil {log.Fatalf("error -- writing md start Render: %v\n")}

	_, err = oFil.Write(stylData)
	if err != nil {log.Fatalf("error -- writing style: %v\n", err)}

	name:= "test"
	md2jsRen := md2js.GetRenderer(name, dbg)
	md := goldmark.New()
	md.SetRenderer(md2jsRen)

	// retrieve yaml data from mdData if present

	// retrieve summary md data from mdData if present

	// retrieve body md data from mdData if present

// func Convert(source []byte, w io.Writer, opts ...parser.ParseOption) error
//	err = goldmark.Convert(source, &buf, parser.WithContext(ctx))

	errcon := md.Convert(mdData, &buf)
	if errcon != nil {
		log.Printf("error -- converting: %v\n",errcon)
	} else {
		log.Printf("*** success converting ***\n")
	}
	// save
//fmt.Printf("dbg -- buf length: %d\n", len(buf.Bytes()))
	_, err = oFil.Write(buf.Bytes())
	if err != nil {log.Fatalf("error -- writing md js body: %v\n")}

	_, err = oFil.Write(siteData)
	if err != nil {log.Fatalf("error -- writing site: %v\n", err)}

	if errcon != nil {
		log.Println("*** error conversion ***")
	} else {
		log.Println("*** success ***")
	}
}
