// Package main provides a high-throughput synthetic data generator CLI script
// that outputs thousands or millions of structured automotive catalog, vehicle fitment,
// and dealer inventory permutations formatted for Vertex AI Search, BigQuery, or Cloud SQL ingestion.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

type PermutationRecord struct {
	ID                     string            `json:"id"`
	EntityType             string            `json:"entity_type"`
	Brand                  string            `json:"brand"`
	Model                  string            `json:"model"`
	ModelYear              int               `json:"model_year"`
	Trim                   string            `json:"trim"`
	PartNumber             string            `json:"part_number"`
	PartName               string            `json:"part_name"`
	Category               string            `json:"category"`
	MSRP                   float64           `json:"msrp"`
	Compatibility          string            `json:"compatibility"`
	DealerPostalCode       string            `json:"dealer_postal_code"`
	QuantityAvailable      int               `json:"quantity_available"`
	SearchKeywords         []string          `json:"search_keywords"`
	Attributes             map[string]string `json:"attributes"`
	IngestionTimestampUTC  string            `json:"ingestion_timestamp_utc"`
}

var (
	brands    = []string{"ApexMotors", "Meridian", "Northstar", "Voltline"}
	models    = []string{"Ridge 1500", "Hauler EV", "LX SUV", "Cross Crossover", "Trailblazer 2500", "Pioneer Hybrid"}
	trims     = []string{"Trail Z71", "LT Deluxe", "Premier Edition", "EV First Edition", "Pro Work Truck"}
	years     = []int{2022, 2023, 2024, 2025, 2026}
	categories = []string{"wheels", "brakes", "lighting", "suspension", "towing", "electrical", "interior_accessories"}
	cities    = []string{"78701", "78702", "78704", "75201", "77002", "90210", "10001"}
)

func main() {
	countFlag := flag.Int("count", 10000, "Number of synthetic permutations to generate")
	outFileFlag := flag.String("out", "var/vertex_search_permutations.jsonl", "Output path for JSONL file")
	flag.Parse()

	log.Printf("Starting generation of %d synthetic automotive discovery permutations...", *countFlag)

	if err := os.MkdirAll("var", 0755); err != nil {
		log.Fatalf("Failed creating var directory: %v", err)
	}

	file, err := os.Create(*outFileFlag)
	if err != nil {
		log.Fatalf("Failed creating output file %s: %v", *outFileFlag, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	start := time.Now()

	for i := 1; i <= *countFlag; i++ {
		brand := brands[rng.Intn(len(brands))]
		model := models[rng.Intn(len(models))]
		trim := trims[rng.Intn(len(trims))]
		year := years[rng.Intn(len(years))]
		category := categories[rng.Intn(len(categories))]
		postal := cities[rng.Intn(len(cities))]

		partNum := fmt.Sprintf("%04d-%04d", rng.Intn(9000)+1000, rng.Intn(9000)+1000)
		partName := fmt.Sprintf("%s %s %s", brand, trim, category)

		record := PermutationRecord{
			ID:                fmt.Sprintf("perm-%08d", i),
			EntityType:        "product_fitment_inventory_aggregate",
			Brand:             brand,
			Model:             model,
			ModelYear:         year,
			Trim:              trim,
			PartNumber:        partNum,
			PartName:          partName,
			Category:          category,
			MSRP:              float64(rng.Intn(800)+50) + 0.99,
			Compatibility:     "direct_fit",
			DealerPostalCode:  postal,
			QuantityAvailable: rng.Intn(50),
			SearchKeywords: []string{
				brand, model, fmt.Sprintf("%d", year), trim, category, partNum,
			},
			Attributes: map[string]string{
				"engine":     "V8 5.3L Dual-Turbo",
				"drivetrain": "4WD",
				"finish":     "Gloss Black",
			},
			IngestionTimestampUTC: time.Now().UTC().Format(time.RFC3339),
		}

		raw, err := json.Marshal(record)
		if err != nil {
			log.Fatalf("Error marshaling record %d: %v", i, err)
		}

		_, _ = writer.Write(raw)
		_, _ = writer.WriteString("\n")

		if i%25000 == 0 {
			_ = writer.Flush()
			log.Printf("  -> Generated %d records...", i)
		}
	}

	_ = writer.Flush()
	elapsed := time.Since(start)

	log.Printf("Completed! Successfully written %d synthetic records to %s in %v.", *countFlag, *outFileFlag, elapsed)
}
