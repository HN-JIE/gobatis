package gobatis

import (
	"database/sql"
	"errors"
	"log"
	"reflect"
)

type resultTypeProc = func(rows *sql.Rows, res interface{}) error

var resSetProcMap = map[ResultType]resultTypeProc{
	resultTypeMap:     resMapProc,
	resultTypeMaps:    resMapsProc,
	resultTypeSlice:   resSliceProc,
	resultTypeSlices:  resSlicesProc,
	resultTypeValue:   resValueProc,
	resultTypeStructs: resStructsProc,
	resultTypeStruct:  resStructProc,
}

func resStructProc(rows *sql.Rows, res interface{}) error {
	resVal := reflect.ValueOf(res)
	if resVal.Kind() != reflect.Ptr {
		return errors.New("Struct query result must be ptr")
	}

	resVal = resVal.Elem()
	if resVal.Kind() != reflect.Struct {
		return errors.New("Struct query result must be struct")
	}

	arr, err := rowsToStructs(rows, resVal.Type())
	if nil != err {
		return err
	}

	if len(arr) > 1 {
		return errors.New("Struct query result more than one row")
	}

	if len(arr) > 0 {
		resVal.Set(reflect.ValueOf(arr[0]).Elem())
	}

	return nil
}

func resStructsProc(rows *sql.Rows, res interface{}) error {
	sliceVal := reflect.ValueOf(res)
	if sliceVal.Kind() != reflect.Ptr {
		return errors.New("Structs query result must be ptr")
	}

	slicePtr := reflect.Indirect(sliceVal)
	if slicePtr.Kind() != reflect.Slice && slicePtr.Kind() != reflect.Array {
		return errors.New("Structs query result must be slice")
	}

	//get elem type
	elem := slicePtr.Type().Elem()
	resultType := elem
	isPtr := elem.Kind() == reflect.Ptr
	if isPtr {
		resultType = elem.Elem()
	}

	if resultType.Kind() != reflect.Struct {
		return errors.New("Structs query results item must be struct")
	}

	arr, err := rowsToStructs(rows, resultType)
	if nil != err {
		return err
	}

	for i := 0; i < len(arr); i++ {
		if isPtr {
			slicePtr.Set(reflect.Append(slicePtr, reflect.ValueOf(arr[i])))
		} else {
			slicePtr.Set(reflect.Append(slicePtr, reflect.Indirect(reflect.ValueOf(arr[i]))))
		}
	}

	return nil
}

func resValueProc(rows *sql.Rows, res interface{}) error {
	resPtr := reflect.ValueOf(res)
	if resPtr.Kind() != reflect.Ptr {
		return errors.New("Value query result must be ptr")
	}

	arr, err := rowsToSlices(rows)
	if nil != err {
		return err
	}

	if len(arr) > 1 {
		return errors.New("Value query result more than one row")
	}

	tempResSlice := arr[0].([]interface{})
	if len(tempResSlice) > 1 {
		return errors.New("Value query result more than one col")
	}

	if len(tempResSlice) > 0 {
		if nil != tempResSlice[0] {
			value := reflect.Indirect(resPtr)
			value.Set(reflect.ValueOf(tempResSlice[0]))
		}

	}

	return nil
}

func resSlicesProc(rows *sql.Rows, res interface{}) error {
	resPtr := reflect.ValueOf(res)
	if resPtr.Kind() != reflect.Ptr {
		return errors.New("Slices query result must be ptr")
	}

	value := reflect.Indirect(resPtr)
	if value.Kind() != reflect.Slice {
		return errors.New("Slices query result must be slice ptr")
	}

	arr, err := rowsToSlices(rows)
	if nil != err {
		return err
	}

	for i := 0; i < len(arr); i++ {
		value.Set(reflect.Append(value, reflect.ValueOf(arr[i])))
	}

	return nil
}

func resSliceProc(rows *sql.Rows, res interface{}) error {
	resPtr := reflect.ValueOf(res)
	if resPtr.Kind() != reflect.Ptr {
		return errors.New("Slice query result must be ptr")
	}

	value := reflect.Indirect(resPtr)
	if value.Kind() != reflect.Slice {
		return errors.New("Slice query result must be slice ptr")
	}

	arr, err := rowsToSlices(rows)
	if nil != err {
		return err
	}

	if len(arr) > 1 {
		return errors.New("Slice query result more than one row")
	}

	if len(arr) > 0 {
		tempResSlice := arr[0].([]interface{})
		value.Set(reflect.AppendSlice(value, reflect.ValueOf(tempResSlice)))
	}

	return nil
}

func resMapProc(rows *sql.Rows, res interface{}) error {
	resBean := reflect.ValueOf(res)
	if resBean.Kind() == reflect.Ptr {
		return errors.New("Map query result can not be ptr")
	}

	if resBean.Kind() != reflect.Map {
		return errors.New("Map query result must be map")
	}

	arr, err := rowsToMaps(rows)
	if nil != err {
		return err
	}

	if len(arr) > 1 {
		return errors.New("Map query result more than one row")
	}

	if len(arr) > 0 {
		resMap := res.(map[string]interface{})
		tempResMap := arr[0].(map[string]interface{})
		for k, v := range tempResMap {
			resMap[k] = v
		}
	}

	return nil
}

func resMapsProc(rows *sql.Rows, res interface{}) error {
	resPtr := reflect.ValueOf(res)
	if resPtr.Kind() != reflect.Ptr {
		return errors.New("Maps query result must be ptr")
	}

	value := reflect.Indirect(resPtr)
	if value.Kind() != reflect.Slice {
		return errors.New("Maps query result must be slice ptr")
	}
	arr, err := rowsToMaps(rows)
	if nil != err {
		return err
	}

	for i := 0; i < len(arr); i++ {
		value.Set(reflect.Append(value, reflect.ValueOf(arr[i])))
	}

	return nil
}

func rowsToMaps(rows *sql.Rows) ([]interface{}, error) {
	res := make([]interface{}, 0)
	for rows.Next() {
		resMap := make(map[string]interface{})
		cols, err := rows.Columns()
		if nil != err {
			log.Println(err)
			return res, err
		}

		vals := make([]interface{}, len(cols))
		scanArgs := make([]interface{}, len(cols))
		for i := range vals {
			scanArgs[i] = &vals[i]
		}

		rows.Scan(scanArgs...)
		for i := 0; i < len(cols); i++ {
			val := vals[i]
			if nil != val {
				v := reflect.ValueOf(val)
				if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
					val = string(val.([]uint8))
				}
			}
			resMap[cols[i]] = val
		}

		res = append(res, resMap)
	}

	return res, nil
}

func rowsToSlices(rows *sql.Rows) ([]interface{}, error) {
	res := make([]interface{}, 0)
	for rows.Next() {
		resSlice := make([]interface{}, 0)
		cols, err := rows.Columns()
		if nil != err {
			log.Println(err)
			return nil, err
		}

		vals := make([]interface{}, len(cols))
		scanArgs := make([]interface{}, len(cols))
		for i := range vals {
			scanArgs[i] = &vals[i]
		}

		rows.Scan(scanArgs...)
		for i := 0; i < len(cols); i++ {
			val := vals[i]
			if nil != val {
				v := reflect.ValueOf(val)
				if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
					val = string(val.([]uint8))
				}
			}
			resSlice = append(resSlice, val)
		}

		res = append(res, resSlice)
	}

	return res, nil
}

func rowsToStructs(rows *sql.Rows, resultType reflect.Type) ([]interface{}, error) {
	fieldsMapper := make(map[string]string)
	fields := resultType.NumField()
	for i := 0; i < fields; i++ {
		field := resultType.Field(i)
		fieldsMapper[field.Name] = field.Name
		tag := field.Tag.Get("field")
		if tag != "" {
			fieldsMapper[tag] = field.Name
		}
	}

	res := make([]interface{}, 0)
	for rows.Next() {
		cols, err := rows.Columns()
		if nil != err {
			log.Fatal(err)
			return nil, err
		}

		vals := make([]interface{}, len(cols))
		scanArgs := make([]interface{}, len(cols))
		for i := range vals {
			scanArgs[i] = &vals[i]
		}

		rows.Scan(scanArgs...)

		obj := reflect.New(resultType).Elem()
		objPtr := reflect.Indirect(obj)
		for i := 0; i < len(cols); i++ {
			colName := cols[i]
			fieldName := fieldsMapper[colName]
			field := objPtr.FieldByName(fieldName)
			//设置相关字段的值,并判断是否可设值
			if field.CanSet() && vals[i] != nil {

				//获取字段类型并设值
				data := dataToFieldVal(vals[i], field.Type())
				if nil != data {
					field.Set(reflect.ValueOf(data))
				}
			}
		}

		if objPtr.CanInterface() {
			res = append(res, objPtr.Addr().Interface())
		}
	}

	return res, nil
}
