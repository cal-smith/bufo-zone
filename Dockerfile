FROM golang:1.21-bullseye as builder
WORKDIR /gocode
COPY gobufo/. /gocode/
RUN go build -o bufo_zone . 

FROM python:3.11-slim-bullseye
ENV PYTHONUNBUFFERED=1
ENV PROD=true
# Latest releases available at https://github.com/aptible/supercronic/releases
ENV SUPERCRONIC_URL=https://github.com/aptible/supercronic/releases/download/v0.2.26/supercronic-linux-amd64 \
    SUPERCRONIC=supercronic-linux-amd64 \
    SUPERCRONIC_SHA1SUM=7a79496cf8ad899b99a719355d4db27422396735

RUN mkdir -p /code
WORKDIR /code

RUN apt-get update && apt-get install -y supervisor curl

RUN curl -fsSLO "$SUPERCRONIC_URL" \
 && echo "${SUPERCRONIC_SHA1SUM}  ${SUPERCRONIC}" | sha1sum -c - \
 && chmod +x "$SUPERCRONIC" \
 && mv "$SUPERCRONIC" "/usr/local/bin/${SUPERCRONIC}" \
 && ln -s "/usr/local/bin/${SUPERCRONIC}" /usr/local/bin/supercronic

RUN pip install poetry
COPY pyproject.toml poetry.lock /code/
RUN poetry config virtualenvs.create false
RUN poetry install --only main --no-root --no-interaction
COPY . /code
COPY --from=builder /gocode/bufo_zone /code/
COPY --from=builder /gocode/static /code/static

# RUN python manage.py collectstatic --noinput

EXPOSE 8000

CMD ["supervisord", "-c", "supervisor.conf"]
