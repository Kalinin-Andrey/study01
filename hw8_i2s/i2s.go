package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	dataVal := reflect.ValueOf(data)
	//dataValElem := reflect.ValueOf(data).Elem()
	outVal := reflect.ValueOf(out).Elem()
	//outValElem := reflect.ValueOf(out).Elem()

	if outVal.Kind() != reflect.Ptr {
		return fmt.Errorf("Parameter out must be a Ptr")
	}
	outPtr := outVal.Addr().Interface().(*interface{})


	switch dataVal.Kind() {
	case reflect.Bool:
		res := dataVal.Bool()
		//outVal.SetBool(res)
		*outPtr = res
	case reflect.Int:
		res := int64(dataVal.Int())
		outVal.SetInt(res)
	case reflect.Float32, reflect.Float64:
		res := int64(dataVal.Float())
		//if outVal.CanSet() {
		//	outVal.SetInt(res)
		//}
		*outPtr = res
	case reflect.String:
		res := dataVal.String()
		//if outVal.CanSet() {
		//	outVal.SetString(res)
		//}
		*outPtr = res
	case reflect.Slice:
		len := dataVal.Len()
		res := make([]interface{}, len, len)

		for i := 0; i < len; i++ {
			var val interface{}
			v := dataVal.Index(i)
			err := i2s(v.Interface(), &val)
			if err != nil {
				return err
			}
			res = append(res, val)
		}
		//resVal := reflect.ValueOf(res)
		//if outVal.CanSet() {
		//	outVal.Set(resVal)
		//}
		*outPtr = res
		//slice := valueReflectOut.Elem().FieldByName(key.String())
		//for i := 0; i < valueSlice.Len(); i++ {
		//	slice = reflect.Append(slice, outSlice.Index(i))
		//}
		//valueReflectOut.Elem().FieldByName(key.String()).Set(slice)
	case reflect.Map:
		res := make(map[string]interface{}, dataVal.Len())
		iter := dataVal.MapRange()

		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			var val interface{}
			err := i2s(v.Interface(), &val)
			if err != nil {
				return err
			}
			res[k.String()] = val
		}
		//resVal := reflect.ValueOf(res)
		//if outVal.CanSet() {
		//	outVal.Set(resVal)
		//}
		*outPtr = res
	case reflect.Struct:
		return fmt.Errorf("Kind is Struct! Kind() = %v ; data = %#v\n", dataVal.Kind(), data)
	default:
		return fmt.Errorf("Kind unknown! Kind() = %v ; data = %#v\n", dataVal.Kind(), data)
	}


	//for i := 0; i < numOfFields ; i++ {
	//
	//}

	fmt.Println(dataVal.Kind())
	fmt.Println(outVal.Kind())
	//fmt.Println(dataValElem.Kind())
	//fmt.Println(outValElem.Kind())


	return nil
}

