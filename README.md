# Predicting stock prices using Monte Carlo Simulations in Go
See the full blog post [here](https://medium.com/@matt.a.french/predicting-stock-prices-using-monte-carlo-simulations-in-go-26060ab2836)

## What is this script?
This script uses historical data for a particular stock, and runs a Monte Carlo simulation to predict possible future returns.

The data is exported as a CSV, allowing you to use Excel (or another tool) to make snazzy diagrams like this one:

![Graph showing Monte Carlo predictions](https://cdn-images-1.medium.com/v2/resize:fit:1600/1*_0Z6h1XcCEqOgXYL9QUC6A.png)

## What is a Monte Carlo Simulation?
> Monte Carlo experiments, are a broad class of computational algorithms that rely on repeated random sampling to obtain numerical results. The underlying concept is to use randomness to solve problems that might be deterministic in principle.

The principle of a Monte Carlo simualtion is fairly simple, generate a whole bunch of random numbers, test them, and analyse the results. 

It is excellent for problems where you may not be able to determine the underlying behaviour of system, but you know that it behaves in a determinitisc way.

Simulations that follow the Monte Carlo method can be used on a whole range of problems, from calculating Pi to particle dynamics.

In our case, we're going to be using the Monte Carlo method by generating random numbers to simulate the volatility of a stock.

## What is Brownian Motion?
In physics, Brownian Motion describes the seemingly random behaviour of particles in a gas or liquid. However, what we're interesting in the mathematical definition, also know as the Weiner Process.

The fundamental idea is that a system (our stock) has standard drift, and a random shock or volatility. So on average, it will trend in the direction of the drift, but it will also experience smaller, random, direction changes. 

This applies very well to stock prices, where a company's value tends to increase or decrease by a standard factor, and also experiences random jumps and drops in value based on market factors.

In our simulation, we'll be using Geometric Brownian Motion (GBM), which is defined as:

![Geometric Brownian Motion Equation](https://cdn-images-1.medium.com/v2/resize:fit:1600/1*VU3iODREo4GWqhjOHZwSRQ.png)

Where µ is the expected daily logarithmic percentage return, σ is the standard deviation of daily returns, and ϵ is the normally-distributed random variable.

For each day in our simulation, we'll calculate the GBM based on the randomly generate ϵ, and multiply yesterdays price by it.

## Why Go (Golang)?
Predominately, performace. Go is extremely fast an memory efficent, while being relatively easy to work with.
