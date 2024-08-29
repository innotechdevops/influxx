# influxx

Influx helper

## Install

```shell
go get github.com/innotechdevops/influxx
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
dataStruct := influxx.TryParser[MinewSensor](values, func(element []MinewSensor) Struct {
    return MinewSensor {
        Timestamp:   element[0],
        Temperature: element[1],
        Humidity:    element[2],
        Battery:     element[3],
        RSSI:        element[4],
        Code:        element[5],
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