package main


import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)



type CR map[string]interface{}
	

	
		
// Wrapper for the struct MyApi
	type MyApiWrapperStruct struct {
		api     *MyApi
		IsError bool
		Error   ApiError
		result  CR
	}
	// Object of wrapper for the struct MyApi
	var MyApiWrapper = MyApiWrapperStruct{
		IsError: false,
		result: CR{
			"error": "",
		},
	}
	
	func (apiWrapper *MyApiWrapperStruct) SetError(err error) {

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

	func (apiWrapper *MyApiWrapperStruct) GetError() ApiError {
		return apiWrapper.Error
	}

	func (apiWrapper *MyApiWrapperStruct) GetIsError() bool {
		return apiWrapper.IsError
	}

	
	func (apiWrapper *MyApiWrapperStruct) SetResponse(res interface{}) {

		if res == nil || reflect.ValueOf(res).IsNil() {
			return
		}
		apiWrapper.result["response"] = res

	}
	
	func (apiWrapper *MyApiWrapperStruct) GetResult() CR {
		return apiWrapper.result
	}
	
	func (api *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		MyApiWrapper.api = api
		MyApiWrapper.IsError = false
		MyApiWrapper.result = CR{
			"error": "",
		}
	
		switch r.URL.Path {


		case "/user/create":
			handlerFunc := MyApiWrapper.Create
	
			//	method check
			handlerFunc = checkMethodsMiddleware(handlerFunc, &MyApiWrapper, "POST")
	
	
			//	access check
			handlerFunc = accessMiddleware(handlerFunc, &MyApiWrapper)
	
			handlerFunc = panicMiddleware(handlerFunc, &MyApiWrapper)
			handlerFunc(w, r)

		case "/user/profile":
			handlerFunc := MyApiWrapper.Profile
	
	
			handlerFunc = panicMiddleware(handlerFunc, &MyApiWrapper)
			handlerFunc(w, r)

		default:
			notFound(w, r, &MyApiWrapper)
		}
	
	}
	

	
// Wrapper for the method Create of the structure MyApi
func (apiWrapper *MyApiWrapperStruct) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()


	

	
	params := &CreateParams{} // <--
	


	err := validateAndGetParams(w, r, params)
	if err != nil {
		returnResult(w, apiWrapper, nil, ApiError{
			Err:        err,
			HTTPStatus: http.StatusBadRequest,
		})
		return
	}

	res, err := apiWrapper.api.Create(ctx, *params) // <--

	returnResult(w, apiWrapper, res, err)
}



	
// Wrapper for the method Profile of the structure MyApi
func (apiWrapper *MyApiWrapperStruct) Profile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()


	

	
	params := &ProfileParams{} // <--
	


	err := validateAndGetParams(w, r, params)
	if err != nil {
		returnResult(w, apiWrapper, nil, ApiError{
			Err:        err,
			HTTPStatus: http.StatusBadRequest,
		})
		return
	}

	res, err := apiWrapper.api.Profile(ctx, *params) // <--

	returnResult(w, apiWrapper, res, err)
}




	

	

	
		
// Wrapper for the struct OtherApi
	type OtherApiWrapperStruct struct {
		api     *OtherApi
		IsError bool
		Error   ApiError
		result  CR
	}
	// Object of wrapper for the struct OtherApi
	var OtherApiWrapper = OtherApiWrapperStruct{
		IsError: false,
		result: CR{
			"error": "",
		},
	}
	
	func (apiWrapper *OtherApiWrapperStruct) SetError(err error) {

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

	func (apiWrapper *OtherApiWrapperStruct) GetError() ApiError {
		return apiWrapper.Error
	}

	func (apiWrapper *OtherApiWrapperStruct) GetIsError() bool {
		return apiWrapper.IsError
	}

	
	func (apiWrapper *OtherApiWrapperStruct) SetResponse(res interface{}) {

		if res == nil || reflect.ValueOf(res).IsNil() {
			return
		}
		apiWrapper.result["response"] = res

	}
	
	func (apiWrapper *OtherApiWrapperStruct) GetResult() CR {
		return apiWrapper.result
	}
	
	func (api *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		OtherApiWrapper.api = api
		OtherApiWrapper.IsError = false
		OtherApiWrapper.result = CR{
			"error": "",
		}
	
		switch r.URL.Path {


		case "/user/create":
			handlerFunc := OtherApiWrapper.Create
	
			//	method check
			handlerFunc = checkMethodsMiddleware(handlerFunc, &OtherApiWrapper, "POST")
	
	
			//	access check
			handlerFunc = accessMiddleware(handlerFunc, &OtherApiWrapper)
	
			handlerFunc = panicMiddleware(handlerFunc, &OtherApiWrapper)
			handlerFunc(w, r)

		default:
			notFound(w, r, &OtherApiWrapper)
		}
	
	}
	

	
// Wrapper for the method Create of the structure OtherApi
func (apiWrapper *OtherApiWrapperStruct) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()


	

	
	params := &OtherCreateParams{} // <--
	


	err := validateAndGetParams(w, r, params)
	if err != nil {
		returnResult(w, apiWrapper, nil, ApiError{
			Err:        err,
			HTTPStatus: http.StatusBadRequest,
		})
		return
	}

	res, err := apiWrapper.api.Create(ctx, *params) // <--

	returnResult(w, apiWrapper, res, err)
}




	

	

	

	

	


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

