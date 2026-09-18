package main

import (
	"fmt"
	"khairul169/garage-webui/router"
	"khairul169/garage-webui/ui"
	"khairul169/garage-webui/utils"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {
	// Initialize app
	godotenv.Load()
	utils.InitCacheManager()
	sessionMgr := utils.InitSessionManager()

	if err := utils.Garage.LoadConfig(); err != nil {
		log.Println("Cannot load garage config!", err)
	}

	basePath := os.Getenv("BASE_PATH")
	mux := http.NewServeMux()

	// Serve API
	apiPrefix := basePath + "/api"
	mux.Handle(apiPrefix+"/", http.StripPrefix(apiPrefix, router.HandleApiRouter()))

	// Static files
	ui.ServeUI(mux)

	// Redirect to UI if BASE_PATH is set
	if basePath != "" {
		mux.Handle("/", http.RedirectHandler(basePath, http.StatusMovedPermanently))
	}

	handler := sessionMgr.LoadAndSave(mux)

	// Listen on a unix domain socket if SOCKET_PATH is set
	if socketPath := os.Getenv("SOCKET_PATH"); len(socketPath) > 0 {
		listener, err := listenUnix(socketPath, os.Getenv("SOCKET_MODE"))
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Starting server on unix:%s", socketPath)

		if err := http.Serve(listener, handler); err != nil {
			log.Fatal(err)
		}
		return
	}

	host := utils.GetEnv("HOST", "0.0.0.0")
	port := utils.GetEnv("PORT", "3909")

	addr := fmt.Sprintf("%s:%s", host, port)
	log.Printf("Starting server on http://%s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

// listenUnix listens on the given unix socket path, replacing a stale socket
// left behind by a previous run, and applies the optional octal file mode.
func listenUnix(path string, mode string) (net.Listener, error) {
	if info, err := os.Stat(path); err == nil {
		if info.Mode()&os.ModeSocket == 0 {
			return nil, fmt.Errorf("%s exists and is not a socket", path)
		}
		if err := os.Remove(path); err != nil {
			return nil, err
		}
	}

	listener, err := net.Listen("unix", path)
	if err != nil {
		return nil, err
	}

	if len(mode) > 0 {
		perm, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			listener.Close()
			return nil, fmt.Errorf("invalid SOCKET_MODE %q: %w", mode, err)
		}
		if err := os.Chmod(path, os.FileMode(perm)); err != nil {
			listener.Close()
			return nil, err
		}
	}

	return listener, nil
}
