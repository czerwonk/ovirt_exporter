package statistic

// Statistics is a collection of Statistics
type Statistics struct {
	Statistic []Statistic `xml:"statistic"`
}

// Statistic represents the statistic resource
type Statistic struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
	Kind        string `xml:"kind"`
	Type        string `xml:"type"`
	Unit        string `xml:"unit"`
	Values      struct {
		Value struct {
			Datum float64 `xml:"datum"`
		} `xml:"value"`
	} `xml:"values"`
}
