package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
)

type keyValue map[string]interface{}

// sliceOfKeyValue implements sort.Interface for []keyValue based on
// the name field.
type sliceOfKeyValue []keyValue

func (a sliceOfKeyValue) Len() int           { return len(a) }
func (a sliceOfKeyValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sliceOfKeyValue) Less(i, j int) bool { return a[i]["name"].(string) < a[j]["name"].(string) }

func getFileList(path string, isPrintFiles bool) (sliceOfKeyValue, error) {
	var filesRes = make(sliceOfKeyValue, 0, 20)
	var fileInfo keyValue = make(keyValue, 4)

	files, err := ioutil.ReadDir(path)

	if err != nil {
		return filesRes, err
	}

	for _, file := range files {
		path2file := path + string(os.PathSeparator) + file.Name()
		fi, err := os.Lstat(path2file)

		if err != nil {
			return filesRes, err
		}
		isDir := fi.Mode().IsDir()

		if !isPrintFiles && !isDir {
			continue
		}

		fileInfo = keyValue{
			"name":      file.Name(),
			"path2file": path2file,
			"size":      file.Size(),
			"isDir":     isDir,
		}
		filesRes = append(filesRes, fileInfo)

	}
	sort.Sort(sliceOfKeyValue(filesRes))
	return filesRes, nil
}

func dirTreeRec(output io.Writer, path string, isPrintFiles bool, indent string) error {
	const t string = "├"
	const o string = "\t"
	const l string = "└"
	const v string = "│"
	const h string = "───"

	var indent4file string
	var indent4embededFile string
	var filesCount int
	var sFileSize string

	files, err := getFileList(path, isPrintFiles)

	if err != nil {
		return err
	}
	filesCount = len(files)

	for i, file := range files {

		if i+1 == filesCount {
			indent4file = indent + l
			indent4embededFile = indent + o
		} else {
			indent4file = indent + t
			indent4embededFile = indent + v + o
		}

		if file["isDir"].(bool) {
			sFileSize = ""
		} else {

			if file["size"].(int64) == 0 {
				sFileSize = " (empty)"
			} else {
				sFileSize = fmt.Sprintf(" (%vb)", file["size"])
			}
		}
		//fmt.Fprintln(output, indent4file+h+file["name"].(string)+" ("+file["size"].(string)+")")
		fmt.Fprintln(output, indent4file+h+file["name"].(string)+sFileSize)

		if err != nil {
			return err
		}

		if file["isDir"].(bool) {
			dirTreeRec(output, file["path2file"].(string), isPrintFiles, indent4embededFile)
		}
	}

	return nil
}

func dirTree(output io.Writer, path string, isPrintFiles bool) error {
	//fmt.Fprintln(output, dirTreeRec(output, path, isPrintFiles, ""))
	dirTreeRec(output, path, isPrintFiles, "")
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
