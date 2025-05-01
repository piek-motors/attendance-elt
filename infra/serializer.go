package infra

import (
	"encoding/csv"
	"log"
	"strings"
)

type ParserCallback[C any] func(record []string, fieldIndex map[string]int) (C, error)

func SerializeCSVInput[C any](input string, cb ParserCallback[C]) (result []C, err error) {
	data, err := csv.NewReader(strings.NewReader(input)).ReadAll()
	if err != nil {
		log.Println("error parsing csv: ", err)
		return nil, err
	}

	columnNamesIndex := make(map[string]int)
	for i, field := range data[0] {
		columnNamesIndex[field] = i
	}

	for _, line := range data[1:] {
		element, err := cb(line, columnNamesIndex)
		if err != nil {
			log.Println("parsing error:", err)
			continue
		}
		result = append(result, element)
	}

	return result, nil
}
