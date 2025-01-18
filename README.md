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

## jsDom

This program changes the converion of a markdown file to render the conversion into a js file that builds the website by accessing the the DOM directly.

### jsDom

Renderer that reads the ast struct and renders the content in js. The renderer saves the output in the script directory.

### extDom

Renderer and Parser (contained in the ast directory) of gfm (gitflavoured markdown) extensions. The resulting AST is rendered as js output file in the script subdirectory.
The goldmark parser parses the document first and produces an AST tree. The extension performs a second pass to see whether a paragraph contains a gfm extension. So far we have implemented only tables.

