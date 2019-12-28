package main

import (
  "encoding/json"
  "fmt"
  "math/rand"
  "time"
)

type sandwich struct {
  Bread   string
  Mayo    bool
  Veggies []string
  Meat    []string
  Cheese  []string
}

func main() {
  breads := []string{"White", "Wheat", "Rye", "Texas Toast"}
  veg := []string{"Lettuce", "Tomato", "Onion", "Avocado"}
  meat := []string{"Ham", "Salami", "Turkey", "Bacon"}
  cheese := []string{"Cheddar", "Munster", "Provolone", "Mozzarella"}

  rand.Seed(time.Now().Unix())

  numVeg := rand.Intn(len(veg))
  numMeat := rand.Intn(len(meat))
  numCheese := rand.Intn(len(cheese))

  var veggies, meats, cheeses []string

  for i := 0; i <= numVeg; i++ {
    veggies = append(veggies, veg[rand.Intn(len(veg))])
  }

  for i := 0; i <= numMeat; i++ {
    meats = append(meats, meat[rand.Intn(len(meat))])
  }

  for i := 0; i <= numCheese; i++ {
    cheeses = append(cheeses, cheese[rand.Intn(len(cheese))])
  }

  wich, _ := json.Marshal(sandwich{
    Bread:   breads[rand.Intn(len(breads))],
    Mayo:    rand.Intn(1) == 1,
    Veggies: veggies,
    Meat:    meats,
    Cheese:  cheeses,
  })

  fmt.Printf("%s", wich)
}
