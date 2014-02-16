build:
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

copy_pi:
	cp rpi/* dist/
	cp rpimon dist/
	ssh k@pi sudo service k_rpimon stop
	rsync -arv --delete dist/* k@pi:rpimon/
	ssh k@pi "mkdir rpimon/notepad"
	ssh k@pi sudo service k_rpimon start

install: build build_static
	cp *.json dist/
	cp rpimon dist/

install_pi: build_pi build_static copy_pi

run: clean
	mkdir temp
	go-reload server.go

certs:
	openssl genrsa 2048 > key.pem
	openssl req -new -x509 -key key.pem -out cert.pem -days 1000

debug: clean
	go build -gcflags "-N -l" server.go
	gdb -tui ./server -d $GOROOT


build_static:
	rm -rf dist/static dist/templates
	mkdir "dist" || true
	cp -r static templates dist/
	find dist -name *.css -print -exec yui-compressor -v -o "{}.tmp" "{}" ';' -exec  mv "{}.tmp" "{}" ';'
	find dist -name *.js -print -exec closure-compiler --language_in ECMASCRIPT5 --js_output_file "{}.tmp" --js "{}" ';' -exec  mv "{}.tmp" "{}" ';'
	find dist -iname '*.css' -exec gzip -f --best -k {} ';'
	find dist -iname '*.js' -exec gzip -f --best -k {} ';'

