package dtos

import "github.com/CunhazadanoDale/trads-market-test/internal/core/domain"

type CityResponse struct {
	ID         int64              `json:"id"`
	IBGECode   int64              `json:"ibge_code"`
	Name       string             `json:"name"`
	Indicators IndicatorsResponse `json:"indicators"`
}

func NewCityResponse(city domain.CityWithIndicators) CityResponse {
	return CityResponse{
		ID:       city.ID,
		IBGECode: city.IBGECode,
		Name:     city.Name,
		Indicators: IndicatorsResponse{
			Population: newIndicatorResponse(city.Population),
			Income:     newIndicatorResponse(city.Income),
			GDP:        newIndicatorResponse(city.GDP),
		},
	}
}

type CityDetailResponse struct {
	ID         int64              `json:"id"`
	IBGECode   int64              `json:"ibge_code"`
	Name       string             `json:"name"`
	State      StateResponse      `json:"state"`
	Indicators IndicatorsResponse `json:"indicators"`
}

type IndicatorsResponse struct {
	Population *IndicatorResponse[int64]   `json:"population"`
	Income     *IndicatorResponse[float64] `json:"income"`
	GDP        *IndicatorResponse[float64] `json:"gdp"`
}

type IndicatorResponse[T any] struct {
	Year  int `json:"year"`
	Value T   `json:"value"`
}

func NewCityDetailResponse(detail domain.CityDetail) CityDetailResponse {
	return CityDetailResponse{
		ID:       detail.City.ID,
		IBGECode: detail.City.IBGECode,
		Name:     detail.City.Name,
		State:    NewStateResponse(detail.State),
		Indicators: IndicatorsResponse{
			Population: newIndicatorResponse(detail.Population),
			Income:     newIndicatorResponse(detail.Income),
			GDP:        newIndicatorResponse(detail.GDP),
		},
	}
}

func newIndicatorResponse[T any](indicator *domain.Indicator[T]) *IndicatorResponse[T] {
	if indicator == nil {
		return nil
	}

	return &IndicatorResponse[T]{
		Year:  indicator.Year,
		Value: indicator.Value,
	}
}
