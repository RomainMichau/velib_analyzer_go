package main

import (
	"github.com/RomainMichau/velib_finder/clients"
	"log"
	"sync"
	"time"
)

func main() {
	api := clients.InitVelibApi()
	res, _ := api.GetAllStations()
	i := 0
	start := time.Now()
	var wg sync.WaitGroup
	concurrencyLimit := 7
	semaphore := make(chan struct{}, concurrencyLimit)
	for _, v := range res {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(name string) {
			defer wg.Done()
			defer func() {
				<-semaphore
			}()
			resp, err := api.GetVelibAtStations(name)
			if err != nil {
				panic(err)
			}
			if len(resp) == 0 {
				print(resp)
			}
			if i%10 == 0 {
				elapsed := time.Since(start)
				log.Printf("%s", resp[0].Station.Name)
				log.Printf("%d | time:  %s", i, elapsed)
			}
			i++
		}(v.Name)

	}
	wg.Wait()
	log.Printf("time:%s", time.Since(start))
	// time:3m18.13306545s
	print(res)
}
