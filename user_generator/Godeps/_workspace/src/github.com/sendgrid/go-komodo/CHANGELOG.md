Change log
==========
## 2.0.0
- Updating healthcheck endpoint based on json format to conform to https://wiki.sendgrid.net/pages/viewpage.action?pageId=8295814#MonitoringDesign-Healthcheck
- Adding Name() method to Adminable interface

## 1.0.2
- Changed Err in HealthCheckresult to string so it can be unmarshalled
- Fixed closure bug that caused healthchecks call to return the last healthcheck added
