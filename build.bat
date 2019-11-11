@echo "Building Wlppr!"
rsrc -manifest ./cmd/resource/wlppr.exe.manifest -ico ./cmd/resource/icon.ico -o ./cmd/wlppr.syso
@cd ./cmd/
go build -ldflags="-H windowsgui" -o ../bin/wlppr.exe -v
@del wlppr.syso
@cd ..