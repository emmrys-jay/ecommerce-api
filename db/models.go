package db

type Feature struct {
	F string `json:"feature"`
}

type Product struct {
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Currency    string    `json:"currency"`
	Quantity    int64     `json:"quantity"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Features    []Feature `json:"features"`
	Reviews     []Review  `json:"reviews"`
}

type Review struct {
	User    string `json:"user"`
	Stars   int8   `json:"stars"`
	Comment string `json:"comment"`
}
