# Skattification of the Naiserator

WIP

## Repos

https://github.com/skatteetaten-trial/liberator
https://github.com/skatteetaten-trial/naiserator

## Getting started

```console
git clone git@github.com:skatteetaten-trial/naiserator.git
```

```console
git clone git@github.com:skatteetaten-trial/liberator.git
```

```console
cd liberator
```

Edit Makefile and set CONTROLLER_GEN_VERSION to "v0.6.2".
Make sure you have `controller-gen` installed and run
```console
make generate
```

Apply CRD´s:
```console
for i in config/crd/bases/nais.io_*;do kubectl apply -f $i;done
```

Go to naiserator repo:
```console
cd ../naiserator
```

Modify to use a locally checked out copy of Liberator
```console
go mod edit -replace github.com/nais/liberator=../liberator
make build #?
```

Start naiserator
```console
make local
```
