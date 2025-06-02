# syntax=docker/dockerfile:1.4

# obraz Golang w wersji Alpine dla mniejszego rozmiaru
FROM golang:1.21-alpine3.18 AS builder

# katalog rboczy
WORKDIR /build

# kopiowanie tylko potrzebnych plików dla optymalizacji cache
COPY go.mod ./

# pobranie zależności (osobna warstwa dla lepszego cachowania)
RUN go mod download

# kopiowanie kodu źródłowego
COPY main.go ./

# kompilacja z maksymalnymi flagami optymalizacji do binarki statycznej
# -ldflags="-s -w": usunięcie symboli debugowania 
# -trimpath: usunięcie ścieżek absolutnych dla bezpieczeństwa i mniejszego rozmiaru
# CGO_ENABLED=0: kompilacja statyczna bez zależności C
RUN CGO_ENABLED=0 go build \
    -ldflags="-s -w -extldflags '-static'" \
    -trimpath \
    -o app-small main.go

# Alpine do kompresji UPX, co dodatkowo zmniejsza rozmiar pliku binarnego
FROM alpine:3.18 AS compressor
RUN apk add --no-cache upx
COPY --from=builder /build/app-small /app-small

# kompresja z maksymalnymi opcjami
# --best: najlepszy poziom kompresji
# --ultra-brute: najbardziej agresywna kompresja
# --lzma: algorytm LZMA oferujący lepszą kompresję
# --overlay=strip: usuniecie sekcji overlay dla mniejszego rozmiaru
RUN upx --best --ultra-brute --lzma --overlay=strip --no-backup /app-small -o /app-compressed

# warstwa scratch dla absolutnie minimalnego rozmiaru obrazu
# scratch nie zawiera żadnych narzędzi, tylko gołą binarkę
FROM scratch

# metadane obrazu, które są zgodne ze standardem OCI (dodalam created i version)
LABEL org.opencontainers.image.authors="Ewelina Musińska"
LABEL org.opencontainers.image.title="Aplikacja pogodowa"
LABEL org.opencontainers.image.description="Aplikacja pogodowa napisana w Go, która korzysta z API OpenWeatherMap i wyświetla prognozę pogody."
LABEL org.opencontainers.image.created="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
LABEL org.opencontainers.image.version="1.0.0"

# kopiowanie certyfikatów CA dla obsługi HTTPS
# jest to konieczne, ponieważ warstwa scratch nie zawiera żadnych certyfikatów
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# kopiowanie skompresowanej binarki
COPY --from=compressor /app-compressed /app

# informacja o porcie nasłuchiwania
EXPOSE 8080

# health-check sprawdzająacy, czy proces nadal działa
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app", "health"] || exit 1

# uzytkownik 65534 (nobody) jest często używany jako użytkownik nieprivilegowany
# aplikacje uruchamiane jako root mogą być podatne na ataki
# uzycie UID 65534 (nobody) dla bezpieczeństwa
USER 65534:65534

# uruchomienie aplikacji
ENTRYPOINT ["/app"]