package main

import (
	"fmt" ///
	"image/color"
	"log"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/stat/distuv"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

//Obtains a Bayesian estimate of the integrated posterior probability P(TC|H) in an interval between TC_L and TC_U (including both),
//given that H is in an interval between HL and HU (including both) and P(H|TC) follows a binomial distribution with TC trials and
//H "successes" and p known. The prior is assumed uniform.
//The original application for this toy program is to estimate the true number of NCOVID-19 in a given country, given the number
//of hospitalized people (H) and the true hospitalization rate (taken from a country with many cases and high detection,
//such as Germany or S. Korea.

var par map[string]float64 = map[string]float64{
	"Pl": -1, //Posterior lower limit
	"Pu": -1, // Posterior upper limit (for integration)
	"Cl": -1, //Lower limitabnew  for our
	"Cu": -1, //Cond upper
	"Tp": -1, //True probability for the conditional probability distribution
}

//This has a lot of "teaching" comments.
func main() {
	fmt.Println("Educational Bayes program")
	//First we deal with the user options
	//We will just use a string of data from sys.Args which we parse here. The reason is that
	//later, for the Mobile App we'll let the Android part handle the user input and send us the string here.
	//The order for the numbers will be:
	// TCL TCU  HL HU TrueP

	//***********The next 4 variabes are for plotting only
	//c is the upper limit for the plot, which we will take as the upper limit for our interval of interest (Pu) plus 5000.
	//cdf will contain the values for CDF at all the plotted range (between 0 and our upper limit + 500)
	//cdbars is similar, but everything outside our range of interest (i.e. not between Pl and Pu) is set to zero. With the way I'm using now to plot, we don't actually need cdbars!
	//realcases is just our X axis
	c := int(par["Pu"]) + 5000
	cdfbars := make([]float64, 0, c)   //here we'll put the CDF values if we are within the range of interest. For points away from our range, we just put zeros.
	cdf := make([]float64, 0, c)       //the value for all cdfs, whether they are on our range of interest, o not.
	realcases := make([]float64, 0, c) //just the x axis for our later plot.

	//the next var, PH is P(H) integrated over the whole CT space (actually, we just use 0 + c). I am pretty sure that this value is simply
	//the true probability, which is given (par[Tp]), but we calculate it anyway for teaching purposes.
	var PH float64 = 0.0
	err := numberStringParse(os.Args[1])
	var integratedPHCT float64 = 0.0 //Here we accumulate the values for P(H|CT) at all the CT in the integration range (its really a sum, not an integral)
	if err != nil {
		log.Fatal(err)
	}
	//PHCT is the binomial distribution (look it up to see what P and N means. We will not set N (number of trials) yet, as it is one of our integrations (sumation) variables.
	PHCT := distuv.Binomial{
		P: par["Tp"],
	}

	/*In this loop we go through a large amount of Total cases. We accumuluate the relevant values of P(H|CT), and also the values for P(H).
	We also gather a large range of P(H|CT) values to plot.*/
	for i := 0; i <= int(par["Pu"])+5000; i++ {
		if i == 0 {
			PHCT.N = float64(i) + 0.001 //N can't be 0 so I just set it to a small number for the first iteration.
		} else {
			PHCT.N = float64(i)
		}
		tmpcdf := PHCT.CDF(par["Cu"]) - PHCT.CDF(par["Cl"]) // CDF(X) is the integral from 0 to X, so CDF(U)-CDF(L) will be the integral from L to U
		cdf = append(cdf, tmpcdf)
		PH = PH + tmpcdf
		cdfbars = append(cdfbars, 0)
		if i >= int(par["Pl"]) && i <= int(par["Pu"]) {
			cdfbars[len(cdfbars)-1] = tmpcdf
			integratedPHCT = integratedPHCT + tmpcdf
		}
		realcases = append(realcases, float64(i))
	}

	//we now have all the data!
	//We have our P(H|CT) summation, we just need to divide it by P(H), which we also have, and, assuming a uniform prior, we have our result.
	totalprob := integratedPHCT / PH

	//We now print the answer
	fmt.Printf("The probability of having P(TC|H) in the range %d-%d, for H between %3.2f-%3.2f is: %6.4f\n", int(par["Pl"]), int(par["Pu"]), par["Cl"], par["Cu"], totalprob)

	//Now the plot.
	err = plottingFuncCuatica(realcases, realcases[int(par["Pl"]):int(par["Pu"])+1], cdf, cdfbars, par["Cl"], par["Cu"])
	if err != nil {
		log.Fatal(err)
	}

}

//numberStringParser is the function that take the string given by the client, and transform it into
//the data we need. Note that it ensures that Pl and Pu are actually ints, before transforming them
//to floats, so we can safely transform those back into ints when needed.
func numberStringParse(str string) error {
	var err error
	order := []string{"Pl", "Pu", "Cl", "Cu", "Tp"}
	fields := strings.Fields(str)
	for key, val := range fields {
		if strings.HasPrefix(order[key], "P") {
			tmp, err := strconv.Atoi(val) //we ensure that Pu and Pl are ints.
			if err != nil {
				return err
			}
			par[order[key]] = float64(tmp)
		} else {
			par[order[key]], err = strconv.ParseFloat(val, 64)
		}
		if err != nil {
			return err
		}
	}
	return err
}

//*******The science ends here, the following are just functions to handle the making of the figure. It doesn't matter for
//*******teaching purposes.

func plottingFunc(x1, x2, y1, y2 []float64, h1, h2 float64) error {
	p, err := plot.New()
	if err != nil {
		return err
	}
	p.Title.Text = fmt.Sprintf("Probability for P(TC|H) if H is between %2.1f-%2.1f", h1, h2)
	p.X.Label.Text = "TC|H"
	p.Y.Label.Text = "Prob."
	err = plotutil.AddLinePoints(p, "Studied Area", getpoints(x2, y2))
	if err != nil {
		return err
	}
	err = plotutil.AddLines(p, "Dist", getpoints(x1, y1), getpoints(x2, y2))
	if err != nil {
		return err
	}
	// Save the plot to a PNG file.
	if err := p.Save(7*vg.Inch, 7*vg.Inch, "distri.png"); err != nil {
		return err
	}
	return nil
}

func plottingFuncCuatica(x1, x2, y1, y2 []float64, h1, h2 float64) error {
	p, err := plot.New()
	if err != nil {
		return err
	}
	p.Title.Text = fmt.Sprintf("Probability for P(TC|H) if H is between %2.1f-%2.1f", h1, h2)
	p.X.Label.Text = "TC|H"
	p.Y.Label.Text = "Prob."
	//fmt.Println(y1)     ///////////////////
	// Draw a grid behind the data
	p.Add(plotter.NewGrid())
	linedata := getpoints(x1, y1)
	pointsdata := getpoints(x2, y2)

	l, err := plotter.NewScatter(linedata)
	if err != nil {
		return err
	}
	//	l.LineStyle.Width = vg.Points(2)
	l.GlyphStyle.Color = color.RGBA{B: 255, A: 255}
	l.GlyphStyle.Radius = l.GlyphStyle.Radius * 0.3
	s, err := plotter.NewScatter(pointsdata)
	if err != nil {
		return err
	}
	s.GlyphStyle.Color = color.RGBA{R: 255, A: 255}

	p.Add(s, l)
	//plot to a PNG file.
	if err := p.Save(7*vg.Inch, 7*vg.Inch, "distri.png"); err != nil {
		return err
	}
	return nil
}

// randomPoints returns some random x, y points.
func getpoints(X, Y []float64) plotter.XYs {
	//fmt.Println(len(X), len(Y)) /////////////////////////
	pts := make(plotter.XYs, len(X))
	for i := range pts {
		pts[i].X = X[i]
		pts[i].Y = Y[i]
	}
	return pts
}
