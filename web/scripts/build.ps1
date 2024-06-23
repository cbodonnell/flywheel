# First need to enable running script for the current user
# `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`
$Env:GOOS = 'js'
$Env:GOARCH = 'wasm'
go build -ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=wasm-test'" -o ./dist/flywheel-client.wasm ./cmd/client/main.go
Remove-Item Env:GOOS
Remove-Item Env:GOARCH

$goroot = go env GOROOT
cp $goroot\misc\wasm\wasm_exec.js ./dist/

cp -r ./web/src/* ./dist/
