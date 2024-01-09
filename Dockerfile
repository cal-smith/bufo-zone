FROM golang:1.21-bullseye as builder
WORKDIR /gocode
COPY . /gocode
RUN go build -o bufo_zone ./gobufo
RUN go build -o bufo_sync ./syncbufo

FROM debian:bullseye-slim
ENV SUPERCRONIC_URL=https://github.com/aptible/supercronic/releases/download/v0.2.26/supercronic-linux-amd64 \
    SUPERCRONIC=supercronic-linux-amd64 \
    SUPERCRONIC_SHA1SUM=7a79496cf8ad899b99a719355d4db27422396735 \
    OVERMIND_URL=https://github.com/DarthSim/overmind/releases/download/v2.4.0/overmind-v2.4.0-linux-amd64.gz \
    OVERMIND_SHA256SUM=1f7cac289b550a71bebf4a29139e58831b39003d9831be59eed3e39a9097311c \
    OVERMIND=overmind-v2.4.0-linux-amd64

RUN mkdir -p /code
WORKDIR /code

RUN apt-get update && apt-get install -y curl tmux

RUN curl -fsSLO "$SUPERCRONIC_URL" \
    && echo "${SUPERCRONIC_SHA1SUM}  ${SUPERCRONIC}" | sha1sum -c - \
    && chmod +x "$SUPERCRONIC" \
    && mv "$SUPERCRONIC" "/usr/local/bin/${SUPERCRONIC}" \
    && ln -s "/usr/local/bin/${SUPERCRONIC}" /usr/local/bin/supercronic
# RUN go install github.com/DarthSim/overmind/v2@latest
RUN curl -fsSLO "$OVERMIND_URL" \
    && echo "${OVERMIND_SHA256SUM}  ${OVERMIND}.gz" | sha256sum -c - \
    && gunzip "${OVERMIND}.gz" \
    && chmod +x "${OVERMIND}" \
    && mv "${OVERMIND}" "/usr/local/bin/${OVERMIND}" \
    && ln -s "/usr/local/bin/${OVERMIND}" /usr/local/bin/overmind

COPY . /code
COPY --from=builder /gocode/bufo_zone /code/
COPY --from=builder /gocode/bufo_sync /code/
COPY --from=builder /gocode/gobufo/static /code/static

EXPOSE 8000

CMD ["overmind", "s"]
