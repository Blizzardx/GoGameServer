set CURR=%cd%
cd ../../../../../../
set GOPATH=%cd%
cd %CURR%
cd ../../InternalMessage
set GoFilePath=%cd%
cd %CURR%

go build -o gen-internal-msg.exe litgame.cn/Server/Core/Modules/tool/internal-gen
@IF %ERRORLEVEL% NEQ 0 pause


gen-internal-msg.exe --out %GoFilePath%/internal-msgid.go %GoFilePath%/server.go
@IF %ERRORLEVEL% NEQ 0 pause


