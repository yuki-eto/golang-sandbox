package main

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/yuki-eto/golang-sandbox/calc_lib/src"
)

func main() {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		a := r.Int63n(15)
		b := r.Int63n(5)
		wg.Add(1)
		go func() {
			defer wg.Done()
			calc(a, b)
		}()
	}

	wg.Wait()
}

func calc(a, b int64) {
	c := calculator.NewCalculator(int(a), int(b))
	defer calculator.DeleteCalculator(c)

	log.Printf("a: %d, b: %d", a, b)
	log.Printf("sum: %d", c.Sum())
	log.Printf("sub: %d", c.Sub())
	log.Printf("fact: %d", c.Factorial())
}
