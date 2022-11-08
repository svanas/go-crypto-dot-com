# go-crypto-dot-com

Go client for the crypto.com API v2 https://exchange-docs.crypto.com

## Installation
```shell
$ go get github.com/svanas/go-crypto-dot-com
```

## Importing
```golang
import (
    exchange "github.com/svanas/go-crypto-dot-com"
)
```

## Setup
```golang
client := exchange.New("your API key", "your API secret")
```
