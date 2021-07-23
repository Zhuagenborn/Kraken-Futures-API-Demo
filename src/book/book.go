package book

type OrderType string

const (
	Ask OrderType = "Ask"
	Bid OrderType = "Bid"
)

type Order struct {
	Price    float64
	Quantity float64
	Type     OrderType
}

type Book struct {
	bids map[float64]float64
	asks map[float64]float64
}

func New(snapshot map[string]interface{}) *Book {
	init := func(key string) map[float64]float64 {
		records := make(map[float64]float64)
		for _, v := range snapshot[key].([]interface{}) {
			item := v.(map[string]interface{})
			records[item["price"].(float64)] = item["qty"].(float64)
		}

		return records
	}

	return &Book{bids: init("bids"), asks: init("asks")}
}

func (book *Book) Update(delta map[string]interface{}) Order {
	var records map[float64]float64
	var order Order
	if delta["side"].(string) == "sell" {
		records = book.asks
		order.Type = Ask
	} else {
		records = book.bids
		order.Type = Bid
	}

	order.Price = delta["price"].(float64)
	order.Quantity = delta["qty"].(float64)
	if order.Quantity != 0 {
		records[order.Price] = order.Quantity
	} else {
		delete(records, order.Price)
	}

	return order
}
