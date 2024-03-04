# influxx

Influx helper

## Install

```shell
go get github.com/innotechdevops/influxx
```

## How to use

- Parse value to struct

```go
var values []influxdb1.Result
dataStruct := TryParser[Struct](values, func(element []Struct) Struct {
    return Struct {
        A: element[0],
        B: element[1],
        C: element[2],
    }
})
```