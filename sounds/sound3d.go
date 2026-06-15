package sounds

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/ungerik/go3d/vec3"
)

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: ${SRCDIR}/../cgo/audio_wrapper.o -lm
#include "../cgo/audio_wrapper.h"
#include <stdlib.h>
*/
import "C"

type Speaker3D struct {
	Position  vec3.T
	SoundID   int
	Volume    float32
	sounds    []unsafe.Pointer
	nextVoice int
}

type SoundSystem struct {
	Listener vec3.T
	speakers map[int]*Speaker3D
	mu       sync.RWMutex
	nextID   int
}

func NewSoundSystem(listener vec3.T) (*SoundSystem, error) {
	if C.audio_init() != 0 {
		return nil, fmt.Errorf("failed to initialize audio engine")
	}

	C.audio_set_listener_position(
		C.float(listener[0]),
		C.float(listener[1]),
		C.float(listener[2]),
	)

	return &SoundSystem{
		Listener: listener,
		speakers: make(map[int]*Speaker3D),
		nextID:   1,
	}, nil
}

func (s *SoundSystem) AddSpeaker(path string, looped int, poolSize int) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	if poolSize < 1 {
		poolSize = 1
	}

	cpath := C.CString(path)
	defer C.free(unsafe.Pointer(cpath))

	sounds := make([]unsafe.Pointer, poolSize)
	for i := 0; i < poolSize; i++ {
		ptr := C.audio_load_sound(cpath, C.int(looped))
		if ptr == nil {
			for j := 0; j < i; j++ {
				C.audio_delete_sound(sounds[j])
			}
			return -1
		}
		sounds[i] = ptr
	}

	id := s.nextID
	s.nextID++

	speaker := &Speaker3D{
		Position:  vec3.T{0, 0, 0},
		SoundID:   id,
		Volume:    1.0,
		sounds:    sounds,
		nextVoice: 0,
	}

	s.speakers[id] = speaker
	return id
}

func (s *SoundSystem) PlayID(id int) {
	s.mu.RLock()
	speaker, ok := s.speakers[id]
	s.mu.RUnlock()

	if !ok {
		return
	}

	var targetVoice unsafe.Pointer = nil
	for _, ptr := range speaker.sounds {
		if C.audio_is_playing(ptr) == 0 {
			targetVoice = ptr
			break
		}
	}

	if targetVoice == nil {
		targetVoice = speaker.sounds[speaker.nextVoice]
		speaker.nextVoice = (speaker.nextVoice + 1) % len(speaker.sounds)
		C.audio_stop(targetVoice)
		C.audio_seek_to_start(targetVoice)
	}

	C.audio_set_position(targetVoice,
		C.float(speaker.Position[0]),
		C.float(speaker.Position[1]),
		C.float(speaker.Position[2]))
	C.audio_set_volume(targetVoice, C.float(speaker.Volume))
	C.audio_play(targetVoice)
}

func (s *SoundSystem) StopID(id int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if speaker, ok := s.speakers[id]; ok {
		for _, ptr := range speaker.sounds {
			C.audio_stop(ptr)
		}
	}
}

func (s *SoundSystem) SetVolume(id int, volume float32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if speaker, ok := s.speakers[id]; ok {
		speaker.Volume = volume
		for _, ptr := range speaker.sounds {
			C.audio_set_volume(ptr, C.float(volume))
		}
	}
}

func (s *SoundSystem) SetMasterVolume(volume float32) {
	C.audio_set_master_volume(C.float(volume))
}

func (s *SoundSystem) DeleteID(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if speaker, ok := s.speakers[id]; ok {
		for _, ptr := range speaker.sounds {
			C.audio_stop(ptr)
			C.audio_delete_sound(ptr)
		}
		delete(s.speakers, id)
	}
}

func (s *SoundSystem) End() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, speaker := range s.speakers {
		for _, ptr := range speaker.sounds {
			C.audio_stop(ptr)
			C.audio_delete_sound(ptr)
		}
		delete(s.speakers, id)
	}

	C.audio_cleanup()
}

func (s *SoundSystem) UpdateListener(listener vec3.T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Listener = listener
	C.audio_set_listener_position(
		C.float(listener[0]),
		C.float(listener[1]),
		C.float(listener[2]),
	)
}

func (s *SoundSystem) ChangeIDPosition(id int, pos vec3.T) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if speaker, ok := s.speakers[id]; ok {
		speaker.Position = pos
		for _, ptr := range speaker.sounds {
			C.audio_set_position(ptr,
				C.float(pos[0]),
				C.float(pos[1]),
				C.float(pos[2]))
		}
	}
}
