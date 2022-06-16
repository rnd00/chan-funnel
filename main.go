package main

import (
	"log"
	"sync"
	"time"
)

func main() {

	var wg sync.WaitGroup
	dataC := make(chan int)
	printerC := make(chan *data)
	printerQC := make(chan struct{})

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		funnelC := make(chan *data)

		go worker(i, i+1, dataC, funnelC)
		go funnel(i, funnelC, printerC, &wg)
	}

	go printer(printerC, printerQC)

	// insert data
	for n := 1; n < 20; n++ {
		dataC <- n
		time.Sleep(time.Second / 4)
	}
	// data has been inserted, close dataC
	close(dataC)
	// printer should be done as well, close printerC
	close(printerC)

	// wait until everything has done
	log.Println("Waiting")
	wg.Wait()

	// close printerQuitChannel at the end
	printerQC <- struct{}{}
	// close(printerQC)
}

type data struct {
	workertag  int
	former     int
	later      int
	multiplier int
}

func worker(wtag, mult int, dataC <-chan int, funnelC chan<- *data) {
	for {
		num, ok := <-dataC
		if !ok {
			log.Printf("Worker %d closing", wtag)
			close(funnelC)
			break
		}
		laterNum := num * mult
		nData := &data{
			workertag:  wtag,
			former:     num,
			later:      laterNum,
			multiplier: mult,
		}
		funnelC <- nData
	}
}

func funnel(wtag int, funnelC <-chan *data, printerC chan<- *data, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		data, ok := <-funnelC
		if !ok {
			log.Printf("Funnel on worker %d closing", wtag)
			break
		}
		printerC <- data
	}
}

func printer(printerC <-chan *data, printerQC <-chan struct{}) {
	for {
		select {
		case data, ok := <-printerC:
			if ok {
				log.Printf("%+v\n", data)
			}
		case <-printerQC:
			log.Printf("Printer from funnel on worker closing")
			break
		}
	}
}
