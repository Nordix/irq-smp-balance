# irq-smp-balance

This provides ability to isolate the Guaranteed QoS class pod cores from handling interrupts.

## Quickstart Installation

This is a daemonset running in all the worker nodes watching for Guaranteed QoS class pods
having labels with `irq-load-balancing.docker.io=true` and exclude these pod CPUs from IRQ
balancing.

Clone this GitHub repository, we'll apply a daemonset which installs irq-smp-balance using
`kubectl` from this repo. From the root directory of the clone, apply the daemonset YAML file:

```
$ cat ./deployments/auth.yaml | kubectl apply -f -
$ cat ./deployments/irqsmpbalance-daemonset.yaml | kubectl apply -f -
```

### Note

The irqsmpbalance-daemonset.yml considers `irqbalance` config file present at `/etc/sysconfig/`
directory.

This may not be a case with Ubuntu as this file would present at `/etc/default/` directory.

Hence `hostPath` in `irqbalanceconf` have to be updated accoringly before the deployment.
