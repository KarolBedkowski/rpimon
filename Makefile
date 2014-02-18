build:
	GOGCCFLAGS="-s -fPIC -O4 -Ofast -march=native" go build

build_pi:
	CGO_ENABLED="0" GOGCCFLAGS="-fPIC -O4 -Ofast -march=native -s" GOARCH=arm GOARM=5 go build -o rpimon
	#CGO_ENABLED="0" GOGCCFLAGS="-g -O2 -fPIC" GOARCH=arm GOARM=5 go build server.go 

clean:
	go clean
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
	ssh k@pi "(mkdir rpimon/notepad || true); (sudo service k_rpimon start)"

install: build build_static
	cp *.json dist/
	cp rpimon dist/

install_pi: build_pi build_static copy_pi

run: clean
	mkdir temp || true
	go-reload server.go

certs:
	openssl genrsa 2048 > key.pem
	openssl req -new -x509 -key key.pem -out cert.pem -days 1000

debug: clean
	go build -gcflags "-N -l" server.go
	gdb -tui ./server -d $GOROOT


build_static:
	# create dist dir if not exists
	if [ ! -d dist ]; then mkdir -p "dist"; rm -f .stamp; fi
	cp -r templates dist/
	if [ ! -e .stamp ]; then touch -t 200001010000 .stamp; fi
	# copy dir structure
	find static -type d -exec mkdir -p -- dist/{} ';'
	# copy non-js and non-css files
	find static -type f ! -name *.js ! -name *.css -exec cp {} dist/{} ';'
	# minify updated css
	find static -name *.css -newer .stamp -print -exec yui-compressor -v -o "./dist/{}" "{}" ';' 
	# minify updated js
	find static -name *.js -newer .stamp -print -exec closure-compiler --language_in ECMASCRIPT5 --js_output_file "dist/{}" --js "{}" ';' 
	# compress updated css
	find dist -iname '*.css' -newer .stamp -print -exec gzip -f --best -k {} ';'
	# compress updated js
	find dist -iname '*.js' -newer .stamp -print -exec gzip -f --best -k {} ';'
	touch .stamp

