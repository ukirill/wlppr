echo "Building Wlppr!"
rsrc -manifest ./cmd/wlppr.exe.manifest -ico .cmd/resources/icon.ico -o ./cmd/wlppr.syso
cd ./cmd/
go build -ldflags="-H windowsgui" -o wlppr.exe -v