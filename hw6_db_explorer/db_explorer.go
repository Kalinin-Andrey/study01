package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type DbExplorer struct{
	DB		*sql.DB
	Tables	Tables
	TableNames	[]string
}

const GetTables = "GetTables"
const GetRecordsets = "GetRecordsets"
const GetRecordset = "GetRecordset"
const CreateRecordset = "CreateRecordset"
const UpdateRecordset = "UpdateRecordset"
const DeleteRecordset = "DeleteRecordset"

const DefaultLimit = 10
const DefaultOffset = 0


type Tables	map[string]*Table

type Table struct{
	Name	string
	PK		string
	FieldsNames []string
	Fields	map[string]*Field
}

type Field struct {
	Position	int
	Name	string
	Type	string
	IsNull	bool
}

type Recordsets []Recordset

type Recordset map[string]interface{}

type ApiError struct {
	HTTPStatus int
	Err        error
}

type R map[string]interface{}

type HandleData struct {
	tableName	string
	limit		int
	offset		int
	methodName	string
	id			int
	params		Recordset
}

func (t *DbExplorer) init(){
	t.Tables = make(map[string]*Table, 10)
	rows, err := t.DB.Query("SHOW TABLES")
	defer rows.Close()

	if err != nil {
		log.Fatalln(err.Error())
	}

	for rows.Next() {
		table := &Table{}
		err = rows.Scan(&table.Name)
		if err != nil {
			log.Fatalln("DbExplorer.init() error: " + err.Error())
		}
		t.Tables[table.Name] = table
		t.Tables[table.Name].FieldsNames = make([]string, 0, 10)
		t.Tables[table.Name].Fields = make(map[string]*Field, 10)
	}

	for _, table := range t.Tables {
		//rows, err = t.DB.Query("SHOW FULL COLUMNS FROM " + table.Name) // table.Name == "items"
		rows, err = t.DB.Query("select ORDINAL_POSITION, COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_KEY, TABLE_SCHEMA from information_schema.COLUMNS where TABLE_NAME=?;", table.Name)
		if err != nil {
			log.Fatalln(err.Error())
		}
		prevTableSchema := ""
		fields := make([]Field, 0, len(t.Tables[table.Name].FieldsNames))

		for rows.Next() {
			field := &Field{}
			isNullable := ""
			columnKey := ""
			tableSchema := ""
			position := ""
			err = rows.Scan(&position, &field.Name, &field.Type, &isNullable, &columnKey, &tableSchema)
			if err != nil {
				log.Fatalln("DbExplorer.init() error: " + err.Error())
			}
			if prevTableSchema != "" && prevTableSchema != tableSchema {
				continue
			}
			prevTableSchema = tableSchema

			field.Position, err = strconv.Atoi(position)
			if err != nil {
				log.Fatalln("DbExplorer.init() error: " + err.Error())
			}

			field.IsNull = (isNullable == "YES")

			if columnKey == "PRI" {
				table.PK = field.Name
			}
			table.Fields[field.Name] = field
			fields = append(fields, *field)
		}

		sort.Slice(fields, func(i, j int) bool {
			return fields[i].Position < fields[j].Position
		})
		for _, field := range fields {
			table.FieldsNames = append(table.FieldsNames, field.Name)
		}
		log.Printf("table: %#v\n\n", table)
	}
	log.Printf("\n\nt.Tables: %#v\n\n", t.Tables)

}

func (t *DbExplorer) GetTables() []string{

	if len(t.TableNames) == 0 {
		for tableName, _ := range t.Tables {
			t.TableNames = append(t.TableNames, tableName)
		}
	}
	sort.Strings(t.TableNames)
	return t.TableNames
}

func (t *DbExplorer) GetRecordsets(tableName string, limit int, offset int) (recordsets Recordsets, apiError ApiError) {
	table, ok := t.Tables[tableName]
	if !ok {
		return nil, ApiError{
			HTTPStatus: http.StatusNotFound,
			Err: fmt.Errorf("Not found"),
		}
	}

	recordsets = make([]Recordset, 0, limit)
	rows, err := t.DB.Query("SELECT * FROM " + table.Name + " LIMIT " + strconv.Itoa(offset) + ", " + strconv.Itoa(limit))
	defer rows.Close()

	if err != nil {
		return nil, ApiError{
			HTTPStatus: http.StatusInternalServerError,
			Err: fmt.Errorf("Query error"),
		}
	}

	for rows.Next() {
		// init
		count := len(t.Tables[table.Name].Fields)
		rst := make(Recordset, count)
		columns := make(Recordset, count)
		valuePtr := make([]interface{}, 0, count)

		for _, fieldName := range t.Tables[table.Name].FieldsNames {
			columns[fieldName] = new(interface{})
			valuePtr = append(valuePtr, columns[fieldName])
		}
		// get
		err = rows.Scan(valuePtr...)
		if err != nil {
			return nil, ApiError{
				HTTPStatus: http.StatusInternalServerError,
				Err:        err,
			}
		}
		// convert
		rst := t.recordsetTypeCasting(table.Name, valuePtr)
		log.Printf("Table: %v\n", table.Name)
		for i, fieldName := range t.Tables[table.Name].FieldsNames {
			log.Printf("field: %v\n", fieldName)
			/*var v interface{}
			val := rst[fieldName]*/
			val := valuePtr[i].(*interface{})
			//rst[fieldName] = *val
			//v := *val
			var v interface{}
			var value interface{}

			switch val := (*val).(type) {
			case []byte:
				log.Printf("val byte: %#v", val)
				v = string(val)
			case nil:
				log.Printf("val is nil: %#v", val)
				v = nil
			default:
				log.Printf("val case default: %#v", val)
				panic("Val not a slice byte!")
			}

			if v != nil || !t.Tables[table.Name].Fields[fieldName].IsNull {

				switch t.Tables[table.Name].Fields[fieldName].Type {
				case "int(11)", "bigint(20)":
					value, err = strconv.Atoi(v.(string))
				default:
					value = v.(string)
				}
			}
			rst[fieldName] = value
		}

		recordsets = append(recordsets, rst)
	}
	/*sort.Slice(recordsets, func(i, j int) bool {
		iValElem := reflect.ValueOf(recordsets[i])
		jValElem := reflect.ValueOf(recordsets[j])
		iId := iValElem.FieldByName(t.Tables[table.Name].PK).Int()
		jId := jValElem.FieldByName(t.Tables[table.Name].PK).Int()
		return iId < jId
	})*/

	return recordsets, ApiError{
		HTTPStatus: http.StatusOK,
		Err: nil,
	}
}

func (t *DbExplorer) GetRecordset(tableName string, id int) (rst Recordset, apiError ApiError) {
	table, ok := t.Tables[tableName]
	if !ok {
		return nil, ApiError{
			HTTPStatus: http.StatusNotFound,
			Err: fmt.Errorf("record not found"),
		}
	}

	row := t.DB.QueryRow("SELECT * FROM " + table.Name + " WHERE " + t.Tables[tableName].PK + " = ? ", id)

	// init
	count := len(t.Tables[table.Name].Fields)
	rst = make(Recordset, count)
	valuePtr := make([]interface{}, 0, count)

	for _, fieldName := range t.Tables[table.Name].FieldsNames {
		rst[fieldName] = new(interface{})
		valuePtr = append(valuePtr, rst[fieldName])
	}
	// get
	err := row.Scan(valuePtr...)
	if err != nil {
		log.Fatalln(err.Error())
	}
	// http.StatusNotFound "record not found"
	// convert
	for _, fieldName := range t.Tables[table.Name].FieldsNames {
		var v interface{}
		val := rst[fieldName]

		b, ok := val.([]byte)

		if (ok) {
			v = string(b)
		} else {
			v = val
		}

		switch t.Tables[table.Name].Fields[fieldName].Type {
		case "int(11)", "bigint(20)":

			if val, ok := v.(string); ok {
				v, err = strconv.Atoi(val)
				if err != nil {
					log.Fatalln(err.Error())
					return nil, ApiError{
						HTTPStatus: http.StatusInternalServerError,
						Err: err,
					}
				}
			} else if val, ok := v.(int); ok {
				v = val
			}
		}
		rst[fieldName] = v
	}


	return rst, ApiError{
		HTTPStatus: http.StatusOK,
		Err: nil,
	}

}

func (t *DbExplorer) FilterPK(tableName string, fieldsNames []string) (newFieldsNames []string, err error) {
	count := len(fieldsNames)
	pk := t.Tables[tableName].PK
	newFieldsNames = make([]string, 0, count)

	for _, v := range fieldsNames {
		if v != pk {
			newFieldsNames = append(newFieldsNames, v)
		}
	}
	return newFieldsNames, nil
}
func (t *DbExplorer) CreateRecordset(tableName string, recordset Recordset) (int64, ApiError) {
	// <-- валидация
	fieldsNames, err := t.FilterPK(tableName, t.Tables[tableName].FieldsNames)
	if err != nil {
		return 0, ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err:        err,
		}
	}
	count := len(fieldsNames)
	var vals []interface{} = make([]interface{}, 0, count)

	for _, fieldName := range fieldsNames {
		vals = append(vals, recordset[fieldName])
	}
	fmt.Printf(
		"INSERT INTO " + tableName + " (`" + strings.Join(fieldsNames, "`, `") + "`) VALUES (" + strings.TrimRight(strings.Repeat("%v,", count), ",") + ")",
		vals...
	)
	result, err := t.DB.Exec(
		"INSERT INTO " + tableName + " (`" + strings.Join(fieldsNames, "`, `") + "`) VALUES (" + strings.TrimRight(strings.Repeat("?,", count), ",") + ")",
		vals...
	)
	if err != nil {
		log.Fatalln(err.Error())
		return 0, ApiError{
			HTTPStatus: http.StatusInternalServerError,
			Err: nil,
		}
	}

	_, err = result.RowsAffected()
	if err != nil {
		log.Fatalln(err.Error())
		return 0, ApiError{
			HTTPStatus: http.StatusInternalServerError,
			Err: nil,
		}
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		log.Fatalln(err.Error())
		return 0, ApiError{
			HTTPStatus: http.StatusInternalServerError,
			Err: nil,
		}
	}

	return lastID, ApiError{
		HTTPStatus: http.StatusOK,
		Err: nil,
	}
}

func (t *DbExplorer) UpdateRecordset(tableName string, id int, recordset Recordset) (int64, ApiError) {
	pk := t.Tables[tableName].PK
	if _, ok := recordset[pk]; ok {
		return 0, ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err: fmt.Errorf("field id have invalid type"),
		}
	}
	// <-- валидация
	count := len(recordset)
	var fieldsNames []string = make([]string, 0, count)
	var vals []interface{} = make([]interface{}, 0, count)

	for fieldName, fieldVal := range recordset {
		fieldsNames = append(fieldsNames, fieldName)
		vals = append(vals, fieldVal)
	}
	vals = append(vals, id)
	fmt.Printf(
		"UPDATE " + tableName + " SET `" + strings.Join(fieldsNames, "` = %v, `") + "` = %v WHERE `" + pk + "` = %v\n",
		vals...
	)
	result, err := t.DB.Exec(
		"UPDATE " + tableName + " SET `" + strings.Join(fieldsNames, "` = ?, `") + "` = ? WHERE `" + pk + "` = ?",
		vals...
	)
	if err != nil {
		log.Fatalln(err.Error())
		return 0, ApiError{
			HTTPStatus: http.StatusInternalServerError,
			Err: nil,
		}
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		log.Fatalln(err.Error())
		return 0, ApiError{
			HTTPStatus: http.StatusInternalServerError,
			Err: nil,
		}
	}

	return affectedRows, ApiError{
		HTTPStatus: http.StatusOK,
		Err: nil,
	}
}

func (t *DbExplorer) DeleteRecordset(tableName string, id int) (int64, ApiError) {
	pk := t.Tables[tableName].PK
	// <-- валидация
	fmt.Printf(
		"DELETE FROM " + tableName + " WHERE `" + pk + "` = %v\n",
		id,
	)
	result, err := t.DB.Exec(
		"DELETE FROM " + tableName + " WHERE `" + pk + "` = ?",
		id,
	)
	if err != nil {
		log.Fatalln(err.Error())
		return 0, ApiError{
			HTTPStatus: http.StatusInternalServerError,
			Err: nil,
		}
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		log.Fatalln(err.Error())
		return 0, ApiError{
			HTTPStatus: http.StatusInternalServerError,
			Err: nil,
		}
	}

	return affectedRows, ApiError{
		HTTPStatus: http.StatusOK,
		Err: nil,
	}
}

func (t *DbExplorer) HttpHandle(w http.ResponseWriter, r *http.Request) {
	data, apiError := t.validateAndGetData(r)
	if apiError.Err != nil {
		returnResult(w, nil, apiError)
		return
	}
	result := make(R, 1)
	apiError = ApiError{
		HTTPStatus: http.StatusOK,
		Err: nil,
	}

	switch data.methodName{
	case GetTables:
		res := t.GetTables()
		result["response"] = map[string][]string{
			"tables": res,
		}
	case GetRecordsets:
		res, err := t.GetRecordsets(data.tableName, data.limit, data.offset)
		if err.Err != nil {
			returnResult(w, nil, err)
			return
		}
		result["response"] = map[string]Recordsets{
			"records": res,
		}
	case GetRecordset:
		res, err := t.GetRecordset(data.tableName, data.id)
		if err.Err != nil {
			returnResult(w, nil, err)
			return
		}
		result["response"] = map[string]Recordset{
			"record": res,
		}
	case CreateRecordset:
		res, err := t.CreateRecordset(data.tableName, data.params)
		if err.Err != nil {
			returnResult(w, nil, err)
			return
		}
		result["response"] = map[string]int64{
			"id": res,
		}
	case UpdateRecordset:
		res, err := t.UpdateRecordset(data.tableName, data.id, data.params)
		if err.Err != nil {
			returnResult(w, nil, err)
			return
		}
		result["response"] = map[string]int64{
			"updated": res,
		}
	case DeleteRecordset:
		res, err := t.DeleteRecordset(data.tableName, data.id)
		if err.Err != nil {
			returnResult(w, nil, err)
			return
		}
		result["response"] = map[string]int64{
			"deleted": res,
		}
	}

	if apiError.Err != nil {
		returnResult(w, nil, apiError)
		return
	}

	returnResult(w, result, apiError)
	return
}

func (t *DbExplorer) validateAndGetData(r *http.Request) (HandleData, ApiError) {
	var data HandleData

	apiError := t.validateAndSetMethod(r, &data)
	if apiError.Err != nil {
		return data, apiError
	}

	apiError = t.validateAndSetParams(r, &data)
	if apiError.Err != nil {
		return data, apiError
	}

	return data, ApiError{
		HTTPStatus: http.StatusOK,
		Err: nil,
	}
}

func (t *DbExplorer) validateAndSetMethod(r *http.Request, data *HandleData) (ApiError) {
	path := strings.Trim(r.URL.Path, "/")
	pathArr := strings.Split(path, "/")
	count := len(pathArr)

	if count > 2 {
		return ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err:        fmt.Errorf("Address is fail."),
		}
	}

	switch {
	case count == 0 || (count == 1 && pathArr[0] == ""):
		data.methodName = GetTables
	case count == 1:
		data.tableName = pathArr[0]
		data.methodName = GetRecordsets
	case count == 2:
		data.tableName = pathArr[0]
		id, err := strconv.Atoi(pathArr[1])
		if err != nil {
			return ApiError{
				HTTPStatus: http.StatusBadRequest,
				Err:        fmt.Errorf("id mast be integer, err: %v", err.Error()),
			}
		}
		data.id = id
		data.methodName = GetRecordset
	}

	switch r.Method {
	case http.MethodGet:
		if data.methodName == GetRecordsets {
			data.limit = DefaultLimit
			data.offset = DefaultOffset
			p := r.URL.Query().Get("limit")

			if p != "" {
				limit, err := strconv.Atoi(p)
				if err != nil {
					return ApiError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("limit mast be integer, err: %v", err.Error()),
					}
				}
				if limit > 0 {
					data.limit = limit
				}
			}
			p = r.URL.Query().Get("offset")

			if p != "" {
				offset, err := strconv.Atoi(p)
				if err != nil {
					return ApiError{
						HTTPStatus: http.StatusBadRequest,
						Err:        fmt.Errorf("offset mast be integer, err: %v", err.Error()),
					}
				}
				if offset >= 0 {
					data.offset = offset
				}
			}
		}
	case http.MethodPut:
		if data.tableName == "" {
			return ApiError{
				HTTPStatus: http.StatusBadRequest,
				Err:        fmt.Errorf("Put: table name mast be set"),
			}
		}
		if data.id > 0 {
			data.id = 0 // игнорируем при вставке
		}
		data.methodName = CreateRecordset
	case http.MethodPost:
		if data.tableName == "" {
			return ApiError{
				HTTPStatus: http.StatusBadRequest,
				Err:        fmt.Errorf("Post: table name mast be set"),
			}
		}
		if data.id == 0 {
			return ApiError{
				HTTPStatus: http.StatusBadRequest,
				Err:        fmt.Errorf("Post: id mast be set"),
			}
		}
		data.methodName = UpdateRecordset
	case http.MethodDelete:
		if data.tableName == "" {
			return ApiError{
				HTTPStatus: http.StatusBadRequest,
				Err:        fmt.Errorf("Delete: table name mast be set"),
			}
		}
		if data.id == 0 {
			return ApiError{
				HTTPStatus: http.StatusBadRequest,
				Err:        fmt.Errorf("Delete: id mast be set"),
			}
		}
		data.methodName = DeleteRecordset
	}

	if data.methodName != "GetTables" && data.tableName == "" {
		return ApiError{
			HTTPStatus: http.StatusBadRequest,
			Err:        fmt.Errorf(data.methodName + ": table name mast be set"),
		}
	}

	if _, ok := t.Tables[data.tableName]; !ok && data.tableName != "" {
		return ApiError{
			HTTPStatus: http.StatusNotFound,
			Err:        fmt.Errorf("unknown table"),
		}
	}

	return ApiError{
		HTTPStatus: http.StatusOK,
		Err: nil,
	}
}

func (t *DbExplorer) validateAndSetParams(r *http.Request, data *HandleData) (ApiError) {

	if data.methodName == CreateRecordset || data.methodName == UpdateRecordset {
		bodyBytes, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			return ApiError{
				HTTPStatus: http.StatusBadRequest,
				Err:        fmt.Errorf(data.methodName + ": table name mast be set"),
			}
		}
		var params Recordset
		json.Unmarshal(bodyBytes, params)

		for k, v := range params {
			if _, ok := t.Tables[data.tableName].Fields[k]; !ok {
				continue	// если такого поля нет, то просто игнорим
			}

			switch t.Tables[data.tableName].Fields[k].Type {
			case "int(11)", "bigint(20)":
				v, err = strconv.Atoi(v.(string))
				if err != nil {
					log.Fatalln(err.Error())
					return ApiError{
						HTTPStatus: http.StatusInternalServerError,
						Err: err,
					}
				}
			default:
				v = v.(string)
			}
		}
	}

	return ApiError{
		HTTPStatus: http.StatusOK,
		Err: nil,
	}
}

func returnResult(w http.ResponseWriter, res interface{}, apiError ApiError) {

	if apiError.Err != nil {
		res = R{
			"error": apiError.Err.Error(),
		}
	}

	response, err := json.Marshal(res)
	if err != nil {
		apiError = ApiError{
			HTTPStatus: http.StatusInternalServerError,
			Err: fmt.Errorf("returnResult() json.Marshal() cannot marshal a result: %v ; error: %v", res, err.Error()),
		}
		res = R{
			"error": apiError.Err.Error(),
		}
		response, _ = json.Marshal(res)
	}

	if apiError.Err != nil {
		w.WriteHeader(apiError.HTTPStatus)
		log.Printf("returnResult() status: %v; ", strconv.Itoa(apiError.HTTPStatus))
	} else {
		w.WriteHeader(http.StatusOK)
		log.Printf("returnResult() status: %v; ", strconv.Itoa(http.StatusOK))
	}
	log.Printf("returnResult() response: %v", string(response))

	w.Write(response)
}

func NewDbExplorer(db *sql.DB) (http.Handler, error){
	dbExplorer := &DbExplorer{
		DB: db,
	}
	dbExplorer.init()

	//test(*dbExplorer)

	mux := http.NewServeMux()
	mux.HandleFunc("/", dbExplorer.HttpHandle)

	return mux, nil
}

func test (dbExplorer DbExplorer){
	/*recordset, apiError := dbExplorer.GetRecordset("users", 1)
	if apiError.Err != nil {
		log.Fatalln(apiError.Err.Error())
	}
	log.Printf("recordset: %#v", recordset)
	*/
	tableName := "items"
	id := 45
	/*recordset := &Recordset{
		"id":          42,
		"title":       "db_crud",
		"description": "",
	}

	_, apiError := dbExplorer.CreateRecordset(tableName, *recordset)
	if apiError.Err != nil {
		log.Fatalln(apiError.Err.Error())
	}*/

	/*recordset := &Recordset{
		"title":       "db_crud1",
		"description": "1",
	}

	_, apiError := dbExplorer.UpdateRecordset(tableName, id, *recordset)
	if apiError.Err != nil {
		log.Fatalln(apiError.Err.Error())
	}*/

	_, apiError := dbExplorer.DeleteRecordset(tableName, id)
	if apiError.Err != nil {
		log.Fatalln(apiError.Err.Error())
	}

}