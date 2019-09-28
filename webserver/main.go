package main

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

func maxClients(h http.Handler, n int) http.Handler {
	sema := make(chan struct{}, n)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sema <- struct{}{}
		defer func() { <-sema }()

		h.ServeHTTP(w, r)
	})
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
	w.Write([]byte("OK"))
}

func playSound(sound string, level string) error {
	file := getFilePath() + sound + ".mp3"
	cmd := os.Getenv("SNAP") + "/bin/client-wrapper"
	args := []string{"usr/bin/mpg123.bin", "-o", "pulse", "-q", "--scale", level, file}
	err := exec.Command(cmd, args...).Run()
	return err
}

// level should be a percentage, i.e. "26%"
func setVolume(level string) error {
	cmd := "/snap/bin/pactl"
	args := []string{"set-sink-volume", "0", level + "%"}
	err := exec.Command(cmd, args...).Run()
	return err
}

func play(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	soundParam, ok := r.URL.Query()["sound"]

	if !ok || len(soundParam[0]) < 1 {
		w.Write([]byte("Url Param 'sound' is missing"))
		return
	}

	// Query()["sound"] will return an array of items,
	// we only want the single item.
	sound := string(soundParam[0])
	if err := assertFile(sound); err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	if err := playSound(sound, "100"); err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte(sound + " OK"))
}

func volume(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	levelParam, ok := r.URL.Query()["level"]

	if !ok || len(levelParam[0]) < 1 {
		w.Write([]byte("Url Param 'level' is missing, level should be a value between 0 and 100"))
		return
	}

	// Query()["level"] will return an array of items,
	// we only want the single item.
	level := string(levelParam[0])

	if err := setVolume(level); err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte(level + " OK"))
}

func assertFile(sound string) error {
	file := getFilePath() + sound + ".mp3"
	_, err := os.Stat(file)
	// if the sound does not exist
	if os.IsNotExist(err) {
		// then download it
		if err := DownloadSound(sound); err != nil {
			return err
		}
		err = nil
	}
	return err
}

func main() {
	// start web server
	println("Starting Server")

	assertDirectoryExists()

	playHandler := http.HandlerFunc(play)
	http.Handle("/play", maxClients(playHandler, 20))

	volumeHandler := http.HandlerFunc(volume)
	http.Handle("/volume", maxClients(volumeHandler, 1))

	statusHandler := http.HandlerFunc(getStatus)
	http.Handle("/status", maxClients(statusHandler, 5))

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

	println("Ready")

}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
}

func getFilePath() string {
	return os.Getenv("SNAP_COMMON") + "/sounds/"
}

func assertDirectoryExists() {
	path := getFilePath()
	// make the directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		println("Creating folder " + path)
		os.Mkdir(path, os.ModePerm)
	}
}

// DownloadSound will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadSound(sound string) error {
	url := "http://sounds.codeverse.com/" + sound + ".mp3"
	file := getFilePath() + sound + ".mp3"

	println("Downloading " + url)

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
