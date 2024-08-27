# megafuel-go-sdk

This is Golang SDK for the MegaFuel clients, it provides a simple way to interact with the MegaFuel.

For more information, please refer to the [API documentation](https://docs.nodereal.io/docs/megafuel-api).

## Quick Start
1. Install dependency
 ```shell
 $ go get -u github.com/nodereal/megafuel-go-sdk
 ```
   
2. Example
```go
package main

import (
	"context"
	
	"github.com/node-real/megafuel-go-sdk/pkg/paymasterclient"
)

const (
    SPONSOR_URL = "https://sponsor-api.nodereal.io"
)

func main() {
   sponsorClient, err := paymasterclient.New(context.Background(), SPONSOR_URL)
   if err != nil {
      panic(err)
   }

   success, err := sponsorClient.AddToWhitelist(context.Background(), sponsorclient.WhiteListArgs{
      PolicyUUID:    PolicyUUID,
      WhitelistType: sponsorclient.ToAccountWhitelist,
      Values:        []string{TokenContractAddress.String()},
   })
   if err != nil || !success {
      panic("failed to add token contract whitelist")
   }
   
   println(success)
}

```

3. More examples can be found in the [examples](https://github.com/node-real/megafuel-client-example).

