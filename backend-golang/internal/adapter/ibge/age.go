package ibge

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/CunhazadanoDale/trads-market-test/internal/adapter/http/dtos"
)

const ageBatchSize = 5

var ageGroupIDs = []string{
	"93070", "93084", "93085", "93086", "93087",
	"93088", "93089", "93090", "93091", "93092",
	"93093", "93094", "93095", "93096", "93097",
	"93098", "49108", "49109", "60040", "60041",
	"6653",
}

func (c *Client) GetAgeGroups2022(ctx context.Context) ([]dtos.AgeEntry, error) {
	entries := make([]dtos.AgeEntry, 0)

	for start := 0; start < len(ageGroupIDs); start += ageBatchSize {
		end := start + ageBatchSize
		if end > len(ageGroupIDs) {
			end = len(ageGroupIDs)
		}

		batch := ageGroupIDs[start:end]

		query := url.Values{}
		query.Set("localidades", "N6[all]")
		query.Add("classificacao", "287["+strings.Join(batch, ",")+"]")

		var response []dtos.AgeRecord

		err := c.Get(ctx,
			"/agregados/9514/periodos/2022/variaveis/93",
			query,
			&response)
		if err != nil {
			return nil, fmt.Errorf("get age groups batch %d: %w",
				start/ageBatchSize,
				err,
			)
		}

		if len(response) == 0 {
			return nil, fmt.Errorf(
				"IBGE returned empty age response",
			)
		}

		if len(response[0].Resultados) == 0 {
			return nil, fmt.Errorf(
				"IBGE returned no age results",
			)
		}

		for _, result := range response[0].Resultados {
			ageGroup := ""

			for _, class := range result.Classificacoes {
				if class.Nome != "Idade" {
					continue
				}

				for _, nome := range class.Categoria {
					ageGroup = nome
				}
			}

			if ageGroup == "" {
				return nil, fmt.Errorf(
					"age group name missing in IBGE response",
				)
			}

			for _, serie := range result.Series {
				entries = append(entries, dtos.AgeEntry{
					Localidade: serie.Localidade,
					AgeGroup:   ageGroup,
					Serie:      serie.Serie,
				})
			}
		}
	}

	return entries, nil
}
