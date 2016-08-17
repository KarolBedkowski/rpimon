VERSION=`git describe --always`
DATE=`date +%Y%m%d%H%M%S`
LDFLAGS="-X k.prv/rpimon/app.AppVersion='$(VERSION)-$(DATE)'"
LDFLAGS_PI="-w -s -X k.prv/rpimon/app.AppVersion='$(VERSION)-$(DATE)'"

.PHONY: resources build

build: resources
	GOGCCFLAGS="-s -fPIC -O4 -Ofast -march=native" go build -v -ldflags $(LDFLAGS)

build_pi: resources
#	CGO_ENABLED="0" GOGCCFLAGS="-fPIC -O4 -Ofast -march=native -s" GOARCH=arm GOARM=5 go build -v -o rpimon -ldflags $(LDFLAGS_PI)
	GOGCCFLAGS="-fPIC -O4 -Ofast -march=native -pipe -mcpu=arm1176jzf-s -mfpu=vfp -mfloat-abi=hard -s" \
		CHOST="armv6j-hardfloat-linux-gnueabi" \
		CXX=arm-linux-gnueabihf-g++ CC=arm-linux-gnueabihf-gcc \
		GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 \
		go build -v --ldflags '-extldflags "-static"' --ldflags $(LDFLAGS) -o rpimon
#

clean:
	go clean
	rm -fr server rpimon dist build
	find . -iname '*.orig' -delete
	git checkout resources/resources.go

install_pi: build_pi
	ssh k@pi sudo service rpimon stop
	scp rpimon k@pi:rpimon/
	ssh k@pi sudo service rpimon start

run:
	# mkdir temp || true
	git checkout resources/resources.go
	go-reload server.go

certs:
	openssl genrsa 2048 > key.pem
	openssl req -new -x509 -key key.pem -out cert.pem -days 1000

debug: clean
	go build -gcflags "-N -l" server.go
	gdb -tui ./server -d $GOROOT


build_static:
	# create build dir if not exists
	if [ ! -d build ]; then mkdir -p "build"; fi
	cp -r templates build/
	if [ ! -e build/.stamp ]; then touch -t 200001010000 build/.stamp; fi
	# copy dir structure
	find static -type d -exec mkdir -p -- build/{} ';'
	# copy non-js and non-css files
	find static -type f ! -name *.js ! -name *.css -exec cp {} build/{} ';'
	# minify updated css
	find static -name *.css -newer build/.stamp -print -exec yui-compressor -v -o "./build/{}" "{}" ';' 
	# minify updated js
	find static -name *.js -newer build/.stamp -print -exec closure-compiler --language_in ECMASCRIPT5 --js_output_file "build/{}" --js "{}" ';' 
	# compress updated css
	if [ -x /usr/bin/zopfli ]; then \
		find build -iname '*.css' -newer build/.stamp -print -exec zopfli -v {} ';' ;  \
	else \
		find build -iname '*.css' -newer build/.stamp -print -exec gzip -f --best -k {} ';' ; \
	fi;
	# compress updated js
	if [ -x /usr/bin/zopfli ]; then \
		find build -iname '*.js' -newer build/.stamp -print -exec zopfli -v {} ';' ; \
	else \
		find build -iname '*.js' -newer build/.stamp -print -exec gzip -f --best -k {} ';' ; \
	fi;
	touch build/.stamp

resources: build_static
	go-assets-builder -p=resources -o=resources/resources.go -s="/build/" build/

deps:
	go get -d -v .
	go get -v github.com/jessevdk/go-assets-builder


dist: clean
	tar cJ -C .. \
		--exclude=.git --exclude=logs --exclude='*.log' --exclude='*.kvdb' \
		--exclude=worker-log --exclude=temp --exclude=".stamp" \
		-f ../rpimon.tar.xz rpimon
