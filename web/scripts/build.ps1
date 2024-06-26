# First need to enable running script for the current user
# `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`
$Env:GOOS = 'js'
$Env:GOARCH = 'wasm'
$version = git rev-parse --short HEAD
if (git status --porcelain) {
    $version = $version + "-dirty"
}

echo "Building version $version"

go build -ldflags="-X 'github.com/cbodonnell/flywheel/pkg/version.version=$version'" -o ./web/dist/flywheel-client.wasm ./cmd/client/main.go
Remove-Item Env:GOOS
Remove-Item Env:GOARCH

$goroot = go env GOROOT
cp $goroot\misc\wasm\wasm_exec.js ./web/dist/

cp -r ./web/src/* ./web/dist/
