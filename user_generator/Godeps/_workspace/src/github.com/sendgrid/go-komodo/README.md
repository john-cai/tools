#Komodo 
###Monitor Service

![the monitor lizard service](http://i.imgur.com/NcGVoxr.jpg)

Basic Usage
---
Embed the package you wish to monitor:

Steps
---

1. Implement Adminable interface and pass it to komodo.NewServer()

2. Create a listener and pass it to komodo.Serve() (or use komodo.ListenAndServe())

See [example file](https://github.com/sendgrid/go-komodo/blob/master/example_test.go)

Run Tests
---
```bash
go test
```
