package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	dataVal := reflect.ValueOf(data)
	//dataValElem := reflect.ValueOf(data).Elem()
	outVal := reflect.ValueOf(out)
	outValElem := outVal.Elem()
	outType := reflect.TypeOf(out)
	//outValElem := reflect.ValueOf(out).Elem()

	//if outType.Kind() != reflect.Ptr && outType.Kind() != reflect.Uintptr {
	if outType.Kind() != reflect.Ptr {
		return fmt.Errorf("Parameter out must be a Ptr")
	}

	if !outValElem.CanSet() {
		return fmt.Errorf("!outValElem.CanSet()")
	}
	outPtrType := reflect.Indirect(outVal).Kind()

	switch outPtrType {
	//switch dataVal.Kind() {
	case reflect.Bool:
		res, ok := data.(bool)
		if !ok {
			return fmt.Errorf("Wrong type of data")
		}
		outValElem.SetBool(res)
	case reflect.Int:
		fl, ok := data.(float64)
		if !ok {
			return fmt.Errorf("Wrong type of data")
		}
		res := int64(fl)
		outValElem.SetInt(res)
	case reflect.Float32, reflect.Float64:
		res, ok := data.(float64)
		if !ok {
			return fmt.Errorf("Wrong type of data")
		}
		outValElem.SetFloat(res)
	case reflect.String:
		res, ok := data.(string)
		if !ok {
			return fmt.Errorf("Wrong type of data")
		}
		outValElem.SetString(res)
	case reflect.Slice:
		len := dataVal.Len()
		elemType := outType.Elem().Elem()
		slice := reflect.MakeSlice(outValElem.Type(), 0, len)

		for i := 0; i < len; i++ {
			dataElem := dataVal.Index(i)
			elem := reflect.New(elemType)

			err := i2s(dataElem.Interface(), elem.Interface())
			if err != nil {
				return err
			}
			//elemVal := reflect.ValueOf(elem)
			//elemValElem := elemVal.Elem()
			e := reflect.Indirect(elem)
			slice = reflect.Append(slice, e)
			//outValElem = reflect.Append(outValElem, e)
		}
		outValElem.Set(slice)

		//slice := valueReflectOut.Elem().FieldByName(key.String())
		//for i := 0; i < valueSlice.Len(); i++ {
		//	slice = reflect.Append(slice, outSlice.Index(i))
		//}
		//valueReflectOut.Elem().FieldByName(key.String()).Set(slice)
	case reflect.Map:
		return fmt.Errorf("Kind Map\n")
	case reflect.Struct:
		iter := dataVal.MapRange()

		for iter.Next() {
			k := iter.Key()
			v := iter.Value()

			field := outValElem.FieldByName(k.String())

			if !field.CanAddr() {
				panic("Cannot get address!")
			}
			err := i2s(v.Interface(), field.Addr().Interface())
			if err != nil {
				return err
			}
			//res[k.String()] = val
		}
		//outValElem.Set(reflect.ValueOf(&res))
	default:
		return fmt.Errorf("Kind unknown! Kind() = %v ; data = %#v\n", dataVal.Kind(), data)
	}


	//fmt.Println(dataVal)
	//fmt.Println(outVal)


	return nil
}

