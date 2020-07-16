# EiriniX Ingress extension

This is a simple PoC for Eirini ingress extension.

![Eirini-ingress](https://user-images.githubusercontent.com/2420543/87640475-10ff6580-c747-11ea-8937-25df4b6a42ca.png)

The extension has a simple duty: create the appropriate Kubernetes Services and Ingress endpoints for application pushed with CloudFoundry on K8s.

The extension can work in HA mode.

- Simple - simple to hack and understand
- Fault tolerant - if the component goes down, the apps are still served
- Stateful - no need to update and re-register routes
- High integration - pick up the ingress-controller you like to handle workload to
- Customizable - Easy to tweak and change the generated resources

## How to use

TODO