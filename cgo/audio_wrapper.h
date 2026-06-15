#ifndef AUDIO_WRAPPER_H
#define AUDIO_WRAPPER_H

#ifdef __cplusplus
extern "C" {
#endif

int audio_init();
void* audio_load_sound(const char* path, int loop);
void audio_set_position(void* sound_ptr, float x, float y, float z);
void audio_play(void* sound_ptr);
void audio_stop(void* sound_ptr);
void audio_delete_sound(void* sound_ptr);
void audio_set_listener_position(float x, float y, float z);
void audio_set_listener_direction(float x, float y, float z);
void audio_cleanup();

void audio_set_volume(void* sound_ptr, float volume);
void audio_set_master_volume(float volume);

int audio_is_playing(void* sound_ptr);
void audio_seek_to_start(void* sound_ptr);

#ifdef __cplusplus
}
#endif

#endif
