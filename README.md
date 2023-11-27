# Healthcheck and Watchdog service

<div style="text-align: left;">
    <img width="100%" src="assets/logo.png" width="48">
</div>

Surveillance and helpchecks service allows you to connect via http/https and ws/wss
points and upload information about the answer to prometheus.

Healthcheck:

- HTTP requests;
- HTTP/HTTPS requests with OAuth-authentication;
- Monitoring websocket connections;
- Connection dependencies. Choose which task should be success
  Then start another task;
- Control of going out of memory limits;

Watchdog:

- Kubernetes/Openshift:
  - Delete pod;
  - Scale deployment/statefulset;
  - Custom scenarios;
- Redis;
  - Execute command FLUSHALL;
