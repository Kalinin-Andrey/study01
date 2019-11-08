package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"reflect"
	"strings"
	"text/template"
)

// код писать тут

// Итак, что нужно выцепить
type Structures map[string]*Structure

type Structure struct {
	Name    string
	Methods Methods
}

type Methods map[string]*Method

type Method struct {
	Name			string
	StructName		string
	Params			map[string]string
	ResultTypes		[]string
	GenHandlerData GenHandlerDataStruct
}

// {"url": "/user/create", "auth": true, "method": "POST"}
type GenHandlerDataStruct struct {
	Url    string `json:"url"`
	Auth   bool   `json:"auth"`
	Method string `json:"method"`
}


// Шаблоны
func getTemplate() (*template.Template) {
	//	под структуру
	//		wraperTypeAndObjTpl: тип myApiWrapper и переменная myApiWrapperObj
	//		serveHTTPTlp
t := template.Must(template.New("structureTpl").Parse(`
// Wrapper for the struct {{.Name}}
	type {{.Name}}WrapperStruct struct {
		api     *{{.Name}}
		IsError bool
		Error   ApiError
		result  CR
	}
	// Object of wrapper for the struct {{.Name}}
	var {{.Name}}Wrapper = {{.Name}}WrapperStruct{
		IsError: false,
		result: CR{
			"error": "",
		},
	}
	
	func (apiWrapper *{{.Name}}WrapperStruct) SetError(err error) {

		if err == nil {
			return
		}
		apiWrapper.IsError = true

		error, ok := err.(ApiError)
		if !ok {
			apiWrapper.Error = ApiError{
				Err:        err,
				HTTPStatus: http.StatusInternalServerError,
			}
		} else {
			apiWrapper.Error = error
		}

		/*apiWrapper.Error = ApiError{
			Err:        fmt.Errorf("apiWrapper.returnResult() error: %v", err.Error()),
			HTTPStatus: http.StatusInternalServerError,
		}*/

		if apiWrapper.IsError {
			apiWrapper.result["error"] = apiWrapper.Error.Err.Error()
		}

	}

	func (apiWrapper *{{.Name}}WrapperStruct) GetError() ApiError {
		return apiWrapper.Error
	}

	func (apiWrapper *{{.Name}}WrapperStruct) GetIsError() bool {
		return apiWrapper.IsError
	}

	
	func (apiWrapper *{{.Name}}WrapperStruct) SetResponse(res interface{}) {

		if res == nil || reflect.ValueOf(res).IsNil() {
			return
		}
		apiWrapper.result["response"] = res

	}
	
	func (apiWrapper *{{.Name}}WrapperStruct) GetResult() CR {
		return apiWrapper.result
	}
	
	func (api *{{.Name}}) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		{{.Name}}Wrapper.api = api
		{{.Name}}Wrapper.IsError = false
		{{.Name}}Wrapper.result = CR{
			"error": "",
		}
	
		switch r.URL.Path {

{{range $methodName, $method := .Methods}}
		case "{{$method.GenHandlerData.Url}}":
			handlerFunc := {{.StructName}}Wrapper.{{$method.Name}}
	{{if $method.GenHandlerData.Method}}
			//	method check
			handlerFunc = checkMethodsMiddleware(handlerFunc, &{{.StructName}}Wrapper, "{{$method.GenHandlerData.Method}}")
	{{end}}
	{{if $method.GenHandlerData.Auth}}
			//	access check
			handlerFunc = accessMiddleware(handlerFunc, &{{.StructName}}Wrapper)
	{{end}}
			handlerFunc = panicMiddleware(handlerFunc, &{{.StructName}}Wrapper)
			handlerFunc(w, r)
{{end}}
		default:
			notFound(w, r, &{{.Name}}Wrapper)
		}
	
	}
	
{{range .Methods}}
	{{template "methodTpl" .}}
{{end}}
`))
	//	под методы
	//		setHandlerFuncTpl
	//		handlerFuncTpl
	t = template.Must(t.New("methodTpl").Parse(`
// Wrapper for the method {{.Name}} of the structure {{.StructName}}
func (apiWrapper *{{.StructName}}WrapperStruct) {{.Name}}(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

{{range $paramName, $paramType := .Params}}
	{{if (ne $paramName "ctx")}}
	params := &{{$paramType}}{} // <--
	{{end}}
{{end}}

	err := validateAndGetParams(w, r, params)
	if err != nil {
		returnResult(w, apiWrapper, nil, ApiError{
			Err:        err,
			HTTPStatus: http.StatusBadRequest,
		})
		return
	}

	res, err := apiWrapper.api.{{.Name}}(ctx, *params) // <--

	returnResult(w, apiWrapper, res, err)
}

`))
	//	один общий шаблон
	//		validateAndGetParams
	//		returnResult
	//		все midleware
	t = template.Must(t.New("commonTpl").Parse(`

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

{{range .}}
	{{if .Methods}}
		{{template "structureTpl" .}}
	{{end}}
{{end}}

type ResultWrapper interface {
	SetError(error)
	GetError() ApiError
	GetIsError() bool
	SetResponse(interface{})
	GetResult() CR
}

func notFound(w http.ResponseWriter, r *http.Request, rw ResultWrapper) {
	returnResult(w, rw, nil, ApiError{
		HTTPStatus: http.StatusNotFound,
		Err:        fmt.Errorf("unknown method"),
	})
	return
}

func returnResult(w http.ResponseWriter, rw ResultWrapper, res interface{}, err error) {

	rw.SetError(err)
	rw.SetResponse(res)

	if rw.GetIsError() {
		w.WriteHeader(rw.GetError().HTTPStatus)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	response, err := json.Marshal(rw.GetResult())
	if err != nil {
		http.Error(w, "returnResult() json.Marshal() error: "+err.Error(), http.StatusInternalServerError)
	}
	log.Printf("response: %v", string(response))
	
	w.Write(response)
}

func validateAndGetParams(w http.ResponseWriter, r *http.Request, params interface{}) error {
	paramsValElem := reflect.ValueOf(params).Elem()
	posts := make(map[string]interface{}, len(r.Form))

	err := r.ParseForm()
	if err != nil {
		log.Fatalf("myApiWrapper.validateAndGetParams() http.Request.ParseForm() error: %v", err.Error())
		return fmt.Errorf("myApiWrapper.validateAndGetParams() http.Request.ParseForm() error: %v", err.Error())
	}

	for key, values := range r.Form {
		posts[key] = values[0]
	}

	rules := make(map[string]map[string]interface{}, paramsValElem.NumField())

	for i := 0; i < paramsValElem.NumField(); i++ {
		valueField := paramsValElem.Field(i)
		typeField := paramsValElem.Type().Field(i)
		typeOfField := valueField.Type().String()
		paramName := strings.ToLower(typeField.Name)
		//log.Println("typeOfField " + paramName + " = " + typeOfField)
		rules[typeField.Name] = make(map[string]interface{}, 6)

		if ruleStr, ok := typeField.Tag.Lookup("apivalidator"); ok {
			// filling rules
			for _, r := range strings.Split(ruleStr, ",") {

				if r == "required" {
					rules[typeField.Name]["required"] = true
				} else {
					ruleParts := strings.Split(r, "=")

					switch ruleParts[0] {
					case "paramname":
						rules[typeField.Name]["paramname"] = ruleParts[1]
						paramName = ruleParts[1]
					case "min":
						rules[typeField.Name]["min"] = ruleParts[1]
					case "max":
						rules[typeField.Name]["max"] = ruleParts[1]
					case "enum":
						rules[typeField.Name]["enum"] = strings.Split(ruleParts[1], "|")
					case "default":
						rules[typeField.Name]["default"] = ruleParts[1]

					}
				}
			}

			if _, ok = posts[paramName]; !ok {
				if defaultVal, ok := rules[typeField.Name]["default"]; ok {
					posts[paramName] = defaultVal
				}
			}

			if _, ok = posts[paramName]; ok {
				switch typeOfField {
				case "int":
					i, err := strconv.Atoi(posts[paramName].(string))
					posts[paramName] = i
					if err != nil {
						return fmt.Errorf("%v must be int", paramName)
					}
				case "uint64":
					posts[paramName], _ = strconv.ParseInt(posts[paramName].(string), 10, 64)
				case "string":
					posts[paramName] = posts[paramName].(string)
				default:
					log.Fatalln("apiWrapper.validateAndGetParams() error: conversion type fall")
				}
			}
			// validating by rules
			for ruleKey, ruleValue := range rules[typeField.Name] {

				switch ruleKey {
				case "required":
					if _, ok = posts[paramName]; !ok {
						return fmt.Errorf("%v must me not empty", paramName)
					}
				case "paramname":
				case "min":

					switch posts[paramName].(type) {
					case int:
						ruleValue, _ := strconv.Atoi(ruleValue.(string))
						if posts[paramName].(int) < ruleValue {
							return fmt.Errorf("%v must be >= %v", paramName, ruleValue)
						}
					case int64:
						ruleValue, _ := strconv.ParseInt(ruleValue.(string), 10, 64)
						if posts[paramName].(int64) < ruleValue {
							return fmt.Errorf("%v must be >= %v", paramName, ruleValue)
						}
					case string:
						ruleValue, _ := strconv.Atoi(ruleValue.(string))
						if len(posts[paramName].(string)) < ruleValue {
							return fmt.Errorf("%v len must be >= %v", paramName, ruleValue)
						}
					}
				case "max":

					switch posts[paramName].(type) {
					case int:
						ruleValue, _ := strconv.Atoi(ruleValue.(string))
						if posts[paramName].(int) > ruleValue {
							return fmt.Errorf("%v must be <= %v", paramName, ruleValue)
						}
					case int64:
						ruleValue, _ := strconv.ParseInt(ruleValue.(string), 10, 64)
						if posts[paramName].(int64) > ruleValue {
							return fmt.Errorf("%v must be <= %v", paramName, ruleValue)
						}
					case string:
						ruleValue, _ := strconv.Atoi(ruleValue.(string))
						if len(posts[paramName].(string)) > ruleValue {
							return fmt.Errorf("%v len must be <= %v", paramName, ruleValue)
						}
					}
				case "enum":
					data, ok := rules[typeField.Name]["enum"].([]string)
					if !ok {
						return fmt.Errorf("myApiWrapper.validateAndGetParams() error: cannot convert %v into []string", rules[typeField.Name]["enum"])
					}
					ok = false

					for _, item := range data {
						if item == posts[paramName] {
							ok = true
							break
						}
					}

					if !ok {
						s := rules[typeField.Name]["enum"].([]string)
						return fmt.Errorf("%v must be one of [%v]", paramName, strings.Join(s, ", "))
					}
				case "default":
					if _, ok = posts[paramName]; !ok {
						posts[paramName] = ruleValue
					}
				}
			}
		}
		val := reflect.ValueOf(posts[paramName])

		if !valueField.CanSet() {
			return fmt.Errorf("myApiWrapper.validateAndGetParams() CanSet value: %#v to paramsValElem Field: %#v", posts[paramName], valueField)
		}
		valueField.Set(val)
	}

	log.Printf("Validation: msg is correct\n\n")
	return nil
}

func accessMiddleware(nextFunc http.HandlerFunc, rw ResultWrapper) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header.Get("X-Auth") != "100500" {
			returnResult(w, rw, nil, ApiError{
				Err:        fmt.Errorf("unauthorized"),
				HTTPStatus: http.StatusForbidden,
			})
			return
		}
		login := r.FormValue("login")

		switch login {
		case "bad_username":
			{
				returnResult(w, rw, nil, ApiError{
					Err:        fmt.Errorf("bad user"),
					HTTPStatus: http.StatusInternalServerError,
				})
				return
			}
		case "not_exist_user":
			{
				returnResult(w, rw, nil, ApiError{
					Err:        fmt.Errorf("user not exist"),
					HTTPStatus: http.StatusNotFound,
				})
				return
			}
		}

		nextFunc(w, r)
	})
}

func checkMethodsMiddleware(nextFunc http.HandlerFunc, rw ResultWrapper, method string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != method {
			returnResult(w, rw, nil, ApiError{
				Err:        fmt.Errorf("bad method"),
				HTTPStatus: http.StatusNotAcceptable,
			})
			return
		}

		nextFunc(w, r)
	})
}

func logMiddleware(nextFunc http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		nextFunc(w, r)
	})
}

func panicMiddleware(nextFunc http.HandlerFunc, rw ResultWrapper) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		nextFunc(w, r)

		defer func() {
			if err := recover(); err != nil {
				returnResult(w, rw, nil, ApiError{
					Err:        fmt.Errorf("panic: %v", err),
					HTTPStatus: http.StatusInternalServerError,
				})
				return
			}
		}()
	})
}

`))
return t
}

var Structs Structures = make(Structures, 10)

func main() {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// Сделаем, всё таки, два прогона:
	// первым - собираем все структуры
	for _, f := range node.Decls {

		if g, ok := f.(*ast.GenDecl); ok {
			log.Printf("GenDecl: %#v", g)

			for _, spec := range g.Specs {
				currType, ok := spec.(*ast.TypeSpec)
				if !ok {
					log.Printf("SKIP %#T is not ast.TypeSpec\n", spec)
					continue
				}

				currStruct, ok := currType.Type.(*ast.StructType)
				if !ok {
					log.Printf("SKIP %#T is not ast.StructType\n", currStruct)
					continue
				}
				structName := currType.Name.Name
				Structs[structName] = &Structure{
					Name: structName,
					Methods: make(Methods, 10),
				}
				log.Printf("Get structure: %v\n", Structs[structName])
			}

		} else {
			log.Printf("SKIP %#T is not *ast.GenDecl\n", f)
			continue
		}
	}
	// вторым - собираем методы для генерации
	for _, f := range node.Decls {

		if g, ok := f.(*ast.FuncDecl); ok {
			log.Printf("FuncDecl: %#v", g)

			if g.Doc == nil || reflect.ValueOf(g.Doc).IsNil() || len(g.Doc.List) < 1 {
				continue
			}
			field := g.Recv.List[0]
			strucName := field.Type.(*ast.StarExpr).X.(*ast.Ident).Name

			str, ok := Structs[strucName]
			if !ok {
				log.Printf("Method name=%v of structure name=%v, that not in Structs list: %v", g.Name, strucName, Structs)
				continue
			}
			needCodegen := false
			genHandlerData := GenHandlerDataStruct{}

			for _, comment := range g.Doc.List {

				if strings.HasPrefix(comment.Text, "// apigen:api ") {
					commentStr := strings.TrimSpace(strings.TrimPrefix(comment.Text, "// apigen:api "))
					err = json.Unmarshal([]byte(commentStr), &genHandlerData)
					if err != nil {
						log.Fatalln("Codegen -> json.Unmarshal() error: " + err.Error())
						return
					}

					log.Printf("handlerData: %#v\n", genHandlerData)
					needCodegen = true
					break
				}
			}

			if !needCodegen {
				log.Printf("SKIP struct %#v doesnt have cgen mark\n", g.Name.Name)
				continue
			}

			str.Methods[g.Name.Name] = &Method{
				Name:           g.Name.Name,
				StructName:		str.Name,
				Params:			make(map[string]string, 2),
				ResultTypes:	make([]string, 0, 2),
				GenHandlerData: genHandlerData,
			}
			// filling params
			for _, param := range g.Type.Params.List {
				paramName := param.Names[0].Name
				paramType := ""

				if paramTypeObj, ok := param.Type.(*ast.SelectorExpr); ok {
					paramType = paramTypeObj.X.(*ast.Ident).Name
				}

				if paramTypeObj, ok := param.Type.(*ast.Ident); ok {
					paramType = paramTypeObj.Name
				}
				str.Methods[g.Name.Name].Params[paramName] = paramType
			}
			// filling resulting types
			//resultTypes := str.Methods[g.Name.Name].ResultTypes

			for _, res := range g.Type.Results.List {
				resType := ""

				if resTypeObj, ok := res.Type.(*ast.StarExpr); ok {
					resType = resTypeObj.X.(*ast.Ident).Name
				}

				if resTypeObj, ok := res.Type.(*ast.Ident); ok {
					resType = resTypeObj.Name
				}
				str.Methods[g.Name.Name].ResultTypes = append(str.Methods[g.Name.Name].ResultTypes, resType)
			}
			//str.Methods[g.Name.Name].ResultTypes = resultTypes

			log.Printf("Get method %v.%v()\n", strucName, g.Name.Name)
		} else {
			log.Printf("SKIP %#T is not *ast.GenDecl\n", f)
			continue
		}
	}
	log.Printf("Init data is completed: %v\n", Structs)
	log.Printf("\n\nStart to building output file!\n\n")

	out, _ := os.Create(os.Args[2])

	fmt.Fprintln(out, `package `+node.Name.Name)
	//commonTpl.Execute(out, Structs)
	//fmt.Fprintln(out)      // empty line
	/*t := template.Must(template.New("outerTpl").Parse(`
outerTpl
{{template "innerTpl"}}
`))

	t = template.Must(t.New("innerTpl").Parse(`
innerTpl
`))*/



	t := getTemplate()

	t.ExecuteTemplate(out, "commonTpl", Structs)

	log.Println("All done!")
}
