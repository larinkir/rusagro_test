package main

import (
	"errors"
	"fmt"
	"log"
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
	var host string = "ya.ru"
	var portStart uint16 = 1
	var portEnd uint16 = 1000

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

	log.Printf("Test_Stress_Failed_Connect time: %s | Total tasks:  %d", elapsed.String(), MaxTasks)
	log.Println("------------------")
	for i := 0; i < len(res); i++ {
		log.Printf("port: %d is open", res[i])
	}
}
func scanHost(outLog *OutputLog, host string, startPort, endPort uint16) ([]uint16, error) {
	sizeTasks := endPort - startPort + 1
	listPorts := make([]uint16, 0, sizeTasks)
	if host == "" || startPort == 0 || endPort == 0 {
		return listPorts, errors.New("arguments is missing")
	} else if startPort > endPort {
		return listPorts, errors.New("invalid arguments")
	}
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
				log.Println(fmt.Sprintf("port: %d is open", port))
				listPorts = append(listPorts, port)
			} else {
				go outLog.Add(fmt.Sprintf("port: %d scan failed: %s", port, err))

			}
			<-ch
		}(i)
	}
	wg.Wait()
	return listPorts[0:len(listPorts):len(listPorts)], nil
}
func portScan(port uint16, host string) error {
	connect, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), 20*time.Millisecond)
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
