simple http server for upload and download file.


# build

install golang
clone the repo

     go build -o . ./...


# run

    fileserver.exe
    ./fileserver


# access upload/download page

    http://localhost:8080/uploadpage
    http://localhost:8080/downloadpage


# upload file from unix/linx command

    curl -X POST http://localhost:8080/upload -F "file=xxx"
    curl -X POST http://localhost:8080/upload -F "file=xxx" -F "setfilename=yyy" -F "dir=xxx"


# upload file from windows powershell
        $wc = New-Object System.Net.WebClient
        $resp = $wc.UploadFile("http://localhost:8080/upload","xxx")

