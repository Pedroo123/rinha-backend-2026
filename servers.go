package rinhabackend2026

import (
	"encoding/json"
	"log"
	"net/http"
)

type Request struct {
	Data []float64 `json:"data"`
}

type Response struct {
	Approved   bool    `json:"approved"`
	FraudScore float64 `json:"fraud_score"`
}

func startServer(port string) {
	db := InitDB()
	defer db.Close()

	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/fraud-score", func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		maxSimilaridade := 0.0
		for _, ds := range datasets {
			similaridades := GetSimilaridade(req.Data, ds.Vector)
			if similaridades > maxSimilaridade {
				maxSimilaridade = similaridades
			}
		}

		fraudScore := maxSimilaridade * 100
		approved := fraudScore < 70.0

		response := Response{
			Approved:   approved,
			FraudScore: fraudScore,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	log.Printf("Server started on %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
