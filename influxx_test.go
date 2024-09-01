package influxx_test

import (
	"encoding/json"
	"fmt"
	"github.com/influxdata/influxdb1-client/models"
	influxdb1 "github.com/influxdata/influxdb1-client/v2"
	"github.com/innotechdevops/influxx"
	"testing"
	"time"
)

type MyStruct struct {
	Timestamp int64    `influxtime:"timestamp"`
	ID        string   `influxtag:"id" json:"ID,omitempty"`
	Code      string   `influxtag:"code" json:"code,omitempty"`
	Field1    *float64 `influxfield:"field1" json:"field1,omitempty"`
	Field2    *int     `influxfield:"field2" json:"field2,omitempty"`
	Field3    *int     `influxfield:"field3" json:"field3,omitempty"`
}

type TryMapping struct {
	tag1   string
	tag2   string
	tag3   string
	field1 *float64
	field2 *int
	field3 *int
}

type MinewSensor struct {
	Timestamp   time.Time `influxdb:"Timestamp"`
	Temperature float64   `influxdb:"Temperature"`
	Humidity    float64   `influxdb:"Humidity"`
	Battery     float64   `influxdb:"Battery"`
	RSSI        float64   `influxdb:"RSSI"`
	Code        string    `influxdb:"Code"`
}

func TestConvert(t *testing.T) {
	data := []MyStruct{
		{
			Timestamp: time.Now().Unix(),
			ID:        "1",
			Code:      "C01",
			Field1:    influxx.AnyToPointer(9.9),
			Field2:    influxx.AnyToPointer(10),
			Field3:    nil,
		},
		{
			Timestamp: time.Now().Unix(),
			ID:        "2",
			Code:      "C02",
			Field1:    influxx.AnyToPointer(11.9),
			Field2:    influxx.AnyToPointer(22),
			Field3:    nil,
		},
	}

	_ = influxx.Convert(data, func(timestamp time.Time, tags map[string]string, fields map[string]any) {
		fmt.Println("timestamp:", timestamp)
		fmt.Println("tags:", tags)
		fmt.Println("fields:", fields)
	})
}

func TestNewPoint(t *testing.T) {
	data := []MyStruct{
		{
			Timestamp: time.Now().Unix(),
			ID:        "1",
			Code:      "C01",
			Field1:    influxx.AnyToPointer(9.9),
			Field2:    influxx.AnyToPointer(10),
			Field3:    nil,
		},
		{
			Timestamp: time.Now().Unix(),
			ID:        "2",
			Code:      "C02",
			Field1:    influxx.AnyToPointer(11.9),
			Field2:    influxx.AnyToPointer(22),
			Field3:    nil,
		},
	}

	bp, _ := influxdb1.NewBatchPoints(influxdb1.BatchPointsConfig{
		Database:  "my-database",
		Precision: "s",
	})
	_ = influxx.NewPoint(data, "my_name", bp)
}

func TestTryMapping(t *testing.T) {
	tags := map[string]string{}
	fields := map[string]any{}

	influxx.TryMapping("tag1", "1", tags)
	influxx.TryMapping("tag2", "C001", tags)
	influxx.TryMapping("tag3", "", tags)
	influxx.TryMapping("field1", influxx.AnyToPointer(99.99), fields)
	influxx.TryMapping("field2", 100, fields)
	influxx.TryMapping[*string, any]("field3", nil, fields)

	fmt.Println(tags)   // map[tag1:1 tag2:C001]
	fmt.Println(fields) // map[field1:99.99 field2:100]
}

func Benchmark_ManualMapping(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := TryMapping{"1", "C001", "", influxx.AnyToPointer(99.99), influxx.AnyToPointer(100), nil}

		tags := map[string]string{"tag1": t.tag1, "tag2": t.tag2}
		fields := map[string]any{}

		if t.field1 != nil {
			fields["field1"] = *t.field1
		}
		if t.field2 != nil {
			fields["field2"] = *t.field2
		}
		if t.field3 != nil {
			fields["field3"] = *t.field3
		}

		if len(tags) != 2 {
			b.Error("Error tags is not 2 size")
		}
		if len(fields) != 2 {
			b.Error("Error fields is not 2 size")
		}
	}
}

func Benchmark_TryMapping(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tags := map[string]string{}
		fields := map[string]any{}

		influxx.TryMapping("tag1", "1", tags)
		influxx.TryMapping("tag2", "C001", tags)
		influxx.TryMapping("tag3", "", tags)
		influxx.TryMapping("field1", influxx.AnyToPointer(99.99), fields)
		influxx.TryMapping("field2", 100, fields)
		influxx.TryMapping[*string, any]("field3", nil, fields)

		if len(tags) != 2 {
			b.Error("Error tags is not 2 size")
		}
		if len(fields) != 2 {
			b.Error("Error fields is not 2 size", fields)
		}
	}
}

func Benchmark_TagAndFieldMapping(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := TryMapping{"1", "C001", "", influxx.AnyToPointer(99.99), influxx.AnyToPointer(100), nil}

		tags := map[string]string{}
		fields := map[string]any{}

		influxx.TagMapping("tag1", t.tag1, tags)
		influxx.TagMapping("tag2", t.tag2, tags)
		influxx.TagMapping("tag3", t.tag3, tags)
		influxx.FieldMapping("field1", t.field1, fields)
		influxx.FieldMapping("field2", t.field2, fields)
		influxx.FieldMapping("field3", t.field3, fields)

		if len(tags) != 2 {
			b.Error("Error tags is not 2 size")
		}
		if len(fields) != 2 {
			b.Error("Error fields is not 2 size", fields)
		}
	}
}

func Benchmark_FieldMapping(b *testing.B) {
	for i := 0; i < b.N; i++ {
		t := TryMapping{"1", "C001", "", influxx.AnyToPointer(99.99), influxx.AnyToPointer(100), nil}

		tags := map[string]string{"tag1": t.tag1, "tag2": t.tag2}
		fields := map[string]any{}

		influxx.FieldMapping("field1", t.field1, fields)
		influxx.FieldMapping("field2", t.field2, fields)
		influxx.FieldMapping("field3", t.field3, fields)

		if len(tags) != 2 {
			b.Error("Error tags is not 2 size")
		}
		if len(fields) != 2 {
			b.Error("Error fields is not 2 size", fields)
		}
	}
}

func Benchmark_ManualParser(b *testing.B) {
	results := []influxdb1.Result{
		{
			Series: []models.Row{
				{
					Values: [][]any{
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
					},
				},
			},
		},
	}
	for i := 0; i < b.N; i++ {
		if len(results) > 0 {
			if len(results[0].Series) > 0 {
				data := []MinewSensor{}
				for _, element := range results[0].Series[0].Values {
					data = append(data, MinewSensor{
						Timestamp:   influxx.GetTime(element[0]),
						Temperature: influxx.GetFloat64(element[1]),
						Humidity:    influxx.GetFloat64(element[2]),
						Battery:     influxx.GetFloat64(element[3]),
						RSSI:        influxx.GetFloat64(element[4]),
						Code:        influxx.GetString(element[5]),
					})
				}

				if len(data) != 5 {
					b.Error("Error data is not 5", data)
				}
			}
		}
	}
}

func Benchmark_TryParser(b *testing.B) {
	results := []influxdb1.Result{
		{
			Series: []models.Row{
				{
					Values: [][]any{
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
					},
				},
			},
		},
	}
	for i := 0; i < b.N; i++ {
		data := influxx.TryParser[MinewSensor](results, func(element []any) MinewSensor {
			return MinewSensor{
				Timestamp:   influxx.GetTime(element[0]),
				Temperature: influxx.GetFloat64(element[1]),
				Humidity:    influxx.GetFloat64(element[2]),
				Battery:     influxx.GetFloat64(element[3]),
				RSSI:        influxx.GetFloat64(element[4]),
				Code:        influxx.GetString(element[5]),
			}
		})
		if len(data) != 5 {
			b.Error("Error data is not 5", data)
		}
	}
}

func Benchmark_Parser(b *testing.B) {
	results := []influxdb1.Result{
		{
			Series: []models.Row{
				{
					Values: [][]any{
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
						{time.Now().Unix(), json.Number("1.1"), json.Number("2.2"), json.Number("3.3"), json.Number("4.4"), json.Number("CODE101")},
					},
				},
			},
		},
	}
	values := [][]any{{"Timestamp", "Temperature", "Humidity", "Battery", "RSSI", "Code"}}
	values = append(values, results[0].Series[0].Values...)

	for i := 0; i < b.N; i++ {
		data := influxx.Parser[MinewSensor](values)
		if len(data) != 5 {
			b.Error("Error data is not 5", data)
		}
	}
}
