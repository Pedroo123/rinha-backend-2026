package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"sync/atomic"
)

type LoadBalancer struct {
	servers []string
	counter uint64
}

func NewLoadBalancer(servers []string) *LoadBalancer {
	return &LoadBalancer{
		servers: servers,
		counter: 0,
	}
}

func (lb *LoadBalancer) GetNextServer() string {
	counter := atomic.AddUint64(&lb.counter, 1)
	return lb.servers[counter%uint64(len(lb.servers))]
}

func main() {
	go startServer(":9998", "Server 1")
	go startServer(":9997", "Server 2")

	loadBalancer := NewLoadBalancer([]string{"http://localhost:9998", "http://localhost:9997"})

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/fraud-score", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		balancedUrl := loadBalancer.GetNextServer() + "/fraud-score"

		body, _ := io.ReadAll(r.Body)
		res, err := http.Post(balancedUrl, "application/json", bytes.NewBuffer(body))

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()

		w.WriteHeader(res.StatusCode)
		io.Copy(w, res.Body)
	})

	log.Println("LB running on :9999")
	log.Fatal(http.ListenAndServe(":9999", nil))
}
