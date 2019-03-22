package main

import (
	"fmt"
	"look4you/loadbalancer"
)

func main() {
	lb := loadbalancer.Lb{
		Endpoints: []string{"112.74.200.115:82", "112.74.200.115:81"},
	}

	lb.Init()

	for i := 0; i < 5; i++ {
		if i == 3 {
			lb.Close()
		}

		ep, err := lb.GetEndpoint()
		if err != nil {
			fmt.Println("err", err.Error())
		} else {
			fmt.Println("ep", ep)
		}
	}
}
