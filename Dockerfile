FROM golang:alpine AS builder

WORKDIR /workspace

COPY . .

RUN go build -o tapi .

FROM alpine

COPY --from=builder /workspace/tapi /bin/tapi

EXPOSE 8080

ENTRYPOINT [ "/bin/tapi" ]
