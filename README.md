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