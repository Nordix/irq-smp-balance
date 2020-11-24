FROM golang:1.13-alpine as builder
ADD . /usr/src/irq-smp-balance
RUN apk add --update --virtual build-dependencies build-base linux-headers && \
    cd /usr/src/irq-smp-balance && \
    make

FROM golang:1.13-alpine
RUN apk add --update irqbalance
COPY --from=builder /usr/src/irq-smp-balance/bin/smpaffinity /usr/bin/

CMD ["smpaffinity"]
