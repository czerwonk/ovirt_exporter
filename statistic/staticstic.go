package statistic

type Statistics struct {
	Statistic []Statistic `xml:"statistic"`
}

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
