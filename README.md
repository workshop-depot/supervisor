# supervisor
Supervisor trees for Go

```
This package was a study, not much usage in it's current form. Full blown
solutions (actor models) such as protoactor-go serve difference scenarious 
far better, and using an interface for actors (calling methods on each message) 
is more logical, because we can not force the user to respect <-ctx.Done() and 
that will make a mess of things. 

Previous implementation is available at v0 branch.
```