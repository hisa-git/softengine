package main

import (
	"fmt"
	"log"
	"time"

	"github.com/striter-no/softengine/sounds"
	"github.com/ungerik/go3d/vec3"
)

func main() {
	sys, err := sounds.NewSoundSystem(vec3.T{0, 0, 0})
	if err != nil {
		panic(err)
	}
	defer sys.End()

	// sys.SetMasterVolume(2.0)

	stepsID := sys.AddSpeaker("./assets/sounds/wind.mp3", 1, 1)
	if stepsID == -1 {
		log.Fatal("Failed to load sound")
	}

	stepsID2 := sys.AddSpeaker("./assets/sounds/gunshot.mp3", 0, 8)
	if stepsID2 == -1 {
		log.Fatal("Failed to load sound")
	}

	sys.ChangeIDPosition(stepsID, vec3.T{-3, 0, 2})
	sys.ChangeIDPosition(stepsID2, vec3.T{3, 0, 2})

	fmt.Print("Playing...")
	sys.PlayID(stepsID)

	for i := range 300 {
		sys.UpdateListener(vec3.T{float32(i) * 0.1, 0, 0})
		time.Sleep(50 * time.Millisecond)

		if i%10 == 0 {
			sys.PlayID(stepsID2)
		}
	}

	sys.StopID(stepsID)
}
