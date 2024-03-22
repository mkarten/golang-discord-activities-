GOOS=js GOARCH=wasm go build -o ./public/wasm.wasm
cp ./public/wasm.wasm ../App-Client-server/public/wasm.wasm