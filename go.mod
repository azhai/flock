module github.com/azhai/flock

go 1.27

require (
	github.com/JohannesKaufmann/html-to-markdown/v2 v2.5.2
	github.com/azhai/gobus v0.2.1
	github.com/azhai/goent v0.9.9-beta.3
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/labstack/echo/v5 v5.3.1
	golang.org/x/crypto v0.56.0
)

require (
	github.com/JohannesKaufmann/dom v0.3.1 // indirect
	github.com/VictoriaMetrics/fastcache v1.13.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fxamacker/cbor/v2 v2.9.3 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.11.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/x448/float16 v0.8.4 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	modernc.org/libc v1.75.7 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.12.1 // indirect
	modernc.org/sqlite v1.58.0 // indirect
)

replace (
	github.com/azhai/gobus => ../gobus
	github.com/azhai/goent => ../goent
)
