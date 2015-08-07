Change log
==========
## 0.3.0
* Create HealthcheckableClient, and restore Client interface to v.0.1.0 state.

## 0.2.2
* Make fake client safe to run concurrently

## 0.2.1
* Remove go-komodo import statement so that go-apid will compile without it.

## 0.2.0
* Make Client implement komodo.Healthcheck interface and provide an implementation for HTTPClient.

## 0.1.0
* initial