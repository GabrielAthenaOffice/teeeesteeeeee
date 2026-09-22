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

type TopCitiesResponse struct {
	TopGDP        []CityDetailResponse `json:"top_pib"`
	TopIncome     []CityDetailResponse `json:"top_renda"`
	TopPopulation []CityDetailResponse `json:"top_populacao"`
}

func NewTopCitiesResponse(top domain.TopCities) TopCitiesResponse {
	return TopCitiesResponse{
		TopGDP:        newCityDetailResponses(top.TopGDP),
		TopIncome:     newCityDetailResponses(top.TopIncome),
		TopPopulation: newCityDetailResponses(top.TopPopulation),
	}
}

type AgeGroupResponse struct {
	AgeGroup   string `json:"faixa"`
	Population int64  `json:"populacao"`
}

type AgeDistributionResponse struct {
	Year   int                `json:"ano"`
	Total  int64              `json:"total"`
	Groups []AgeGroupResponse `json:"grupos"`
}

func NewAgeDistributionResponse(
	distribution domain.AgeDistribution,
) AgeDistributionResponse {
	groups := make([]AgeGroupResponse, 0, len(distribution.Groups))

	for _, group := range distribution.Groups {
		groups = append(groups, AgeGroupResponse{
			AgeGroup:   group.AgeGroup,
			Population: group.Population,
		})
	}

	return AgeDistributionResponse{
		Year:   distribution.Year,
		Total:  distribution.Total,
		Groups: groups,
	}
}

func newCityDetailResponses(cities []domain.CityDetail) []CityDetailResponse {
	responses := make([]CityDetailResponse, 0, len(cities))

	for _, city := range cities {
		responses = append(responses, NewCityDetailResponse(city))
	}

	return responses
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
