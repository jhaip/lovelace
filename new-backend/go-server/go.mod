module github.com/jhaip/lovelace

go 1.12

require (
	github.com/alecthomas/repr v0.0.0-20181024024818-d37bc2a10ba1
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/jung-kurt/gofpdf v1.5.2
	github.com/kokardy/listing v0.0.0-20140516154625-795534c33c5a
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mattn/go-ciede2000 v0.0.0-20170301095244-782e8c62fec3
	github.com/mattn/go-sqlite3 v1.10.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pebbe/zmq4 v1.0.0
	github.com/stretchr/testify v1.3.0 // indirect
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/uber/jaeger-client-go v2.16.0+incompatible
	github.com/uber/jaeger-lib v2.0.0+incompatible // indirect
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20190912160710-24e19bdeb0f2 // indirect
	room/roomupdate v0.0.0
	zombiezen.com/go/capnproto2 v2.17.0+incompatible
)

replace room/roomupdate => ./roomupdate
