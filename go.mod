module github.com/mudler/eirini-ingress

go 1.14

require (
	github.com/SUSE/eirinix v0.2.1-0.20200430122945-e30cc67ba0be
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/spf13/cobra v0.0.7
	github.com/spf13/viper v1.7.0
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/sys v0.0.0-20200501145240-bc7a7d42d5c3 // indirect
	golang.org/x/tools v0.0.0-20200504193531-9bfbc385433f // indirect
	k8s.io/api v0.0.0-20200404061942-2a93acf49b83
	k8s.io/apimachinery v0.0.0-20200410010401-7378bafd8ae2
	k8s.io/client-go v0.0.0-20200330143601-07e69aceacd6
)

replace code.cloudfoundry.org/cf-operator => code.cloudfoundry.org/quarks-operator v1.0.1-0.20200413083459-fb39a29ad746
