package dtos

import "github.com/CunhazadanoDale/trads-market-test/internal/core/domain"

type NationalMetricsResponse struct {
	Municipalities int64              `json:"municipios"`
	Indicators     IndicatorsResponse `json:"indicators"`
}

func NewNationalMetricsResponse(metrics domain.NationalMetrics) NationalMetricsResponse {
	return NationalMetricsResponse{
		Municipalities: metrics.Municipalities,
		Indicators: IndicatorsResponse{
			Population: newIndicatorResponse(metrics.Population),
			Income:     newIndicatorResponse(metrics.Income),
			GDP:        newIndicatorResponse(metrics.GDP),
		},
	}
}

type StateMetricsResponse struct {
	StateResponse
	Municipalities int64              `json:"municipios"`
	Indicators     IndicatorsResponse `json:"indicators"`
}

func NewStateMetricsResponse(metrics domain.StateMetrics) StateMetricsResponse {
	return StateMetricsResponse{
		StateResponse:  NewStateResponse(metrics.State),
		Municipalities: metrics.Municipalities,
		Indicators: IndicatorsResponse{
			Population: newIndicatorResponse(metrics.Population),
			Income:     newIndicatorResponse(metrics.Income),
			GDP:        newIndicatorResponse(metrics.GDP),
		},
	}
}
