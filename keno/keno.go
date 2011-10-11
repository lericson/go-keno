// keno -- lett o vinne
package keno

import (
	"io"
	"fmt"
	"rand"
	"time"
	"http"
)

const (
	// Number of shuffling goroutines to start per juggling
	numRotorBlades = 5
	// How often per second to simulate a blade "hitting" some balls
	bladeHitFrequency = 5
)

type Juggle struct {
	size        int
	ballsInCage chan int
}

func (juggle *Juggle) insertBalls() {
	for i := 0; i < juggle.size; i++ {
		juggle.ballsInCage <- i
	}
}

// One way to simulate some sort of movement due to juggling
func (juggle *Juggle) smackBall() int {
	ball := <-juggle.ballsInCage
	juggle.ballsInCage <- ball
	return ball
}

// Other way to simulate movement, more random-seeming
func (juggle *Juggle) smackAroundBalls() []int {
	var numHits = juggle.size / numRotorBlades
	hits := make([]int, numHits)
	for _, i := range rand.Perm(numHits) {
		select {
		case ball := <-juggle.ballsInCage:
			hits[i] = ball
		default:
			break
		}
	}
	for _, ball := range hits {
		juggle.ballsInCage <- ball
	}
	return hits
}

func (juggle *Juggle) rotate(freq int) {
	for i := 0; i < numRotorBlades; i++ {
		go func() {
			for _ = range time.NewTicker(int64(1e9 / freq)).C {
				juggle.smackAroundBalls()
			}
		}()
	}
}

func (juggle *Juggle) PrintPick(w io.Writer) {
	select {
	case ball := <-juggle.ballsInCage:
		fmt.Fprintf(w, "Ball %v (~%v remain)\n", ball, len(juggle.ballsInCage))
	default:
		fmt.Fprintf(w, "Inserting %v balls\n", juggle.size)
		juggle.insertBalls()
		go juggle.rotate(bladeHitFrequency)
	}
}

type Juggles []*Juggle

func (juggles Juggles) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	for _, juggle := range juggles {
		juggle.PrintPick(w)
	}
}

func init() {
	juggles := Juggles{
		&Juggle{size: 100},
		&Juggle{size: 200},
	}

	for _, juggle := range juggles {
		juggle.ballsInCage = make(chan int, juggle.size)
	}

	http.Handle("/", juggles)
}

// vim: ts=4 noet
