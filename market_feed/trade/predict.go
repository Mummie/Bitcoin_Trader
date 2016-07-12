package trade

/*
Draw a square on the ground, then inscribe a circle within it.
Uniformly scatter some objects of uniform size (grains of rice or sand) over the square.
Count the number of objects inside the circle and the total number of objects.
The ratio of the two counts is an estimate of the ratio of the two areas, which is π/4. Multiply the result by 4 to estimate π.*/

/*eliminate survivorship bias
test for statistical significance
make sure your standard deviations aren't too large
back-test in different exchanges
back-test for at least ten years
account for trading commissions/fees
account for liquidity issues
paper trade*/
import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

// Change to
func PI(samples int) float64 {
	var inside int = 0

	for i := 0; i < samples; i++ {
		x := rand.Float64()
		y := rand.Float64()
		if (x*x + y*y) < 1 {
			inside++
		}
	}

	ratio := float64(inside) / float64(samples)

	return ratio * 4
}

func runSimulation() {
	fmt.Println("Our value of Pi after 100 runs:\t\t\t", PI(100))
	fmt.Println("Our value of Pi after 1,000 runs:\t\t", PI(1000))
	fmt.Println("Our value of Pi after 10,000 runs:\t\t", PI(10000))
	fmt.Println("Our value of Pi after 100,000 runs:\t\t", PI(100000))
	fmt.Println("Our value of Pi after 1,000,000 runs:\t\t", PI(1000000))
	fmt.Println("Our value of Pi after 10,000,000 runs:\t\t", PI(10000000))
	fmt.Println("Our value of Pi after 100,000,000 runs:\t\t", PI(100000000))
}

func MultiPI(samples int) float64 {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())

	cpus := runtime.NumCPU()

	threadSamples := samples / cpus
	results := make(chan float64, cpus)

	for j := 0; j < cpus; j++ {
		go func() {
			var inside int
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for i := 0; i < threadSamples; i++ {
				x, y := r.Float64(), r.Float64()

				if x*x+y*y <= 1 {
					inside++
				}
			}
			results <- float64(inside) / float64(threadSamples) * 4
		}()
	}

	var total float64
	for i := 0; i < cpus; i++ {
		total += <-results
	}

	return total / float64(cpus)
}
