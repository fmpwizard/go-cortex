package services

import (
	"log"
	"os/exec"
	"time"
)

func RecordCommand() (WitMessage, error) {
	log.Println("about to record")
	// the parameters "-t", "alsa", "hw:1,0" are for the raspberry pi, as it uses a USB microphone
	cmd := exec.Command("sox", "-t", "alsa", "hw:1,0", "-b", "16", "-c", "1", "-r", "16k", "command.wav", "silence", "1", "0.1", "3%", "1", "3.0", "3%")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Please say something...")

	commandDone := make(chan error)
	go func() {
		commandDone <- cmd.Wait()
	}()

	select {
	case <-time.After(7 * time.Second):
		if err := cmd.Process.Kill(); err != nil {
			log.Fatal("failed to kill: ", err)
		}
		<-commandDone // allow goroutine to exit
		log.Println("process killed")
	case err := <-commandDone:
		if err != nil {
			log.Printf("process done with error = %v", err)
			return
		}
	}

	intent, err := FetchVoiceIntent("command.wav")
	if err != nil {
		log.Printf("We got error: '%v' while fetching wit intent using voice command", err)
	}
	return intent, err

}
