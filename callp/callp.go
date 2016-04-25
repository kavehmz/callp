/*
Package callp is a pricing streamer. It is design for an old company which entered concurrency era while refacotring the main code is not an option.

There are several tens of websocket servers which need stream of prices for different markets.

Solution is simple.

Websocket servers will race to announce a pricing detail in a cluster of redis server. First one will win.
callp damons will race to pick the announced pricing jobs. Each one will pick one.
callp will start call a Perl script and ask it to produce price everytime there is a new market signal.

Here, perl script are responsible for pricing that is not rewirteable in Go at the moment.
callp will handle the concurrent contorl of perl processes, communicating with redis and waiting for market signals and after geting price for each signal callp will pushlish it in redis to be consumed by perl websocket processes.

Perl websocket processor will add jobs using a redlock algorithm [1] for handle the race. There is a race as different processes might need the same stream of prices and we want to only generate them once.

[1] http://redis.io/topics/distlock
*/
package callp

// PricerInactivityTimeout sets how long each pricer can be idle before it quits (read or write). This value in millisecond.
var PricerInactivityTimeout = 60000

// PricerReadTimeout sets how long each pricer will wait for a read to come after each write. This value is in millisecond.
var PricerReadTimeout = 1000

// TimeoutMultiplier can change scale of PricerInactivityTimeout and PricerReadTimeout. Timer functions accept nanosecond unit.
// TimeoutMultiplier is used to multiply custom timeouts. if TUnit is 1,000,000,  PricerInactivityTimeout and PricerReadTimeout will be in millisecond.
var TimeoutMultiplier = 1000000

// WaitIfNoJob set the sleep in the loop which waits for next job to be ready.
var WaitIfNoJob = 100

// PricingScript sets path to pricing script.
// This script should communicate by STDIN and STDOUT in an infinite loop until STDIN gets closed from outside.
// First input will be language.
// Second input is pricing parameters
// Every other STDIN input is a signal to generate the next price.
var PricingScript = "./pricer.pl"
