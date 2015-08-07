go-apid
=======

go-apid: Go Apid Client

Usage
-------
```go
apidUrl := "http://stapid-010.sjc1.sendgrid.net:8082"
apidClient := apid.NewHTTPClient(apidUrl)
apidClient.AddFunction(APIdFunction("getCampaign"), apid.FunctionInfo{
	Return: "result",
	Path: "/api/apps/events/get.json",
})

var result campaign
err := apidClient.DoFunction(APIdFunction("getCampaign"),
	url.Values{"id": {strconv.Itoa(campaignId)}},
	&result,

)
if err != nil {
	Logger.Err("Error executing apid do function", nil)
}

```
HealthcheckableClient is provided for easier integration with [komodo.Healthcheck](https://github.com/sendgrid/go-komodo/blob/master/komodo.go#L369)
interface. HTTPClient implements this as well.

Using Healthcheck in your app:
--------
Assuming your app implements [komodo.Adminable](https://github.com/sendgrid/go-komodo/blob/master/komodo.go#L389)
interface, in your Healtchecks you can just add your apidClient.

Actual healtchcheck is done by making sure /api/functions.json is reachable.

```go
type myApp struct {
	...
	myApidClient apid.HealthcheckableClient
}
...

func (m *myApp) Healthchecks() []komodo.Healthcheck {
	h := make([]komodo.Healthcheck, 0)
	h = append(h, m.myApidClient)
	return h
}
```
