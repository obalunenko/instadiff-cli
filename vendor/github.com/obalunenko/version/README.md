# version

This package contains build information generated at build time and compiled into the binary.

```go
// Package version contains build information such as the git commit, app version, build date.
//
// This info generated at build time and compiled into the binary.
//
// Usage:
// At your build script add following lines:
// go install -ldflags '-X github.com/obalunenko/version.version APP_VERSION -X github.com/obalunenko/version.builddate BUILD_DATE -X github.com/obalunenko/version.commit COMMIT -X github.com/obalunenko/version.shortcommit SHORTCOMMIT -X github.com/obalunenko/version.appname APP_NAME'
// and then build your binary
// go build
// Please note that all future binaries will be compiled with the embedded information unless the version package is recompiled with new values.
//
// Alternative is use ldflags on stage of compiling:
// GOVERSION=$(go version | awk '{print $3;}')
// APP=myapp
// BIN_OUT=bin/${APP}
// BUILDINFO_VARS_PKG=github.com/obalunenko/version
// GO_BUILD_LDFLAGS="-s -w \
//-X ${BUILDINFO_VARS_PKG}.version=${VERSION} \
//-X ${BUILDINFO_VARS_PKG}.commit=${COMMIT} \
//-X ${BUILDINFO_VARS_PKG}.shortcommit=${SHORTCOMMIT} \
//-X ${BUILDINFO_VARS_PKG}.builddate=${DATE} \
//-X ${BUILDINFO_VARS_PKG}.appname=${APP}" \
//-X ${BUILDINFO_VARS_PKG}.goversion=${GOVERSION}"
// GO_BUILD_PACKAGE="<PATH to package where main.go located>"

// rm -rf ${BIN_OUT}
//
// go build -o ${BIN_OUT} -a -ldflags "${GO_BUILD_LDFLAGS}" "${GO_BUILD_PACKAGE}"
```
