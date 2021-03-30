package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/tendermint/tendermint/libs/os"
	"github.com/tendermint/tendermint/rpc/jsonrpc/client"
	"time"
)

func main() {
	wsc,err := client.NewWS("127.0.0.1:26657", "/websocket")
	err = wsc.Start()
	if err != nil {
		os.Exit(err.Error())
	}
	defer wsc.Stop()

	// Read a bunch of responses
	go func() {
		for {
			_, ok := <-wsc.ResponsesCh
			if !ok {
				break
			}
			//fmt.Println("Received response", string(wire.JSONBytes(res)))
		}
	}()

	// Make a bunch of requests
	buf := make([]byte, 32)
	for i := 0; ; i++ {
		binary.BigEndian.PutUint64(buf, uint64(i))
		//txBytes := hex.EncodeToString(buf[:n])
		fmt.Print(".")
		err = wsc.Call(context.TODO(), "broadcast_tx", map[string]interface{}{"tx": buf[:8]})
		if err != nil {
			os.Exit(err.Error())
		}
		if i%1000 == 0 {
			fmt.Println(i)
		}
		time.Sleep(time.Microsecond * 1000)
	}
}
