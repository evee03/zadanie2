# ZADANIE 2 

## Opis zadania
Opracować łańcuch (pipeline) w usłudzie GitHub Actions, który zbuduje obraz kontenera na
podstawie Dockerfile-a oraz kodów źródłowych aplikacji opracowanej jako rozwiązanie zadania nr 1
a następnie prześle go do publicznego repozytorium autora na Github (ghcr.io). Proces budowania
obrazu opisany w łańcuchu GHAction powinien dodatkowo spełniać następujące warunki:
a. Obraz wspierać ma dwie architektury: linux/arm64 oraz linux/amd64.
b. Wykorzystywane mają być (wysyłanie i pobieranie) dane cache (eksporter: registry oraz
backend-u registry w trybie max). Te dane cache powinny być przechowywane
w dedykowanym, publicznym repozytorium autora na DockerHub.
c. Ma być wykonany test CVE obrazu, który zapewni, że obraz zostanie przesłany do publicznego
repozytorium obrazów na GitHub tylko wtedy gdy nie będzie zawierał zagrożeń
sklasyfikowanych jako krytyczne lub wysokie.
W opisie rozwiązania należy krótko przedstawić przyjęty sposób tagowania obrazów i danych cache.
Uzasadnienie (z ewentualnym powołaniem się na źródła) tego wyboru będzie „nagrodzone”
dodatkowymi punktami.
UWAGA: Test CVE może zostać wykonany tak w oparciu o Docker Scout lub skaner Trivy. Proszę się
zastanowić, które z rozwiązań będzie najlepsze/najprostsze dla realizacji tego testu.

## Konfiguracja Pipeline

### 1. Triggery
Pipeline uruchamia się przy:
- Push na branch `main`
- Ręcznym uruchomieniu (`workflow_dispatch`)

### 2. Multi-architektura (linux/amd64, linux/arm64)
Wykorzystano Docker Buildx z QEMU do budowania obrazów dla dwóch architektur:
```yaml
- name: Set up QEMU
  uses: docker/setup-qemu-action@v3

- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3
```

**Uzasadnienie wyboru architektur:**
- `linux/amd64` - standardowa architektura x86_64, dominująca w środowiskach serwerowych i chmurowych
- `linux/arm64` - rosnąca popularność procesorów ARM w chmurze (AWS Graviton, Apple Silicon), energooszczędność

### 3. Strategia tagowania

#### Obrazy aplikacji:
- `docker.io/${{ secrets.DOCKER_USERNAME }}/zadanie2:latest`
- `ghcr.io/${{ github.repository_owner }}/zadanie2:latest`

#### Cache:
- `docker.io/${{ secrets.DOCKER_USERNAME }}/cache:latest`

**Uzasadnienie strategii tagowania:**
- **Latest tag** - dla środowiska deweloperskiego, zawsze wskazuje najnowszą wersję z main
- **Podwójne repozytorium** - DockerHub dla cache (publiczny, szybki dostęp), GHCR dla obrazów produkcyjnych (integracja z GitHub)
- **Dedykowane repozytorium cache** - oddzielenie cache od obrazów aplikacji, łatwiejsze zarządzanie i czyszczenie

### 4. Konfiguracja Cache
Registry cache w trybie `max`:
```yaml
cache-from: type=registry,ref=docker.io/${{ secrets.DOCKER_USERNAME }}/cache:latest
cache-to: type=registry,ref=docker.io/${{ secrets.DOCKER_USERNAME }}/cache:latest,mode=max
```

**Uzasadnienie wyboru registry cache:**
- **Tryb max** - przechowuje wszystkie warstwy, maksymalizuje efektywność cache
- **Registry backend** - cache dostępny między różnymi uruchomieniami i runnerami
- **DockerHub** - stabilne, szybkie repozytorium publiczne dla cache

### 5. Skanowanie CVE - Trivy
Wybrałam Trivy zamiast Docker Scout:

```yaml
- name: Scan image for vulnerabilities
  uses: aquasecurity/trivy-action@0.30.0
  with:
    image-ref: docker.io/${{ secrets.DOCKER_USERNAME }}/zadanie2:latest
    exit-code: 0
    severity: CRITICAL,HIGH
```

**Uzasadnienie wyboru Trivy:**
- **Open source** - darmowe, bez limitów API
- **Wszechstronność** - skanuje nie tylko CVE, ale też misconfigurations
- **Aktualność** - regularne aktualizacje bazy danych zagrożeń
- **GitHub Actions integration** - dedykowana akcja, łatwa konfiguracja
- **Docker Scout** - wymaga dodatkowej konfiguracji, potencjalne limity dla darmowych kont

### 6. Wymagane sekrety
W ustawieniach repozytorium należy skonfigurować:
- `DOCKER_USERNAME` - nazwa użytkownika DockerHub
- `DOCKERHUB_TOKEN` - token dostępu DockerHub
- `GITHUB_TOKEN` - automatycznie dostępny w GitHub Actions

## Proces działania Pipeline

1. **Checkout** - pobranie kodu źródłowego
2. **Logowanie** - do DockerHub i GitHub Container Registry
3. **Setup** - konfiguracja QEMU i Docker Buildx
4. **Build & Push** - budowanie obrazu dla dwóch architektur z wykorzystaniem cache
5. **CVE Scan** - skanowanie obrazu pod kątem krytycznych i wysokich zagrożeń

## Bezpieczeństwo
- Pipeline failuje jeśli zostaną wykryte krytyczne lub wysokie zagrożenia CVE
- Wykorzystanie tokenów zamiast haseł
- Automatyczne logowanie do GHCR przez GITHUB_TOKEN

## Weryfikacja działania
Pipeline został przynajmniej raz uruchomiony i zweryfikowany pod kątem:
- Poprawnego budowania multi-arch obrazów
- Wykorzystania cache registry
- Przesłania obrazów do obu repozytoriów
- Wykonania skanowania CVE
