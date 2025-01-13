// program that reads all file references
// ./ProcMdV2 /in=infile /out=outfile /list= [/dbg]
// uses goldmark: github.com/yuin/goldmark
//
// author: prr, azul software
// date: 9 Jan 2025
// copyright 2025 prr, azul software
//
// V2

package main

import (
	"fmt"
	"log"
	"os"
	preproc "goDemo/gmark/MdPreProc"
	util "github.com/prr123/utility/utilLib"
)

func main() {

	numarg := len(os.Args)
    flags:=[]string{"dbg", "in", "out", "list"}

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

    flagMap, err := util.ParseFlags(os.Args, flags)
    if err != nil {log.Fatalf("util.ParseFlags: %v\n", err)}

    dbg:= false
    _, ok := flagMap["dbg"]
    if ok {dbg = true}

    list:= false
    _, ok = flagMap["list"]
    if ok {list = true}

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
	} else {
        if outval.(string) == "none" {
			outFil = inFil
		} else {
	        outFil = outval.(string)
		}
    }

	inFilnam := "md/" + inFil + ".md"
	outFilnam := "md/" + outFil + "_out.md"
	listFilnam := "md/" + outFil + "_list.yaml"

	if dbg {
		fmt.Printf("input:  %s\n", inFilnam)
		fmt.Printf("output: %s\n", outFilnam)
		if list {fmt.Printf("list:    %s\n", listFilnam)}
	}

	source, err := os.ReadFile(inFilnam)
	if err != nil {log.Fatalf("error -- open inp file: %v\n", err)}

	// get file references
	fileList, err := preproc.GetMdFiles(source)
	if err != nil {log.Fatalf("error -- GetMdFiles: %v\n", err)}

	err = preproc.CreateYaml(fileList, listFilnam)
	if err != nil {log.Fatalf("error -- Create Yaml List: %v\n", err)}

	dest, err := preproc.SubstMd(source)
	if err != nil {log.Fatalf("error -- Substituting: %v\n", err)}

	// save
	err = os.WriteFile(outFilnam, dest, 0666)
	if err != nil {log.Fatalf("error -- write file: %v\n")}

	log.Println("*** success ***")
}
