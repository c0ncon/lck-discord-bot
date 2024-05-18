FROM golang:latest as build
ENV CGO_ENABLED=0
WORKDIR /src
COPY . .
RUN go build -a .

FROM scratch
ENV TZ=Asia/Seoul
WORKDIR /app

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /src/config ./config
COPY --from=build /src/schedule.json .
COPY --from=build /src/lck-discord-bot .

ENTRYPOINT ["./lck-discord-bot"]