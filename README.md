# [![Docker Repository on Quay](https://quay.io/repository/mudler/eirinix-ingress/status "Docker Repository on Quay")](https://quay.io/repository/mudler/eirinix-ingress) ![Build and Test](https://github.com/mudler/eirini-ingress/workflows/Build%20and%20Test/badge.svg) EiriniX Ingress extension

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

```bash
$> kubectl apply -f https://raw.githubusercontent.com/mudler/eirini-ingress/master/contrib/kube.yaml
```

*Note: In case Eirini is not deploying the workload in the namespace `eirini`, you might need to tweak the role binding manually.*

### Uninstall

```bash
$> kubectl delete -f https://raw.githubusercontent.com/mudler/eirini-ingress/master/contrib/kube.yaml
```