# gmark

experimental fork of goldmark/yuin markdown converter.  

goal is creating an image attribute extension  

## image attribute

parses attributes following an image tag.  

test program: ConvMdToHtmlAttr.go  

## indirect references

This feature adds a new tag to markdown files. 
The syntax of the new tag is: $[file name].  

The tab allows the reference to two file types:
 - csv files
 - other markdown files

### getRefMd

This program parses an md file and creates a yaml list file of all md reference links.


### ProcMdV2

This program parsed an md files and substitutes the content.  

Usage: ./ProcMd2V2 /in=md file [/out=outfile] [/dbg]  
