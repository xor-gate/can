// Copyright 2017 can authors. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"github.com/xor-gate/can/socketcan"
)

func main() {
	can, err := socketcan.New("vcan0")
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		f, _ := can.Recv()
		fmt.Printf(" %s %08x   [%d]", can.Name(), f.Id, len(f.Data))
		for n := 0; n < len(f.Data); n++ {
			fmt.Printf(" %02X", f.Data[n])
		}
		fmt.Printf("\n")
	}
}
