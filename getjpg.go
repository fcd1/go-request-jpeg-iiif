package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"strconv"
	"encoding/json"
)

const ConfigFile = "config/config.json"

type Settings struct {
	Scheme string `json:"scheme"`
	Server string `json:"server"`
	Prefix string `json:"prefix"`
	Region string `json:"region"`
	Size string `json:"size"`
	Rotation string `json:"rotation"`
	Quality string `json:"quality"`
	Format string `json:"format"`
	ImageInfo string `json:"image_info"`
	PidFile string `json:"pid_file"`
	DelayInMs int `json:"delay_in_ms"`
	ImageDir string `json:"image_dir"`
}

var settings *Settings = read_settings()

func read_settings() *Settings {
	var settings Settings
	
	config_file, err := os.Open(ConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	json_decoder := json.NewDecoder(config_file)
	err = json_decoder.Decode(&settings)
	return &settings
}

func setup_date_string() string {
	year, month, day := time.Now().Date()
	date_string := strconv.Itoa(year) +
		month.String() +
		strconv.Itoa(day) +
		"_" +
		fmt.Sprintf("%02d",time.Now().Hour()) +
		fmt.Sprintf("%02d",time.Now().Minute()) +
		fmt.Sprintf("%02d",time.Now().Second())
	return date_string
}

func log_settings() {

	log.Print("Here are the settings (i.e. constants in code):")
	log.Print("Scheme: " + settings.Scheme);
	log.Print("Server: " + settings.Server);
	log.Print("Prefix: " + settings.Prefix)
	log.Print("Region: " + settings.Region)
	log.Print("Size: " + settings.Size)
	log.Print("Rotation: " + settings.Rotation)
	log.Print("Quality: " + settings.Quality)
	log.Print("Format: " + settings.Format)
	log.Print("ImageInfo: " + settings.ImageInfo)
	log.Print("PidFile: " + settings.PidFile)
	log.Print("DelayInMs: " + strconv.Itoa(settings.DelayInMs))
	log.Print("ImageDir: " + settings.ImageDir)
}

func main() {

	// Read settings from config file
	settings := read_settings()
	
	// Setup date_string, used to create log filename and image subdir
	date_string := setup_date_string()

	// Setup log file for logging
	log_filename := date_string + ".log"
	log_file, err := os.Create("logs/"+log_filename)
	defer log_file.Close()
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(log_file)
	fmt.Println("Output sent to following logfile:")
	fmt.Println(log_file.Name())

	// Print settings to logfile
	log_settings()

	// Read in the image pids
	pid_file, err := os.Open(settings.PidFile)
	defer pid_file.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Create and populate a slice containing the pids
	pids := make([]string,0,50)
	scanner := bufio.NewScanner(pid_file)
	for scanner.Scan() {
		pid := scanner.Text()
		pids = append(pids, pid)
	}

	// Generate path to image dir
	image_dir := settings.ImageDir + date_string + "/"
	err = os.Mkdir(image_dir,os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Images will be saved in " + image_dir)
	
	// Main processing
	log.Print("Processing started, " + strconv.Itoa(len(pids)) + " pids.")

	for _, pid := range pids {

		log.Print("About to process pid " + pid)
		// formated_pid will be used in the name of the generated files
		formated_pid := strings.Replace(pid,":","_",1)

		// Construct the URL to retrieve the jpg from the image server
		image_url := settings.Scheme + "://" + settings.Server + "/" + settings.Prefix + "/" +
			pid + "/" + settings.Region + "/" + settings.Size + "/" + settings.Rotation +
			"/" + settings.Quality + "." + settings.Format
		log.Print(image_url)

		// Send GET request to server
		resp, err := http.Get(image_url)
		defer resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(resp.Status)
		log.Print(resp.Status)

		// if Status Code is 200, create the file
		if resp.StatusCode == 200 {
			image_filename := image_dir + formated_pid + ".jpg"
			image_file, err := os.Create(image_filename)
			defer image_file.Close()
			if err != nil {
				log.Fatal(err)
			}
			io.Copy(image_file, resp.Body)
			log.Print("Generated file: " + image_filename)
		} else {
			log.Print("File not created due to non-200 status code")
		}

		// Sleep to lessen burden on server
		log.Print("Sleeping for " + strconv.Itoa(settings.DelayInMs) + "ms")
		time.Sleep(time.Duration(settings.DelayInMs) * time.Millisecond)

		// Get JSON info, not really required
		/*
		info_url := settings.Scheme + "://" + settings.Server + "/" + settings.Prefix + "/" +
			pid + "/" + settings.ImageInfo
		fmt.Println(info_url)
		resp, err = http.Get(info_url)
		defer resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}

		// fmt.Println(resp, err)
		json_file, err := os.Create("info/"+formated_pid+".json")
		defer json_file.Close()
		if err != nil {
			log.Fatal(err)
		}
		io.Copy(json_file, resp.Body)
                */
	}
}
