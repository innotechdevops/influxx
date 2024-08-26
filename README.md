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

- Try parse value to struct

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
