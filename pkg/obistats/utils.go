package obistats

import (
	log "github.com/sirupsen/logrus"
	"math"
)

// Lchoose returns logarithms of binomial coefficients
func Lchoose(n,x int) float64 {
	fn := float64(n)
	fx := float64(x)

	ln1, _ := math.Lgamma(fn + 1.0)
	lx1, _ := math.Lgamma(fx + 1.0)
	lnx1, _ := math.Lgamma(fn - fx + 1.0)

	return ln1 - lx1 - lnx1
}

func Choose(n,x int) float64 {
	return math.Exp(Lchoose(x,n))
}


func LogAddExp(x, y float64) float64 {

	tmp := x - y
  
	if tmp > 0 {
  
	  return x + math.Log1p(math.Exp(-tmp))
  
	} else if tmp <= 0 {
  
	  return y + math.Log1p(math.Exp(tmp))
  
	} else {
  
	  // Nans, or infinities of the same sign involved
  
	  log.Errorf("logaddexp %f %f", x, y)
  
	  return x + y
  
	}
  
  }
  
  