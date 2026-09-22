package domain

type City struct {
	ID   int64    `db:"id" json:"id"`
	IBGECode int64      `db:"ibge_code" json:"ibge_code"`
	Name string   `db:"name" json:"name"`
	StateID int64  `db:"state_id" json:"state_id"`
}

type CityDetail struct {
	City       City
	State      State
	Population *Indicator[int64]
	Income     *Indicator[float64]
	GDP        *Indicator[float64]
}

type Indicator[T any] struct {
	Year  int
	Value T
}
