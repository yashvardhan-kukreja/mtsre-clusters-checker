# vi:set ft=dockerfile:
# --- stage 1 ---
FROM registry.redhat.io/rhel8/go-toolset:1.17 as builder

USER root

WORKDIR /build

RUN echo "noroot:x:10001:10001:noroot:/:/sbin/nologin" | tee -a /etc/passwd

# Cache optimization
COPY go.mod go.sum Makefile ./
RUN go mod download

COPY . ./
RUN make build

# --- stage 2 ---
FROM scratch

WORKDIR /
COPY --from=builder /build/bin/clusters-checker /
COPY --from=builder /etc/passwd /etc/passwd

USER "noroot"

ENTRYPOINT ["/clusters-checker"]