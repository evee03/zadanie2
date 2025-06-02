package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

)

// Minimalny zestaw danych pogodowych
type WD struct {
	M struct {
		T  float64 `json:"temp"`
		F  float64 `json:"feels_like"`
		H  int     `json:"humidity"`
		P  int     `json:"pressure"`
	} `json:"main"`
	W []struct {
		D string `json:"description"`
	} `json:"weather"`
	Wi struct {
		S float64 `json:"speed"`
	} `json:"wind"`
	N string `json:"name"`
}

// Krótkie nazwy kluczy
var cs = map[string][]string{
	"Polska":     {"Warszawa", "Kraków", "Gdańsk", "Poznań", "Wrocław"},
	"Niemcy":     {"Berlin", "Monachium", "Hamburg", "Kolonia"},
	"Francja":    {"Paryż", "Marsylia", "Lyon", "Nicea"},
	"W.Brytania": {"Londyn", "Manchester", "Liverpool", "Glasgow"},
	"Włochy":     {"Rzym", "Mediolan", "Wenecja", "Florencja"},
}

// Tłumaczenia opisów pogody
var weatherTranslations = map[string]string{
	"clear sky":           "Bezchmurne niebo",
	"few clouds":          "Kilka chmur",
	"scattered clouds":    "Rozproszone chmury",
	"broken clouds":       "Zachmurzenie",
	"shower rain":         "Przelotne opady",
	"rain":                "Deszcz",
	"thunderstorm":        "Burza",
	"snow":                "Śnieg",
	"mist":                "Mgła",
	"light rain":          "Lekki deszcz",
	"moderate rain":       "Umiarkowany deszcz",
	"heavy intensity rain": "Intensywny deszcz",
	"overcast clouds":     "Całkowite zachmurzenie",
	"fog":                 "Mgła",
	"light snow":          "Lekki śnieg",
	"moderate snow":       "Umiarkowany śnieg",
	"heavy snow":          "Intensywne opady śniegu",
	"drizzle":             "Mżawka",
	"light intensity drizzle": "Lekka mżawka",
	"heavy intensity drizzle": "Intensywna mżawka",
	"thunderstorm with light rain": "Burza z lekkim deszczem",
	"thunderstorm with rain": "Burza z deszczem",
	"thunderstorm with heavy rain": "Burza z intensywnym deszczem",
}

const k = "e580e81d235c7266923e136496f1b655"
const p = "8080"

// HTML template z polskimi nazwami
const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>Pogoda</title>
    <meta charset="UTF-8">
    <style>
        body { font: 14px Arial; max-width: 800px; margin: 0 auto; padding: 20px; }
        select, button { padding: 8px; margin: 10px 0; }
        #w { margin-top: 20px; padding: 15px; border: 1px solid #ddd; border-radius: 5px; display: none; }
    </style>
</head>
<body>
    <h1>Pogoda</h1>
    <div>
        <label for="c">Kraj:</label>
        <select id="c" onchange="updateCities()">
            <option value="">--</option>
            %s
        </select>
    </div>
    <div>
        <label for="y">Miasto:</label>
        <select id="y">
            <option value="">--</option>
        </select>
    </div>
    <button onclick="getWeather()">Sprawdź</button>
    <div id="w">
        <h2><span id="l"></span></h2>
        <div id="d"></div>
        <div>Temperatura: <span id="t"></span>°C</div>
        <div>Odczuwalna temperatura: <span id="f"></span>°C</div>
        <div>Wilgotność powietrza: <span id="h"></span>%%</div>
        <div>Ciśnienie atmosferyczne: <span id="p"></span> hPa</div>
        <div>Prędkość wiatru: <span id="s"></span> m/s</div>
    </div>
    
    <script>
        var d = document;
        
        // Tłumaczenia opisów pogody
        const weatherTranslations = {
            "clear sky": "Bezchmurne niebo",
            "few clouds": "Kilka chmur",
            "scattered clouds": "Rozproszone chmury",
            "broken clouds": "Zachmurzenie",
            "shower rain": "Przelotne opady",
            "rain": "Deszcz",
            "thunderstorm": "Burza",
            "snow": "Śnieg",
            "mist": "Mgła",
            "light rain": "Lekki deszcz",
            "moderate rain": "Umiarkowany deszcz",
            "heavy intensity rain": "Intensywny deszcz",
            "overcast clouds": "Całkowite zachmurzenie",
            "fog": "Mgła",
            "light snow": "Lekki śnieg",
            "moderate snow": "Umiarkowany śnieg",
            "heavy snow": "Intensywne opady śniegu",
            "drizzle": "Mżawka",
            "light intensity drizzle": "Lekka mżawka",
            "heavy intensity drizzle": "Intensywna mżawka",
            "thunderstorm with light rain": "Burza z lekkim deszczem",
            "thunderstorm with rain": "Burza z deszczem",
            "thunderstorm with heavy rain": "Burza z intensywnym deszczem"
        };
        
        function translateWeather(englishDesc) {
            return weatherTranslations[englishDesc] || englishDesc;
        }
        
        function updateCities() {
            var c = d.getElementById("c").value, y = d.getElementById("y");
            y.innerHTML = "";
            
            if (!c) {
                y.innerHTML = "<option value=''>--</option>";
                return;
            }
            
            fetch('/q?c=' + encodeURIComponent(c))
                .then(r => {
                    if (!r.ok) {
                        throw new Error('Błąd HTTP: ' + r.status);
                    }
                    return r.json();
                })
                .then(a => {
                    y.innerHTML = "<option value=''>--</option>";
                    
                    a.forEach(b => {
                        var o = d.createElement("option");
                        o.value = b;
                        o.textContent = b;
                        y.appendChild(o);
                    });
                })
                .catch(err => {
                    alert("Błąd podczas pobierania miast: " + err.message);
                });
        }
        
        function getWeather() {
            var y = d.getElementById("y").value;
            
            if (!y) {
                alert("Wybierz miasto");
                return;
            }
            
            fetch('/z?y=' + encodeURIComponent(y))
                .then(r => {
                    if (!r.ok) {
                        throw new Error('Błąd HTTP: ' + r.status);
                    }
                    return r.json();
                })
                .then(w => {
                    d.getElementById("w").style.display = "block";
                    
                    // Ustawiamy dane pogodowe
                    if (w.name) {
                        d.getElementById("l").textContent = w.name;
                    } else {
                        d.getElementById("l").textContent = 'Brak nazwy miasta';
                    }
                    
                    if (w.weather && w.weather.length > 0 && w.weather[0].description) {
                        const englishDesc = w.weather[0].description;
                        const polishDesc = translateWeather(englishDesc);
                        d.getElementById("d").textContent = polishDesc;
                    } else {
                        d.getElementById("d").textContent = 'Brak danych';
                    }
                    
                    if (w.main) {
                        if (w.main.temp !== undefined) {
                            d.getElementById("t").textContent = w.main.temp.toFixed(1);
                        } else {
                            d.getElementById("t").textContent = 'Brak danych';
                        }
                        
                        if (w.main.feels_like !== undefined) {
                            d.getElementById("f").textContent = w.main.feels_like.toFixed(1);
                        } else {
                            d.getElementById("f").textContent = 'Brak danych';
                        }
                        
                        if (w.main.humidity !== undefined) {
                            d.getElementById("h").textContent = w.main.humidity;
                        } else {
                            d.getElementById("h").textContent = 'Brak danych';
                        }
                        
                        if (w.main.pressure !== undefined) {
                            d.getElementById("p").textContent = w.main.pressure;
                        } else {
                            d.getElementById("p").textContent = 'Brak danych';
                        }
                    } else {
                        d.getElementById("t").textContent = 'Brak danych';
                        d.getElementById("f").textContent = 'Brak danych';
                        d.getElementById("h").textContent = 'Brak danych';
                        d.getElementById("p").textContent = 'Brak danych';
                    }
                    
                    if (w.wind && w.wind.speed !== undefined) {
                        d.getElementById("s").textContent = w.wind.speed;
                    } else {
                        d.getElementById("s").textContent = 'Brak danych';
                    }
                })
                .catch(e => {
                    alert("Błąd: " + e.message);
                });
        }
    </script>
</body>
</html>`

func main() {

	log.Printf("Data uruchomienia: %s | %s | %s", time.Now().Format("06-01-02"), "Autor: Ewelina Musińska",  p)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Generowanie opcji dla krajów
		opts := ""
		for c := range cs {
			opts += fmt.Sprintf("<option value=\"%s\">%s</option>", c, c)
		}
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, htmlTemplate, opts)
	})

	// Zoptymalizowany endpoint miast - krótka ścieżka
	http.HandleFunc("/q", func(w http.ResponseWriter, r *http.Request) {
		c := r.URL.Query().Get("c")
		log.Printf("Zapytanie o miasta dla kraju: %s", c)
		
		if m, ok := cs[c]; ok {
			log.Printf("Znaleziono %d miast dla kraju %s", len(m), c)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(m)
		} else {
			log.Printf("Nie znaleziono miast dla kraju: %s", c)
			http.Error(w, "404", http.StatusBadRequest)
		}
	})

	// Zoptymalizowany endpoint pogody - krótka ścieżka
	http.HandleFunc("/z", func(w http.ResponseWriter, r *http.Request) {
		y := r.URL.Query().Get("y")
		if y == "" {
			log.Printf("Otrzymano puste zapytanie o miasto")
			http.Error(w, "400", http.StatusBadRequest)
			return
		}
		
		log.Printf("Zapytanie o pogodę dla miasta: %s", y)
		u := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", y, k)
		
		resp, err := http.Get(u)
		if err != nil {
			log.Printf("Błąd podczas żądania do OpenWeatherMap: %v", err)
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("OpenWeatherMap zwrócił błąd %d", resp.StatusCode)
			http.Error(w, fmt.Sprintf("Błąd API: %d", resp.StatusCode), http.StatusServiceUnavailable)
			return
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Błąd podczas czytania odpowiedzi: %v", err)
			http.Error(w, "500", http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.Write(bodyBytes)
	})

	// Endpoint zdrowia
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "OK")
	})

	log.Printf("Serwer nasłuchuje na porcie %s", p)
	if err := http.ListenAndServe(":"+p, nil); err != nil {
		log.Fatal(err)
	}
}