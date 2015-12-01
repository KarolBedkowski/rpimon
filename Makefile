VERSION=`git describe --always`
DATE=`date`
LDFLAGS="-X k.prv/rpimon/app.AppVersion '$(VERSION) - $(DATE)'"

.PHONY: resources build

build: resources
	GOGCCFLAGS="-s -fPIC -O4 -Ofast -march=native" go build -ldflags $(LDFLAGS)

build_pi: resources
	CGO_ENABLED="0" GOGCCFLAGS="-fPIC -O4 -Ofast -march=native -s" GOARCH=arm GOARM=5 go build -o rpimon -ldflags $(LDFLAGS)
	#CGO_ENABLED="0" GOGCCFLAGS="-g -O2 -fPIC" GOARCH=arm GOARM=5 go build server.go 

clean:
	go clean
	rm -fr server rpimon dist build
	find . -iname '*.orig' -delete
	git checkout resources/resources.go

install_pi: build_pi
	ssh k@pi sudo service k_rpimon stop
	scp rpimon k@pi:rpimon/
	ssh k@pi sudo service k_rpimon start

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
	find build -iname '*.css' -newer build/.stamp -print -exec gzip -f --best -k {} ';'
	# compress updated js
	find build -iname '*.js' -newer build/.stamp -print -exec gzip -f --best -k {} ';'
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
