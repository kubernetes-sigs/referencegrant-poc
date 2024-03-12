FROM golang:latest as builder

WORKDIR /go/src/github.com/kubernetes-sigs/referencegrant-poc

COPY . /go/src/github.com/kubernetes-sigs/referencegrant-poc

RUN CGO_ENABLED=0 go build -o refauthz

# Create New Base Image.
FROM alpine:latest

WORKDIR /go/bin/
# RUN mkdir config
RUN mkdir -p /usr/.kube/
ENV KUBECONFIG /usr/.kube/config

COPY --from=builder /go/src/github.com/kubernetes-sigs/referencegrant-poc/refauthz .

EXPOSE 8081
CMD ["./refauthz"]
