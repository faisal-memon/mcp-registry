FROM golang:1.26.6-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/mcp-registry ./cmd/mcp-registry

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/mcp-registry /mcp-registry
EXPOSE 8080
ENTRYPOINT ["/mcp-registry"]
