package main

import (
	"fmt" ///
	"html/template"
	"image/color"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/stat/distuv"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

//Obtains a Bayesian estimate of the integrated posterior probability P(TC|H) in an interval between TC_L and TC_U (including both),
//given that H is in an interval between HL and HU (including both) and P(H|TC) follows a binomial distribution with TC trials and
//H "successes" and p known. The prior is assumed uniform.
//The original application for this toy program is to estimate the true number of NCOVID-19 in a given country, given the number
//of hospitalized people (H) and the true hospitalization rate (taken from a country with many cases and high detection,
//such as Germany or S. Korea.

//This has a lot of "teaching" comments.
func bayes(par param) (string, error) {
	//fmt.Println("Educational Bayes program")

	//***********The next 4 variabes are for plotting only
	//c is the upper limit for the plot, which we will take as the upper limit for our interval of interest (Pu) plus 5000.
	//cdf will contain the values for CDF at all the plotted range (between 0 and our upper limit + 500)
	//cdbars is similar, but everything outside our range of interest (i.e. not between Pl and Pu) is set to zero. With the way I'm using now to plot, we don't actually need cdbars!
	//realcases is just our X axis
	c := par.Pu + 5000
	cdfbars := make([]float64, 0, c)   //here we'll put the CDF values if we are within the range of interest. For points away from our range, we just put zeros.
	cdf := make([]float64, 0, c)       //the value for all cdfs, whether they are on our range of interest, o not.
	realcases := make([]float64, 0, c) //just the x axis for our later plot.

	//the next var, PH is P(H) integrated over the whole CT space (actually, we just use 0 + c). I am pretty sure that this value is simply
	//the true probability, which is given (par[Tp]), but we calculate it anyway for teaching purposes.
	var PH float64 = 0.0
	var integratedPHCT float64 = 0.0 //Here we accumulate the values for P(H|CT) at all the CT in the integration range (its really a sum, not an integral)

	//PHCT is the binomial distribution (look it up to see what P and N means. We will not set N (number of trials) yet, as it is one of our integrations (sumation) variables.
	PHCT := distuv.Binomial{
		P: par.Tp,
	}

	/*In this loop we go through a large amount of Total cases. We accumuluate the relevant values of P(H|CT), and also the values for P(H).
	We also gather a large range of P(H|CT) values to plot.*/
	for i := 0; i <= par.Pu+5000; i++ {
		if i == 0 {
			PHCT.N = float64(i) + 0.001 //N can't be 0 so I just set it to a small number for the first iteration.
		} else {
			PHCT.N = float64(i)
		}
		tmpcdf := PHCT.CDF(par.Cu) - PHCT.CDF(par.Cl) // CDF(X) is the integral from 0 to X, so CDF(U)-CDF(L) will be the integral from L to U
		cdf = append(cdf, tmpcdf)
		PH = PH + tmpcdf
		cdfbars = append(cdfbars, 0)
		if i >= par.Pl && i <= par.Pu {
			cdfbars[len(cdfbars)-1] = tmpcdf
			integratedPHCT = integratedPHCT + tmpcdf
		}
		realcases = append(realcases, float64(i))
	}

	//we now have all the data!
	//We have our P(H|CT) summation, we just need to divide it by P(H), which we also have, and, assuming a uniform prior, we have our result.
	totalprob := integratedPHCT / PH

	//We now print the answer
	//fmt.Printf("The probability of having P(TC|H) in the range %d-%d, for H between %3.2f-%3.2f, with a true p for P(H|TC) of %4.3f is: %6.4f\n", par.Pl, par.Pu, par.Cl, par.Cu, par.Tp, totalprob)

	//Now the plot.
	name, err := plottingFuncCuatica(realcases, realcases[par.Pl:par.Pu+1], cdf, cdfbars, par.Cl, par.Cu, totalprob)
	if err != nil {
		return "", err
	}
	return name, nil

}

//*******The science ends here, the following are just functions to handle the making of the figure. It doesn't matter for
//*******teaching purposes.

func plottingFuncCuatica(x1, x2, y1, y2 []float64, h1, h2, totalprob float64) (string, error) {
	p, err := plot.New()
	if err != nil {
		return "", err
	}
	p.Title.Text = fmt.Sprintf("Densidad de probabilidad para TC entre 0-%2.1f, dado H %2.1f-%2.1f.\n La integral en el intervalo %2.1f-%2.1f (en rojo) es: %3.2f", x1[len(x1)-1], h1, h2, x2[0], x2[len(x2)-1], totalprob)
	p.X.Label.Text = "TC|H"
	p.Y.Label.Text = "Prob."
	//fmt.Println(y1)     ///////////////////
	// Draw a grid behind the data
	p.Add(plotter.NewGrid())
	linedata := getpoints(x1, y1)
	pointsdata := getpoints(x2, y2)

	l, err := plotter.NewScatter(linedata)
	if err != nil {
		return "", err
	}
	//	l.LineStyle.Width = vg.Points(2)
	l.GlyphStyle.Color = color.RGBA{B: 255, A: 255}
	l.GlyphStyle.Radius = l.GlyphStyle.Radius * 0.3
	s, err := plotter.NewScatter(pointsdata)
	if err != nil {
		return "", err
	}
	s.GlyphStyle.Color = color.RGBA{R: 255, A: 255}

	p.Add(s, l)
	name, err := getName()
	if err != nil {
		return "", err
	}

	//plot to a PNG file
	if err := p.Save(8*vg.Inch, 8*vg.Inch, "./public/"+name); err != nil {
		return "", err
	}
	return "./public/" + name, nil
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

/// This is the webserver, also irrelevant for the science purposes.
type param struct {
	Pl int
	Pu int
	Cl float64
	Cu float64
	Tp float64
}

func main() {
	tmpl := template.Must(template.ParseFiles("formbayes.html"))
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			tmpl.Execute(w, nil)
			return
		}
		var ok bool = true //this is very dirty but I'm not doing the if err!=nil thing  5 times
		details := param{
			Pl: int(errorwrap(r.FormValue("Pl"), &ok)),
			Pu: int(errorwrap(r.FormValue("Pu"), &ok)),
			Cl: errorwrap(r.FormValue("Cl"), &ok),
			Cu: errorwrap(r.FormValue("Cu"), &ok),
			Tp: (errorwrap(r.FormValue("Tp"), &ok)) / 100.0,
		}
		if !ok {
			//	http.Error(w, "Unable to parse input", http.StatusInternalServerError)
			tmpl.Execute(w, struct {
				Success   bool
				ImageName string
			}{Success: false})
		}
		// do something with details
		name, err := bayes(details)
		if err != nil {
			fmt.Println(err.Error()) ////////////////////
			//	http.Error(w, "Unable Build the probability plot", http.StatusInternalServerError)
			tmpl.Execute(w, struct {
				Success   bool
				ImageName string
			}{Success: false})

		}
		data := struct {
			Success   bool
			ImageName string
		}{Success: true, ImageName: name}

		tmpl.Execute(w, data)
	})

	http.ListenAndServe(":8080", nil)
}

func errorwrap(inp string, ok *bool) float64 {
	if *ok == false { //we won't do anything if a previous call failed
		return -1
	}
	i, err := strconv.ParseFloat(inp, 64)
	if err != nil {
		*ok = false
	}
	return i
}

func getName() (string, error) {
	var files []int = make([]int, 1, 100)
	files[0] = 0
	names, err := filepath.Glob("./public/*.png")
	if err != nil {
		return "", err
	}
	for _, v := range names {
		f := strings.Split(v, "/")
		s := strings.Replace(f[len(f)-1], ".png", "", -1)
		num, err := strconv.Atoi(s)
		if err != nil {
			continue //I don't mind other files.
		}
		files = append(files, num)
	}
	sort.Ints(files) //yeaah I don't care
	return fmt.Sprintf("%d.png", files[len(files)-1]+1), nil

}
