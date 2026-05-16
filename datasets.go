package rinhabackend2026

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"log"
	"os"
	"sync"
)

type Dataset struct {
	Vector []float64 `json:"vector"`
	Label  string    `json:"label"`
}

// Esturua para referencia e evita lock
type ReferenceDatasetIndex struct {
	fraudVectors    [][]float64
	nonFraudVectors [][]float64
	mu              sync.RWMutex
}

var globalIndex = &ReferenceDatasetIndex{
	fraudVectors:    make([][]float64, 0, 100000),
	nonFraudVectors: make([][]float64, 0, 100000),
}

func LoadDatasetsInStream(filepath string) error {
	log.Println("Carregando datasets", filepath)

	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzReader, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	scanner := bufio.NewScanner(gzReader)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // Aumenta o buffer

	var inArray = false
	var bracketCount = 0

	// buffer para acumular os dados do array
	var buffer []byte

	for scanner.Scan() {
		line := scanner.Bytes()

		if !inArray {
			for _, b := range line {
				if b == '[' {
					inArray = true
					break
				}
			}
			continue
		}

		buffer = append(buffer, line...)

		//Contando as chaves
		for _, b := range line {
			if b == '{' {
				bracketCount++
			} else if b == '}' {
				bracketCount--

				if bracketCount == 0 {
					var ref Dataset
					if err := json.Unmarshal(buffer, &ref); err != nil {
						addToIndex(&ref)
					}
					buffer = buffer[:0] //clear
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	log.Println("Datasets carregados com sucesso")
	return nil
}

func addToIndex(ref *Dataset) {
	globalIndex.mu.Lock()
	defer globalIndex.mu.Unlock()

	if ref.Label == "fraud" {
		globalIndex.fraudVectors = append(globalIndex.fraudVectors, ref.Vector)
	} else if ref.Label == "legit" {
		globalIndex.nonFraudVectors = append(globalIndex.nonFraudVectors, ref.Vector)
	}
}

func VerifySimilarity(vector []float64, k int) (fraudSimilarity, nonFraudSimilarity float64) {
	globalIndex.mu.RLock()
	defer globalIndex.mu.RUnlock()

	//Busca nos vetores de fraude
	type similarity struct {
		score   float64
		isFraud bool
	}

	results := make([]similarity, 0, k*2)

	for _, ref := range globalIndex.fraudVectors {
		similar := BuscaVetorialSimilaridade(vector, ref)
		results = append(results, similarity{score: similar, isFraud: true})
	}

	//Vetores de não fraude
	for _, ref := range globalIndex.nonFraudVectors {
		similar := BuscaVetorialSimilaridade(vector, ref)
		results = append(results, similarity{score: similar, isFraud: false})
	}

	if len(results) == 0 {
		return 0, 0
	}

	var fraudSum, nonFraudSum float64
	var fraudCount, nonFraudCount int

	limit := k
	if limit > len(results) {
		limit = len(results)
	}

	//tentativade ordenar por similaridade
	for i := 0; i < limit; i++ {
		maxIdx := i
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[maxIdx].score {
				maxIdx = j
			}
		}
		results[i], results[maxIdx] = results[maxIdx], results[i]

		if results[i].isFraud {
			fraudSum += results[i].score
			fraudCount++
		} else {
			nonFraudSum += results[i].score
			nonFraudCount++
		}
	}

	if fraudCount > 0 {
		fraudSimilarity = fraudSum / float64(fraudCount)
	}

	if nonFraudCount > 0 {
		nonFraudSimilarity = nonFraudSum / float64(nonFraudCount)
	}

	return fraudSimilarity, nonFraudSimilarity
}

func GetIndexStats() (fraudCount, legitCount int) {
	globalIndex.mu.RLock()
	defer globalIndex.mu.RUnlock()
	return len(globalIndex.fraudVectors), len(globalIndex.nonFraudVectors)
}
