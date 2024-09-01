# influxx

Influx helper

## Install

```shell
go get github.com/innotechdevops/influxx@v1.0.3
```

## How to use

- Define struct

```go
type MinewSensor struct {
    Timestamp    time.Time `influxdb:"Timestamp"`
    Temperature  float64   `influxdb:"Temperature"`
    Humidity     float64   `influxdb:"Humidity"`
    Battery      float64   `influxdb:"Battery"`
    RSSI         float64   `influxdb:"RSSI"`
    Code         string    `influxdb:"Code"`
}
```

- Try parse value to struct (Recommended)

```go
var values []influxdb1.Result
dataStruct := influxx.TryParser[MinewSensor](values, func(element []any) MinewSensor {
    return MinewSensor {
        Timestamp:   time.Unix(influxx.GetInt64(element[0]), 0),
        Temperature: influxx.GetFloat64(element[1]),
        Humidity:    influxx.GetFloat64(element[2]),
        Battery:     influxx.GetFloat64(element[3]),
        RSSI:        influxx.GetFloat64(element[4]),
        Code:        influxx.GetString(element[5]),
    }
})
```

- Parse value to struct

```go
query := `
    SELECT
        time,
        temperature,
        humidity,
        battery,
        rssi,
        code
    FROM (
        SELECT
            mean(temperature) as temperature,
            mean(humidity) as humidity,
            mean(battery) as battery,
            mean(rssi) as rssi
        FROM minew_sensor_indoor
        WHERE code =~ /A|B/ AND time >= now() and time <= now()
        GROUP BY time(15m), code fill(previous) tz('Asia/Bangkok')
    )
`

values := [][]any{{"Timestamp", "Temperature", "Humidity", "Battery", "RSSI", "Code"}}
values = append(values, response.Results[0].Series[0].Values...)
sensors := influxx.Parser[MinewSensor](values)
```

- Try Mapping value null-safety

```go
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
```

- Convert struct to influx pattern

```go
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
```

- New points by struct

```go
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
```

## Benchmark

```shell
goos: darwin
goarch: arm64
pkg: github.com/innotechdevops/influxx
cpu: Apple M1 Pro
Benchmark_ManualMapping
Benchmark_ManualMapping-10         	13121960	        78.55 ns/op
Benchmark_TryMapping
Benchmark_TryMapping-10            	 5757108	       204.1 ns/op
Benchmark_TagAndFieldMapping
Benchmark_TagAndFieldMapping-10    	 8432767	       141.7 ns/op
Benchmark_FieldMapping
Benchmark_FieldMapping-10          	 8439260	       141.5 ns/op
Benchmark_ManualParser
Benchmark_ManualParser-10          	 2062294	       580.3 ns/op
Benchmark_TryParser
Benchmark_TryParser-10             	 1949780	       609.9 ns/op
Benchmark_Parser
Benchmark_Parser-10                	   43351	     27374 ns/op
PASS
```