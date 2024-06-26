package main

import (
	"fmt"
	"log"
	"math"
	"net"
	"runtime"
	"sync"
	"time"
)

var MaxTasks = runtime.NumCPU() * runtime.NumCPU()

type OutputLog struct {
	wg sync.WaitGroup
}

func main() {
	var host string = "cht.sh"
	var portStart uint16 = 80
	var portEnd uint16 = 443

	if host == "" || portStart == 0 || portEnd == 0 {
		log.Fatal("arguments is missing")
	}

	if portStart > portEnd {
		log.Fatal("portStart > portEnd")
	}

	outLog := &OutputLog{
		wg: sync.WaitGroup{},
	}
	start := time.Now()
	res, err := scanHost(outLog, host, portStart, portEnd)
	if err != nil {
		log.Fatal(err)
	}
	outLog.wg.Wait()

	elapsed := time.Since(start)

	log.Printf("Time: %s | Max tasks:  %d", elapsed.String(), MaxTasks)
	log.Println("------------------")
	for i := 0; i < len(res); i++ {
		log.Printf("port: %d is open", res[i])
	}
}
func scanHost(outLog *OutputLog, host string, startPort, endPort uint16) ([]uint16, error) {
	sizeTasks := endPort - startPort + 1
	listPorts := make([]uint16, 0, sizeTasks)

	wg := sync.WaitGroup{}
	wg.Add(int(sizeTasks))

	ch := make(chan struct{}, MaxTasks)
	for i := startPort; i <= endPort; i++ {
		ch <- struct{}{}
		go func(port uint16) {
			defer wg.Done()
			outLog.wg.Add(1)
			err := portScan(port, host)
			if err == nil {
				go outLog.Add(fmt.Sprintf("port: %d is open", port))
				listPorts = append(listPorts, port)
			} else {
				go outLog.Add(fmt.Sprintf("port: %d scan failed: %s", port, err))
			}
			<-ch
		}(i)

		if i == math.MaxUint16 {
			break
		}
	}
	wg.Wait()
	return listPorts[0:len(listPorts):len(listPorts)], nil
}
func portScan(port uint16, host string) error {
	connect, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 10*time.Second)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer connect.Close()
	return nil
}

func (t *OutputLog) Add(msg string) {
	log.Println(msg)
	t.wg.Done()
}
