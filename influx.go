package influxx

import (
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	influxdb1 "github.com/influxdata/influxdb1-client/v2"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	minScale     = "15m"
	FnRaw        = "raw"
	FnMean       = "mean"
	FnLast       = "last"
	FillPrevious = "previous"
	FillLinear   = "linear"
	FillNone     = "none"
	FillZero     = "0"
)

type Params struct {
	Start int64  `json:"start"`
	End   int64  `json:"end"`
	Scale string `json:"scale"`
	Last  string `json:"last"`
	Fill  string `json:"fill"`
	Fn    string `json:"fn"`
}

type Query struct {
	Start string `json:"start"`
	Fn    string `json:"fn"`
	End   string `json:"end"`
	Group string `json:"group"`
	Fill  string `json:"fill"`
}

// InRegex example: code =~ /A|B|C/
func InRegex(list []string) string {
	if len(list) == 1 {
		return "/" + list[0] + "/"
	}
	pattern := strings.Join(list, "|")
	return fmt.Sprintf("/%s/", pattern)
}

// InOr example: code = A OR code = B OR code = C
func InOr(where string, list []string) string {
	// Create an empty list to store the WHERE conditions
	whereConditions := []string{}

	// Loop through each element in the string list
	for _, element := range list {
		// Add a condition to the list with single quotes around the element
		whereConditions = append(whereConditions, fmt.Sprintf("%s = '%s'", where, element))
	}

	// Join the conditions with OR operator
	return strings.Join(whereConditions, " OR ")
}

func TimeRangeConvert(data Params) Query {
	qConf := Query{}
	qConf.Fill = setFill(data.Fill)
	qConf.Fn = setFunction(data.Fn)
	if data.Last == "" && data.Start != 0 {
		qConf.Start = strconv.Itoa(int(data.Start)) + "s"
		qConf.End = strconv.Itoa(int(data.End)) + "s"
		if data.Scale == "" {
			qConf.Group = minScale
		} else {
			qConf.Group = data.Scale
		}
	} else {
		if data.Start == 0 {
			qConf.Start = fmt.Sprintf("now() - %s", data.Last)
			qConf.End = "now()"
		} else {
			qConf.Start = fmt.Sprintf("%ss - %s", strconv.Itoa(int(data.Start)), data.Last)
			qConf.End = fmt.Sprintf("%ss", strconv.Itoa(int(data.Start)))
		}
		if data.Scale == "" {
			qConf.Group = setScale(data.Last)
		} else {
			qConf.Group = data.Scale
		}
	}
	return qConf
}

func setFunction(data string) string {
	switch data {
	case "mean", "avg":
		return "mean"
	case "last":
		return "last"
	default:
		return "mean"
	}
}

func setFill(data string) string {
	if data == "" {
		return "previous"
	} else {
		return data
	}
}

func setScale(data string) string {
	switch data {
	case "1h":
		return "15m"
	case "3h":
		return "15m"
	case "6h":
		return "15m"
	case "12h":
		return "15m"
	case "24h", "1d":
		return "1h"
	case "2d":
		return "2h"
	case "7d":
		return "12h"
	case "30d":
		return "1d"
	case "90d":
		return "1w"
	}
	return minScale
}

type model[T any] struct {
	Data T
}

// Parser data to Struct
// How to use
// query := `
//
//	 SELECT
//			time,
//		    temperature,
//		    humidity,
//		    battery,
//		    rssi,
//		    code
//		FROM (
//		    SELECT
//				mean(temperature) as temperature,
//				mean(humidity) as humidity,
//				mean(battery) as battery,
//				mean(rssi) as rssi
//			FROM minew_sensor_indoor
//			WHERE code =~ /A|B/ AND time >= now() and time <= now()
//			GROUP BY time(15m), code fill(previous) tz('Asia/Bangkok')
//		)`
//
// values := [][]any{{"Timestamp", "Temperature", "Humidity", "Battery", "RSSI", "Code"}}
// values = append(values, response.Results[0].Series[0].Values...)
// sensors = influxx.Parser[MinewSensor](values)
func Parser[T any](rows [][]any) []T {
	var structs []T

	if len(rows) == 0 {
		return structs
	}

	header := rows[0]
	for i, row := range rows {
		if i == 0 {
			continue
		}

		record := model[T]{}
		structValue := reflect.ValueOf(&record.Data).Elem()

		for j, field := range row {
			structField := structValue.FieldByNameFunc(func(fieldName string) bool {
				f, _ := reflect.TypeOf(record.Data).FieldByName(fieldName)
				fieldTag := f.Tag.Get("influxdb")
				head := header[j]
				return fieldTag == fmt.Sprintf("%v", head)
			})

			if structField.IsValid() {
				if field == nil {
					continue
				}

				typ := reflect.TypeOf(field)
				switch typ.Kind() {
				case reflect.String:
					switch field.(type) {
					case json.Number:
						value := field.(json.Number)
						switch structField.Kind() {
						case reflect.String:
							structField.SetString(value.String())
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
							vInt, _ := value.Int64()
							structField.SetInt(vInt)
						case reflect.Float32, reflect.Float64:
							vFloat, _ := value.Float64()
							structField.SetFloat(vFloat)
						default:
						}
					case string:
						structField.SetString(field.(string))
					}
				default:
				}
			}
		}

		structs = append(structs, record.Data)
	}

	return structs
}

func TryParser[T any](results []influxdb1.Result, onCompute func(element []any) T) []T {
	var structs []T
	for _, result := range results {
		for _, series := range result.Series {
			for _, element := range series.Values {
				structs = append(structs, onCompute(element))
			}
		}
	}
	return structs
}

func Convert[T any](data []T, onCompute func(timestamp time.Time, tags map[string]string, fields map[string]any)) error {
	if len(data) == 0 {
		return errors.New("data is empty")
	}

	for _, row := range data {

		timestamp := int64(0)
		tags := make(map[string]string)
		fields := make(map[string]any)

		t := reflect.TypeOf(row)
		v := reflect.ValueOf(row)
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)

			if _, ok1 := field.Tag.Lookup("influxtime"); ok1 {
				if value.Kind() == reflect.Ptr && value.IsNil() {
					continue
				}
				// Safely check if the value is a int
				if intValue, ok2 := value.Interface().(int); ok2 {
					timestamp = int64(intValue)
				} else if int32Value, ok3 := value.Interface().(int32); ok3 {
					timestamp = int64(int32Value)
				} else if int64Value, ok4 := value.Interface().(int64); ok4 {
					timestamp = int64Value
				}
			}

			if name, ok1 := field.Tag.Lookup("influxtag"); ok1 {
				if value.Kind() == reflect.Ptr {
					if value.IsNil() {
						continue
					}
					value = reflect.Indirect(value)
				}

				// Safely check if the value is a string
				if strValue, ok2 := value.Interface().(string); ok2 {
					tags[name] = strValue
				}
			}

			if name, ok := field.Tag.Lookup("influxfield"); ok {
				if value.Kind() == reflect.Ptr && value.IsNil() {
					continue
				}
				fields[name] = reflect.Indirect(value).Interface()
			}
		}

		onCompute(time.Unix(timestamp, 0), tags, fields)
	}

	return nil
}

func NewPoint[T any](data []T, name string, batchPoint influxdb1.BatchPoints) error {
	return Convert(data, func(timestamp time.Time, tags map[string]string, fields map[string]any) {
		pt, err := influxdb1.NewPoint(name, tags, fields, timestamp)
		if err != nil {
			fmt.Println("Error:", err.Error())
		}
		batchPoint.AddPoint(pt)
	})
}

func TryMapping[V any, M any](key string, value V, mapping map[string]M) {
	v := reflect.ValueOf(value)

	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}

		v = v.Elem()
	}

	// Check if the value is a string and empty
	if v.Kind() == reflect.String && v.String() == "" {
		return
	}

	if vValue, ok := v.Interface().(M); ok {
		mapping[key] = vValue
	}
}

func FieldMapping[T any](key string, value *T, mapping map[string]any) {
	if value != nil {
		mapping[key] = *value
	}
}

func TagMapping(key string, value string, mapping map[string]string) {
	if value != "" {
		mapping[key] = value
	}
}

func SafetyMapping[T any](value *T, onCompute func(value T)) {
	if value != nil {
		onCompute(*value)
	}
}

func GetInt64(value any) int64 {
	if value == nil {
		return 0
	}
	data, _ := value.(json.Number).Int64()
	return data
}

func GetFloat64(value any, decimal ...int) float64 {
	if value == nil {
		return 0
	}
	data, _ := value.(json.Number).Float64()
	if len(decimal) > 0 && decimal[0] > 0 {
		data = FloatDecimal(data, decimal[0])
	}
	return data
}

func GetString(value any) string {
	if value == nil {
		return ""
	}
	switch value.(type) {
	case json.Number:
		return value.(json.Number).String()
	default:
		return fmt.Sprintf("%v", value)
	}
}

func GetTimeString(timestamp int64) string {
	return time.Unix(timestamp, 0).String()
}

func GetTime(value any) time.Time {
	if timestamp, ok := value.(int64); ok {
		return time.Unix(timestamp, 0)
	}
	return time.Unix(0, 0)
}

func FloatDecimal(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(Round(num*output)) / output
}

func Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func AnyToPointer[T any](data T) *T {
	return &data
}
