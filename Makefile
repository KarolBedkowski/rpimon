build: clean
	GOGCCFLAGS="-s -fPIC -O4 -Ofast -march=native" go build

build_pi:
	CGO_ENABLED="0" GOGCCFLAGS="-fPIC -O4 -Ofast -march=native -s" GOARCH=arm GOARM=5 go build -o rpimon
	#CGO_ENABLED="0" GOGCCFLAGS="-g -O2 -fPIC" GOARCH=arm GOARM=5 go build server.go 

clean:
	go clean
	rm -rf temp dist
	rm -f server rpimon
	find ./static -iname '*.css.gz' -exec rm -f {} ';'
	find ./static -iname '*.js.gz' -exec rm -f {} ';'
	find . -iname '*.orig' -exec rm -f {} ';'

compress:
	find ./static -iname '*.css' -exec gzip -f -k {} ';'
	find ./static -iname '*.js' -exec gzip -f -k {} ';'

install_pi: build_pi dist
	cp rpi/* dist/
	ssh k@pi sudo service k_rpimon stop
	rsync -arv --delete dist/* k@pi:rpimon/
	ssh k@pi sudo service k_rpimon start

run: clean
	mkdir temp
	go-reload server.go

certs:
	openssl genrsa 2048 > key.pem
	openssl req -new -x509 -key key.pem -out cert.pem -days 1000

debug: clean
	go build -gcflags "-N -l" server.go
	gdb ./server


dist:
	rm -rf dist
	mkdir "dist" || true
	cp -r static templates dist/
	cp *.json dist/
	cp rpimon dist/
	find dist -name *.css -exec yui-compressor -o "{}.tmp" "{}" ';' -exec  mv "{}.tmp" "{}" ';'
	find dist -name *.js -exec yui-compressor -o "{}.tmp" "{}" ';' -exec  mv "{}.tmp" "{}" ';'
	find dist -iname '*.css' -exec gzip -f -k {} ';'
	find dist -iname '*.js' -exec gzip -f -k {} ';'

