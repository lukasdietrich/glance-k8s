from docker.io/library/golang:alpine as builder

	workdir /github.com/lukasdietrich/glance-k8s

	copy internal ./internal
	copy cmd ./cmd
	copy go* .

	run go build -v ./cmd/glance-k8s

from docker.io/library/alpine

	workdir /app

	copy --from=builder /github.com/lukasdietrich/glance-k8s/glance-k8s .

	label org.opencontainers.image.authors="Lukas Dietrich <lukas@lukasdietrich.com>"
	label org.opencontainers.image.source="https://github.com/lukasdietrich/glance-k8s"

	cmd [ "/app/glance-k8s" ]
