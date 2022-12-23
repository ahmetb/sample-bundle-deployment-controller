# Copyright 2022 Ahmet Alp Balkan

FROM golang:1.19-alpine
WORKDIR /go/src/bundle-deployment-controller
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/controller ./controller
FROM golang:1.19-alpine
RUN apk add --no-cache git
WORKDIR /bin
COPY --from=0 /bin/controller /bin/controller
ADD "https://dl.k8s.io/release/v1.26.0/bin/linux/amd64/kubectl" /bin/kubectl
RUN chmod +x ./kubectl
ADD "https://github.com/GoogleContainerTools/kpt/releases/download/v1.0.0-beta.24/kpt_linux_amd64" /bin/kpt
RUN chmod +x ./kpt
ENTRYPOINT ["/bin/controller"]
