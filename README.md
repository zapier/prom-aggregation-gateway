# Prom Aggregation Gateway Helm Chart

First, you need to get the repo added to helm!

```sh
helm repo add pag https://zapier.github.io/prom-aggregation-gateway/
helm repo update
helm search repo pag -l
```

Then you can install it:

```sh
helm install pag pag/prom-aggregation-gateway 0.5.1
```