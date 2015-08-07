envy
====
Go configuration using environment variables.

You can set environment variables in several ways:

1. During command execution:

        APP_PORT=80 APP_HOST=localhost go run main.go

2. Exporting into your shell:

        export APP_PORT=80
        export APP_HOST=localhost
        go run main.go

3. Sourcing a script containing your environment:

    **`config.sh`** contains:

		export APP_PORT=80
		export APP_HOST=localhost

    Source it and run your app:

        source config.sh
        go run main.go

### Testing
To run tests, simply run:
		
	go test

### How to use `envy`

[Example code here](https://github.com/sendgrid/go-envy/blob/master/example_test.go)

To use **envy**, define a struct representing your configuration. (The field
types can be `int`, `string`, `bool`, `[]string`, or `[]int`.) Then call `envy.LoadWithPrefix` with
your config struct to populate it.

Envy will unmarshal the environment variables of the form of PREFIX_FIELD
(all-caps) into your struct.