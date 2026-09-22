package domain

type NationalMetrics struct {
	Municipalities int64
	Population     *Indicator[int64]
	Income         *Indicator[float64]
	GDP            *Indicator[float64]
}
