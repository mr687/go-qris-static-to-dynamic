# go-qris-static-to-dynamic

Convert a static QRIS payload into a dynamic one by updating the QRIS TLV fields and recalculating the CRC checksum.

## Features

- Parse QRIS TLV payloads.
- Convert static QRIS payloads to dynamic payloads.
- Update transaction amount and optional tip fields.
- Validate checksum before conversion.

## Requirements

- Go 1.25 or newer.

## Install

To add it to your Go module:

```bash
go get github.com/mr687/go-qris-static-to-dynamic
```

## Usage

```go
package main

import "fmt"

func main() {
    staticQRIS := "00020101021126570011I..."

    dynamicQRIS, err := ToDynamic(staticQRIS, DynamicOption{Amount: 50000})
    if err != nil {
        panic(err)
    }

    fmt.Println(dynamicQRIS)
}
```

### Example with tip

```go
package main

import "fmt"

func main() {
    staticQRIS := "00020101021126570011I..."

    dynamicQRIS, err := ToDynamic(staticQRIS, DynamicOption{
        Amount: 25000,
        Tip: &DynamicTipOption{
            TipType:   TipFixed,
            TipAmount: 2000,
        },
    })
    if err != nil {
        panic(err)
    }

    fmt.Println(dynamicQRIS)
}
```

## Testing

```bash
go test ./...
```

## Benchmarking

```bash
go test -bench . -benchmem ./...
```
