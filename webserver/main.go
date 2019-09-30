package main

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/stianeikeland/go-rpio"
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

func playSound(sound string, level int) error {
	file := getFilePath() + sound + ".mp3"
	cmd := os.Getenv("SNAP") + "/bin/client-wrapper"
	args := []string{"usr/bin/mpg123.bin", "-o", "pulse", "-q", "--scale", strconv.Itoa(level), file}
	err := exec.Command(cmd, args...).Run()
	return err
}

func setVolume(level int) error {
	cmd := os.Getenv("SNAP") + "/bin/client-wrapper"
	args := []string{"usr/bin/pactl", "set-sink-volume", "0", strconv.Itoa(level) + "%"}
	err := exec.Command(cmd, args...).Run()
	return err
}

func enableSoundCard() error {
	// Use mcu pin 26, corresponds to physical pin 37 on the pi
	pin := rpio.Pin(26)

	// Open and map memory to access gpio, check for errors
	err := rpio.Open()
	if err != nil {
		return err
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pin to output mode
	pin.Output()

	// Turn on the pin
	pin.High()
	return err
}

func play(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	// get the sound param
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
	// validate the sound
	soundMatched, _ := regexp.MatchString(`^[a-z]+(_[a-z]+)*$`, sound)
	if !soundMatched {
		w.Write([]byte("Sound name is invalid, must be underscore case"))
		return
	}

	// get the level param
	levelParam, ok := r.URL.Query()["level"]
	if !ok || len(levelParam[0]) < 1 {
		w.Write([]byte("Url Param 'level' is missing"))
		return
	}
	// Query()["level"] will return an array of items,
	// we only want the single item.
	levelString := string(levelParam[0])
	if err := assertFile(levelString); err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	// validate the level
	levelMatched, _ := regexp.MatchString(`^(100|[1-9][0-9]|[0-9])$`, levelString)
	if !levelMatched {
		w.Write([]byte("Level must be between 0 and 100"))
		return
	}
	level, err := strconv.Atoi(levelString)
	if err != nil {
		panic(err)
	}

	if err := playSound(sound, level); err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte(sound + " OK"))
}

func volume(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)

	// get the level param
	levelParam, ok := r.URL.Query()["level"]
	if !ok || len(levelParam[0]) < 1 {
		w.Write([]byte("Url Param 'level' is missing"))
		return
	}
	// Query()["level"] will return an array of items,
	// we only want the single item.
	levelString := string(levelParam[0])
	// validate the level
	levelMatched, _ := regexp.MatchString(`^(100|[1-9][0-9]|[0-9])$`, levelString)
	if !levelMatched {
		w.Write([]byte("Level must be between 0 and 100"))
		return
	}
	level, err := strconv.Atoi(levelString)
	if err != nil {
		panic(err)
	}

	if err := setVolume(level); err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("volume set to " + levelString + "% OK"))
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

	if err := assertDirectoryExists(); err != nil {
		panic(err)
	}

	if err := enableSoundCard(); err != nil {
		panic(err)
	}

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

func assertDirectoryExists() error {
	path := getFilePath()
	// make the directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = nil
		println("Creating folder " + path)
		os.Mkdir(path, os.ModePerm)
	}
	return err
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
