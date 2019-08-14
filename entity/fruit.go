package entity

type Fruit struct {
	FruitName string
	Amount int
}

func (fruit *Fruit) SetFruitName(fruitname string) {
	fruit.FruitName = fruitname
}

func (fruit *Fruit) SetAmount(amount int) {
	fruit.Amount = amount
}

func (fruit *Fruit) GetFruitName() string {
	return fruit.FruitName
}

func (fruit *Fruit) GetAmount() int {
	return fruit.Amount
}

