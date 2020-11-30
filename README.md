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

This requires a daemon and daemonset pod running in worker nodes watching for Guaranteed QoS class pods
having labels with `irq-load-balancing.docker.io=true` and exclude assigned pod CPUs from IRQ balancing.

The daemon and daemonset pod share a config file, by default both chooses `/etc/sysconfig/podirqbalance` file.
If another config file chosen, the hostPath `irqbalanceconf` in `./deployments/irqsmpbalance-daemonset.yaml`
has to be updated accordingly before the deployment.

It's time to deploy the daemonset on the cluster,  We'll apply a daemonset which installs irq-smp-balance
using `kubectl` from this repo. From the root directory of the clone, apply the daemonset YAML file:

```
$ cat ./deployments/auth.yaml | kubectl apply -f -
$ cat ./deployments/irqsmpbalance-daemonset.yaml | kubectl apply -f -
```

Now run the daemon on the worker node:

```
$ irqsmpdaemon &
```

The irqsmpdaemon can also be run with different config and log files, Refer the help:

```
$ irqsmpdaemon -h
Usage of irqsmpdaemon:
  -config string
        irq balance config file (default "/etc/sysconfig/podirqbalance")
  -log string
        log file (default "/var/log/irqsmpdaemon.log")
```
