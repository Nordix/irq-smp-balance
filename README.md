# irq-smp-balance

This provides ability to isolate the Guaranteed QoS class pod cores from handling interrupts in a K8s cluster.

## Quickstart Installation

### Build

Here is the procedure to compile and build irq-smp-balance docker image.

```
$ git clone https://github.com/pperiyasamy/irq-smp-balance.git
$ cd irq-smp-balance
$ make
$ make image
```

### Deployment

This is a daemonset running in all the worker nodes watching for Guaranteed QoS class pods
having labels with `irq-load-balancing.docker.io=true` and exclude assigned pod CPUs from IRQ
balancing.

We'll apply a daemonset which installs irq-smp-balance using `kubectl` from this repo.
From the root directory of the clone, apply the daemonset YAML file:

```
$ cat ./deployments/auth.yaml | kubectl apply -f -
$ cat ./deployments/irqsmpbalance-daemonset.yaml | kubectl apply -f -
```

### irqsmpdaemon

TODO
