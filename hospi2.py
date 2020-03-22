#!/usr/bin/env python

#A toy to illustrate a simple aplication of Bayes theorem.
#Copyright (C) 2020  Raul Mera A.

#This program is free software; you can redistribute it and/or
#modify it under the terms of the GNU General Public License version 2,
#as published by the Free Software Foundation.

#This program is distributed in the hope that it will be useful,
#but WITHOUT ANY WARRANTY; without even the implied warranty of
#MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#GNU General Public License for more details.

#You should have received a copy of the GNU General Public License
#along with this program; if not, write to the Free Software
#Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.

import argparse
import getopt
import sys
import scipy.stats
import matplotlib.pyplot as plt
import matplotlib
import numpy as np

cdf=[]
realcases=[]
pmf=[]
totalprob=0
norma=0

print("This is a _TOY_ program that gives a bayesian estimate for the real cases of COVID-19 in the country, given the hospitalized people. It's only meant for teaching statistics and the like. Also, it could be buggy.\n\n")


#The goal is to obtain P(T.C.|H). T.C.=total cases, H=hospitalized/serious-critical cases.
#We take a uniform prior, as we have no reason to think anything about P(T.C) a priori.
#P(H|T.C) is just a binomial distro, with T.C. trials and H "successes" (not the best word, probably).
#We take the p parameter for the binomial distro from somewhere (the rate of hospitalization of a country
#with many cases and reliable detection, for instance? By default I use the rate from S. Korea.

#We then calculate P(T.C|H) for a large array of T.C., and for a given H interval (H is assumed to be known).
#We finally sum over all the P(T.C.|H) for the T.C. interval for which we want the probability.

#we also plot P(T.C.|H) vs T.C. and fill only the part correpsonding to our interval of interest.

parser = argparse.ArgumentParser()
    

parser.add_argument("RL", type=int, default=0, help="Lower limit for real cases")
parser.add_argument("RU", type=int, default=9000, help="Upper limit for real cases")

parser.add_argument("-s", "--serious", type=int,default=0.01, help="True probability of serious cases. Default taken from s. Korea.")

parser.add_argument("-l", "--hlow", type=float,default=32.0, help="Lower limit for the hospitaized people in the country")

parser.add_argument("-u", "--hupp", type=float,default=34.0, help="Upper limit for the hospitaized people in the country")

args = parser.parse_args()


hrange=args.hupp-args.hlow

rc=[args.RL,args.RU] #limits for real cases we want to test




cdfbars=[]

for j in range(0,rc[1]+5000):
    cdf.append(scipy.stats.binom.cdf(args.hupp,n=j,p=args.serious)-scipy.stats.binom.cdf(args.hlow,n=j,p=args.serious))
    norma=norma+cdf[-1] #The normalization factor is P(H), the probability of being hospitalized regardless of the number of cases (it's actually the same as the success rate of the binomial distro we use, but I calculated explicitly anyway).
    cdfbars.append(0)
    if j>=rc[0] and j<=rc[1]:
        totalprob=totalprob+cdf[-1] ## P(H|T.C.)
        cdfbars[-1]=cdf[-1]
    realcases.append(j)


print("The estimated probability for there being between %d and %d real cases when between %d and %d hospitalized people (serious and critical cases) are observed, and with a %5.3f%% of true serious cases is of:\n%5.3f"%(args.RL,args.RU,args.hlow,args.hupp,args.serious*100,totalprob/norma))

#plt.hist(cdfbars,bins=realcases, facecolor='g', alpha=0.75)
#plt.plot(realcases,cdf)


fig, ax = plt.subplots()
ax.set_title("Prob. for total cases given %d-%d serious cases and a true %3.2f%% serious"%(args.hlow,args.hupp,args.serious*100))
ax.plot(realcases,cdf)


#ax.fill(realcases,cdf, 'b', realcases, cdfbars, 'r', alpha=0.3)
ax.fill_between(realcases,0,cdfbars,color="r", alpha=0.3)



plt.show()





