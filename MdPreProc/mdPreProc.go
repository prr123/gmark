// preprocessor of markdown files
// program parses a md file and generates an expanded md file
//
// author: prr, azul software
// date: 9 Jan 2025
// copyright 2025 prr, azul software
//

package mdPreProc

import (
	"fmt"
	"os"

	csvLib "github.com/prr123/gocsv/csvLib"

)

func SubstMd(src []byte) ([]byte, error) {


	dbg:=true
	dest := make([]byte, len(src) + 1024*16)

	npos := 0
	istate:=0
	filSt := 0
	extSt := 0
	attrSt := 0
	filNam := ""
	ext := ""
//	filEnd := 0
	for i:=0; i< len(src) -1; i++ {
		switch istate {
		case 0:
			if src[i] == '$' {
				if src[i+1] == '[' {
					istate = 1
					filSt = i+2
					i += 1
				}
			} else {
//fmt.Printf("dbg-- %d [%q] %d %s\n", i, src[i], npos, dest[:npos+1])
				dest[npos] = src[i]
				npos++
			}
		case 1:
			if src[i] == '.' {
				istate = 2
				extSt= i+1
			}

		case 2:
			if src[i] == ']' {
				istate = 3
				filNam = string(src[filSt:i])
				ext = string(src[extSt:i])
				if dbg {
					fmt.Printf("dbg -- ref: %s - %s\n",filNam, ext)
				}
			}

		case 3:
			if src[i] == '\n' || src[i] == '{' {
				switch ext {
				case "csv":
					inData, err := os.ReadFile("md/csv/" + filNam)
					if err != nil {return nil, fmt.Errorf("cannot read: %v", err)}
				    lines, err := csvLib.ProcTable(inData)
				    if err != nil {return nil, fmt.Errorf("ProcTable: %v\n", err)}

				    csvLib.PrintLines(lines)

				    tables, err := csvLib.GetTables(lines)
    				if err != nil {return nil, fmt.Errorf("GetTable: %v\n", err)}
/*
    fmt.Printf("tables: %d\n", len(tables))
    for i:=0; i< len(tables); i++ {
        csvLib.PrintTable(tables[i])
    }
*/
					mdData, err := csvLib.Table2Md(tables[0])
					if err != nil {return nil, fmt.Errorf("Table2Md: %v\n", err)}

					copy(dest[npos: npos+len(mdData)], mdData)
					npos += len(mdData)

				case "md":
					data, err := os.ReadFile("md/" + filNam)
					if err != nil {return nil, fmt.Errorf("cannot read: %v", err)}
					copy(dest[npos: npos+len(data)], data)
					npos += len(data)

				default:
					return nil, fmt.Errorf("not a valid ext: %s", ext)
				}
			}

			if src[i] == '{' {
				istate = 4
				attrSt = i
			} else {
				istate = 0
				if src[i] != '\n' {
					dest[npos] = src[i]
					npos++
				}
			}

		case 4:
			if src[i] == '}' {
				istate = 0
				attrStr := string(src[attrSt:i+1])
//				copy(dest[npos: npos+len(attrStr)], src[attrSt:i+1])
//				npos += len(attrStr)

fmt.Printf("dbg -- table attr: %s\n",attrStr)
			}
		default:
			return nil, fmt.Errorf("invalid state: %d", istate)
		}
//fmt.Printf("dbg2 -- %d [%q] %d %s -- %d\n", i, src[i], npos, dest[:npos+1], len(dest))
	}
fmt.Printf("dbg -- last letter\n")
	// copy last letter
	dest[npos] = src[len(src) -1]

	return dest[:npos+1], nil
}


func GetMdFiles(src []byte) ([]string, error) {

	var filList []string

	istate:=0
	filSt := 0
//	attrSt := 0
//	filEnd := 0
	for i:=0; i< len(src) -3; i++ {
		switch istate {
		case 0:
			if src[i] == '$' {
				if src[i+1] == '[' {
					istate = 1
					filSt = i+2
					i += 1
				}
			}
		case 1:
			if src[i] == ']' {
				istate = 2
				filNam :=string(src[filSt:i])
//fmt.Printf("dbg -- file name:%s\n", filNam)
				filList = append(filList, filNam)
			}
		case 2:
			if src[i] == '{' {
				istate = 3
//				attrSt = i
			} else {
				istate = 0
			}
		case 3:
			if src[i] == '}' {
				istate = 0
//				attrStr := string(src[attrSt+1:i])
//fmt.Printf("dbg -- table attr: %s\n",attrStr)
			}
		default:
			return nil, fmt.Errorf("invalid state: %d", istate)
		}
//fmt.Printf("dbg2 -- %d [%q] %d %s -- %d\n", i, src[i], npos, dest[:npos+1], len(dest))
	}

	return filList, nil
}

func CreateYaml(filList []string, listFilnam string) error {

	if len(filList) == 0 {return fmt.Errorf("no files!")}

	lFil, err := os.Create(listFilnam)
	if err != nil {return fmt.Errorf("cannot create list file: %v", err)}
	defer lFil.Close()

	fmt.Fprintln(lFil,"---")
	fmt.Fprintln(lFil,"files:")

	for i:=0; i<len(filList); i++ {
		fmt.Fprintf(lFil," - %s: ", filList[i])
		_, err := os.Stat(filList[i])
		if err != nil {
			fmt.Fprintln(lFil,"false")
		} else {
			fmt.Fprintln(lFil, "true")
		}
	}

	return nil
}



func getTable(filnam string)([]byte, error) {

	tbldat, err := os.ReadFile(filnam)
	if err != nil {
		return nil, err
	}
	return tbldat, nil
}

func ProcTable(src []byte)([]string, error) {

	var lines []string

	lineSt :=0
	for i:=0; i<len(src); i++ {
		if src[i] == '\n' {
			lines = append(lines, string(src[lineSt:i]))
			lineSt = i+1
			i += 1
		}
	}
	return lines, nil
}


func PrintLines(lines []string) {

	fmt.Println("*************** Lines ******************")
	for i:=0; i< len(lines); i++ {
		fmt.Printf("--%d: %s\n",i, lines[i])
	}
	fmt.Println("************* End Lines ****************")
}
