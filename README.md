# supervisor
Supervisor trees for Go

Use it as in `go Supervise(ctx, c1, c2)` and child goroutines/workers/agents should have the signature `func(ctx context.Context, _ ...Tree)`. Children have to respect the `ctx context.Context` and check `ctx.Done()` in each iteration, a proper child implementation would be (just a guideline):

```go
func childGoroutine(ctx context.Context, _ ...Tree) {
	// ... initial preps

	// main loop
	for {
		// single check
		select {
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-ctx.Done():
			return
			// ... check ctx alongside other cases other cases
		}

		// and if there is another inner loop
		for i := 0; i < N; i++ {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// ... loop body
		}
	}
}
```

For using a supervisor with specific settings, a context (`context.Context`) should be created using `WithOptions`:

```go
ctx, _ := WithOptions(
	rootCtx,
	OneForOne,
	2,
	time.Millisecond*300)
```

In which we set the parent context, strategy for restarting the goroutine, intensity for restarting the goroutine (max number of restarts) and the period of restarts. If number of restarts of a goroutine reaches it's intensity, it will no longer gets restarted. When all children of a supervisor are done, the supervisor will stop.

In case of `SimpleOneForOne` strategy, supervisor won't stop even if it has no children, unless either the parent context gets cancels or the channel returned from `WithChannel` gets closed.