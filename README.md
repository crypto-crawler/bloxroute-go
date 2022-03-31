# bloxroute-go

A bloXroute websocket client written in Go.

## Typesafe

This library provides typesafe APIs, not just a raw `chan string` channel. For example, `SubscribeNewTxs()` outputs messages to a `chan *types.Transaction` channel.

## Flexible

Each bloXroute `subscribe` returns a unique subscription ID, see [Creating a Subscription
](https://docs.bloxroute.com/streams/working-with-streams/creating-a-subscription). This library takes advantage of this feature to provide a flexible output mechanism: each `SubscribeXXX()` has a `outpuCh` channel, multiple `SubscribeXXX()` calls can output to the same channel, or to a different channel.
