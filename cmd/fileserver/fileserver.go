package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/yj703/simplefileserver/internal/httpfile"
	"gocloud.dev/server"
)

func health(w http.ResponseWriter, r *http.Request) {
	// This function handles all HTTP requests.
	fmt.Fprintf(w, "OK.")
}

func main() {
	fmt.Println("Starting File server...")

	srvOptions := &server.Options{
		Driver: server.NewDefaultDriver(),
	}

	// Pass the options to the Server constructor.
	srv := server.New(http.DefaultServeMux, srvOptions)

	// If your application will be behind a load balancer that handles graceful
	// shutdown of requests, you may not need to call Shutdown on the server
	// directly. If you need to ensure graceful shutdown directly, it is important
	// to have a separate goroutine, because ListenAndServe blocks indefinitely.
	go func() {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		// Receive off the chanel in a loop, because the interrupt could be sent
		// before ListenAndServe starts.
		for {
			<-interrupt
			srv.Shutdown(context.Background())
		}
	}()

	// Register a route.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "homepage... ")
	})

	http.HandleFunc("/health", health)

	http.HandleFunc("/upload", httpfile.UploadFile)
	http.HandleFunc("/delete", httpfile.DeleteFile)
	http.Handle("/download", http.StripPrefix("/download", http.FileServer(http.Dir(httpfile.UploadDir))))
	http.HandleFunc("/uploadpage", httpfile.UploadPage)
	http.HandleFunc("/downloadpage", httpfile.DownloadPage)

	// Start the server. You will see requests logged to STDOUT.
	// In the absence of an error, ListenAndServe blocks forever.
	if err := srv.ListenAndServe(":8080"); err != nil {
		log.Fatalf("%v", err)
	}

}
