mkdir .\distribution

go get github.com/GeertJohan/go.rice
go get github.com/GeertJohan/go.rice/rice

go build
rice append --exec .\D3pixelbot.exe

[Net.ServicePointManager]::SecurityProtocol = "tls12, tls11, tls"
Invoke-WebRequest -Uri https://github.com/c-smile/sciter-sdk/raw/master/bin/64/sciter.dll -OutFile .\sciter.dll

7z a -t7z .\distribution\Windows.x86-64.7z -m0=lzma2 -mx=9 -aoa D3pixelbot.exe README.md LICENSE config.json sciter.dll

Remove-Item .\D3pixelbot.exe, .\sciter.dll