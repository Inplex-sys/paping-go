package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/fatih/color"
)

type ConnectionStats struct {
	Attempted int
	Connected int
	Failed    int
	MinTime   time.Duration
	MaxTime   time.Duration
	TotalTime time.Duration
}

func ping(host string, port int, stats *ConnectionStats) {
	startTime := time.Now()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Second*5)
	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		stats.Failed++
	} else {
		defer conn.Close()
		duration := time.Since(startTime)
		// fmt.Printf("Connected to %s: time=%.2fms protocol=TCP port=%d\n", host, float64(duration.Microseconds())/1000, port)
		// colors
		fmt.Printf(
			"Connected to "+color.GreenString("%s")+
				": time="+color.GreenString("%.2fms")+
				" protocol="+color.GreenString("TCP")+
				" port="+color.GreenString("%d")+"\n", host, float64(duration.Microseconds())/1000, port)

		stats.Connected++
		stats.TotalTime += duration

		if stats.MinTime == 0 || duration < stats.MinTime {
			stats.MinTime = duration
		}
		if duration > stats.MaxTime {
			stats.MaxTime = duration
		}
	}
	stats.Attempted++
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <host> <port>")
		return
	}

	host := os.Args[1]
	port, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Invalid port number:", err)
		return
	}

	stats := &ConnectionStats{}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		printReport(stats)
		os.Exit(0)
	}()

	for {
		ping(host, port, stats)
		time.Sleep(time.Second)
	}
}

func printReport(stats *ConnectionStats) {
	successRate := float64(stats.Connected) / float64(stats.Attempted) * 100
	fmt.Printf("\nConnection statistics:\n")
	fmt.Printf(
		"  Attempted = "+color.CyanString("%d")+
			", Connected = "+color.CyanString("%d")+
			", Failed = "+color.CyanString("%d")+" ("+color.CyanString("%.2f%%")+")\n", stats.Attempted, stats.Connected, stats.Failed, successRate)
	fmt.Printf("Approximate connection times :\n")
	fmt.Printf(
		"  Minimum = "+color.CyanString("%.2fms")+
			", Maximum = "+color.CyanString("%.2fms")+
			", Average = "+color.CyanString("%.2fms")+"\n", float64(stats.MinTime.Microseconds())/1000, float64(stats.MaxTime.Microseconds())/1000, float64(stats.TotalTime.Microseconds())/float64(stats.Connected)/1000)
}
