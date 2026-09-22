package dtos

import (
	"fmt"
	"strconv"
)

type AgeRecord struct {
	ID         string      `json:"id"`
	Resultados []AgeResult `json:"resultados"`
}

type AgeResult struct {
	Classificacoes []AgeClassification `json:"classificacoes"`
	Series         []AgeSeries         `json:"series"`
}

type AgeClassification struct {
	ID        string            `json:"id"`
	Nome      string            `json:"nome"`
	Categoria map[string]string `json:"categoria"`
}

type AgeSeries struct {
	Localidade Localidade        `json:"localidade"`
	Serie      map[string]string `json:"serie"`
}

type AgeEntry struct {
	Localidade Localidade
	AgeGroup   string
	Serie      map[string]string
}

func (e AgeEntry) Population(year string) (int64, error) {
	value, ok := e.Serie[year]
	if !ok {
		return 0, fmt.Errorf(
			"population not found for year %s in locality %s",
			year,
			e.Localidade.Nome,
		)
	}

	if value == "-" {
		return 0, nil
	}

	population, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(
			"parse population %q for locality %s: %w",
			value,
			e.Localidade.Nome,
			err,
		)
	}

	return population, nil
}
