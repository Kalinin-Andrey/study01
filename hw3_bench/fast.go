package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
)

// вам надо написать более быструю оптимальную этой функции

type UserType struct {
	Browsers []string
	// Company  string
	// Country  string
	Email string
	// Job      string
	Name string
	// Phone    string
}
type UsersType []UserType

// var userPool = sync.Pool{
// 	New: func() interface{} {
// 		return make(map[string]interface{})
// 	},
// }

func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	seenBrowsers := make(map[string]string)
	//user := make(map[string]interface{})
	var user UserType
	// uniqueBrowsers := 0
	// foundUsers := ""
	i := 0

	fmt.Fprintln(out, "found users:")

	reader := bufio.NewReader(file)
	// while the array contains values
	for {
		line, err := reader.ReadBytes('\n')

		if err != nil {

			if err == io.EOF {
				break
			} else {
				panic(err)
			}
		}

		l := jlexer.Lexer{Data: line}
		user.UnmarshalEasyJSON(&l)
		// fmt.Printf("line: %#v\n", string(line))
		// fmt.Printf("user: %#v\n", user)

		parsUser(out, i, user, &seenBrowsers)
		// user = nil
		// userPool.Put(user)
		i++
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
	//fmt.Fprintln(out, "len(seenBrowsers)= ", len(seenBrowsers), "    uniqueBrowsers= ", uniqueBrowsers)

}

func FastSearch1(out io.Writer) {
	file, err := os.Open(filePath)
	defer file.Close()

	if err != nil {
		panic(err)
	}

	seenBrowsers := make(map[string]string)
	//user := make(map[string]interface{})
	var user UserType
	// uniqueBrowsers := 0
	// foundUsers := ""
	i := 0

	reader := bufio.NewReader(file)
	decoder := json.NewDecoder(reader)

	fmt.Fprintln(out, "found users:")
	// while the array contains values
	for decoder.More() {
		//user := userPool.Get().(map[string]interface{})
		// decode an array value (user)
		err := decoder.Decode(&user)

		if err != nil {
			panic(err)
		}

		// fmt.Printf("user: %#v\n", user)

		parsUser(out, i, user, &seenBrowsers)
		// user = nil
		// userPool.Put(user)
		i++
	}

	fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
	//fmt.Fprintln(out, "len(seenBrowsers)= ", len(seenBrowsers), "    uniqueBrowsers= ", uniqueBrowsers)

}

func parsUser(out io.Writer, i int, user UserType, seenBrowsers *map[string]string) error {
	isAndroid := false
	isMSIE := false

	for _, browser := range user.Browsers {

		if strings.Contains(browser, "Android") {
			isAndroid = true
		} else if strings.Contains(browser, "MSIE") {
			isMSIE = true
		} else {
			continue
		}

		(*seenBrowsers)[browser] = ""
	}

	if !(isAndroid && isMSIE) {
		return nil
	}
	// log.Println("Android and MSIE user:", user["name"], user["email"])
	email := strings.Replace(user.Email, "@", " [at] ", 1)
	// *foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)
	fmt.Fprintf(out, "[%d] %s <%s>\n", i, user.Name, email)
	return nil
}

func main() {
	out := os.Stdout
	FastSearch(out)
}

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson3486653aDecodeCourseraHw3Bench(in *jlexer.Lexer, out *UserType) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "email":
			out.Email = string(in.String())
		case "name":
			out.Name = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson3486653aEncodeCourseraHw3Bench(out *jwriter.Writer, in UserType) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"Browsers\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		if in.Browsers == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Browsers {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"Email\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Email))
	}
	{
		const prefix string = ",\"Name\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Name))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v UserType) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson3486653aEncodeCourseraHw3Bench(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v UserType) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson3486653aEncodeCourseraHw3Bench(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *UserType) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson3486653aDecodeCourseraHw3Bench(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *UserType) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson3486653aDecodeCourseraHw3Bench(l, v)
}

// Marshaler is an easyjson-compatible unmarshaler interface.
type Unmarshaler interface {
	UnmarshalEasyJSON(w *jlexer.Lexer)
}

// UnmarshalFromReader reads all the data in the reader and decodes as JSON into the object.
func UnmarshalFromReader(r io.Reader, v Unmarshaler) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	l := jlexer.Lexer{Data: data}
	v.UnmarshalEasyJSON(&l)
	return l.Error()
}
