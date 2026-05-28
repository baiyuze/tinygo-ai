package store

/*
#include <stdint.h>

typedef int (*spiflash_read_fn)(uint32_t, uint32_t*, uint32_t);
typedef int (*spiflash_write_fn)(uint32_t, const uint32_t*, uint32_t);
typedef int (*spiflash_erase_sector_fn)(uint32_t);
typedef int (*spiflash_unlock_fn)(void);
typedef uint32_t (*flash_size_fn)(void);

static int tinygo_flash_read(uint32_t addr, uint32_t *data, uint32_t len) {
	return ((spiflash_read_fn)0x40000a20)(addr, data, len);
}

static int tinygo_flash_write(uint32_t addr, const uint32_t *data, uint32_t len) {
	return ((spiflash_write_fn)0x40000a14)(addr, data, len);
}

static int tinygo_flash_erase_sector(uint32_t sector) {
	return ((spiflash_erase_sector_fn)0x400009fc)(sector);
}

static int tinygo_flash_unlock(void) {
	return ((spiflash_unlock_fn)0x40000a2c)();
}

static uint32_t tinygo_flash_size(void) {
	return ((flash_size_fn)0x40000af8)();
}
*/
import "C"

import (
	"runtime/interrupt"
	"unsafe"

	"esp32s3-demo/src/config"
)

const (
	sectorSize = 4096
	wordCount  = sectorSize / 4

	magic0  = 'T'
	magic1  = 'G'
	magic2  = 'A'
	magic3  = 'I'
	version = 2

	headerSizeV1 = 16
	headerSize   = 20
)

type storeError string

func (e storeError) Error() string {
	return string(e)
}

var (
	cachePresent bool
	lastError    string
)

func Load() (config.Credentials, bool, error) {
	var words [wordCount]uint32
	buf := wordsBytes(&words)

	offset, err := storageOffset()
	if err != nil {
		setStatus(false, err)
		return config.Credentials{}, false, err
	}
	if err := readSector(offset, &words); err != nil {
		setStatus(false, err)
		return config.Credentials{}, false, err
	}
	if isBlank(buf) {
		setStatus(false, nil)
		return config.Credentials{}, false, nil
	}
	if !validMagic(buf) || (buf[4] != 1 && buf[4] != version) {
		err := storeError("store record invalid")
		setStatus(false, err)
		return config.Credentials{}, false, err
	}

	ssidLen := get16(buf, 6)
	passLen := get16(buf, 8)
	keyLen := get16(buf, 10)
	promptLen := uint16(0)
	headerSizeForRecord := headerSizeV1
	checksumOffset := 12
	if buf[4] == version {
		promptLen = get16(buf, 12)
		headerSizeForRecord = headerSize
		checksumOffset = 16
	}
	total := headerSizeForRecord + int(ssidLen) + int(passLen) + int(keyLen) + int(promptLen)
	if ssidLen == 0 || ssidLen > config.MaxSSIDBytes || passLen > config.MaxPasswordBytes || keyLen > config.MaxAPIKeyBytes || promptLen > config.MaxSystemPromptBytes || total > sectorSize {
		err := storeError("store record length invalid")
		setStatus(false, err)
		return config.Credentials{}, false, err
	}
	if get32(buf, checksumOffset) != checksum(buf[:checksumOffset], buf[headerSizeForRecord:total]) {
		err := storeError("store checksum invalid")
		setStatus(false, err)
		return config.Credentials{}, false, err
	}

	pos := headerSizeForRecord
	cfg := config.Credentials{
		SSID:         string(buf[pos : pos+int(ssidLen)]),
		Password:     string(buf[pos+int(ssidLen) : pos+int(ssidLen)+int(passLen)]),
		APIKey:       string(buf[pos+int(ssidLen)+int(passLen) : pos+int(ssidLen)+int(passLen)+int(keyLen)]),
		SystemPrompt: string(buf[pos+int(ssidLen)+int(passLen)+int(keyLen) : total]),
	}
	setStatus(true, nil)
	return cfg, true, nil
}

func Save(cfg config.Credentials) error {
	if len(cfg.SSID) == 0 {
		return setStatus(true, storeError("ssid empty"))
	}
	if len(cfg.SSID) > config.MaxSSIDBytes {
		return setStatus(true, storeError("ssid too long"))
	}
	if len(cfg.Password) > config.MaxPasswordBytes {
		return setStatus(true, storeError("password too long"))
	}
	if len(cfg.APIKey) > config.MaxAPIKeyBytes {
		return setStatus(true, storeError("api key too long"))
	}
	if len(cfg.SystemPrompt) > config.MaxSystemPromptBytes {
		return setStatus(true, storeError("system prompt too long"))
	}

	offset, err := storageOffset()
	if err != nil {
		return setStatus(true, err)
	}

	var words [wordCount]uint32
	buf := wordsBytes(&words)
	for i := range buf {
		buf[i] = 0xff
	}
	buf[0], buf[1], buf[2], buf[3] = magic0, magic1, magic2, magic3
	buf[4] = version
	put16(buf, 6, uint16(len(cfg.SSID)))
	put16(buf, 8, uint16(len(cfg.Password)))
	put16(buf, 10, uint16(len(cfg.APIKey)))
	put16(buf, 12, uint16(len(cfg.SystemPrompt)))

	pos := headerSize
	pos += copy(buf[pos:], cfg.SSID)
	pos += copy(buf[pos:], cfg.Password)
	pos += copy(buf[pos:], cfg.APIKey)
	pos += copy(buf[pos:], cfg.SystemPrompt)
	put32(buf, 16, checksum(buf[:16], buf[headerSize:pos]))

	if err := writeSector(offset, &words); err != nil {
		return setStatus(true, err)
	}
	return setStatus(true, nil)
}

func Status() (bool, string) {
	return cachePresent, lastError
}

func storageOffset() (uint32, error) {
	size := uint32(C.tinygo_flash_size())
	if size < 2*1024*1024 || size == 0xffffffff {
		return 0, storeError("flash size invalid")
	}
	return size - sectorSize, nil
}

func readSector(offset uint32, words *[wordCount]uint32) error {
	state := interrupt.Disable()
	res := C.tinygo_flash_read(C.uint32_t(offset), (*C.uint32_t)(unsafe.Pointer(&words[0])), C.uint32_t(sectorSize))
	interrupt.Restore(state)
	if res != 0 {
		return storeError("flash read failed")
	}
	return nil
}

func writeSector(offset uint32, words *[wordCount]uint32) error {
	state := interrupt.Disable()
	C.tinygo_flash_unlock()
	res := C.tinygo_flash_erase_sector(C.uint32_t(offset / sectorSize))
	if res == 0 {
		C.tinygo_flash_unlock()
		res = C.tinygo_flash_write(C.uint32_t(offset), (*C.uint32_t)(unsafe.Pointer(&words[0])), C.uint32_t(sectorSize))
	}
	interrupt.Restore(state)
	if res != 0 {
		return storeError("flash write failed")
	}
	return nil
}

func wordsBytes(words *[wordCount]uint32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&words[0])), sectorSize)
}

func isBlank(buf []byte) bool {
	for _, b := range buf {
		if b != 0xff {
			return false
		}
	}
	return true
}

func validMagic(buf []byte) bool {
	return buf[0] == magic0 && buf[1] == magic1 && buf[2] == magic2 && buf[3] == magic3
}

func checksum(parts ...[]byte) uint32 {
	sum := uint32(2166136261)
	for _, part := range parts {
		for _, b := range part {
			sum ^= uint32(b)
			sum *= 16777619
		}
	}
	return sum
}

func get16(buf []byte, off int) uint16 {
	return uint16(buf[off]) | uint16(buf[off+1])<<8
}

func put16(buf []byte, off int, value uint16) {
	buf[off] = byte(value)
	buf[off+1] = byte(value >> 8)
}

func get32(buf []byte, off int) uint32 {
	return uint32(buf[off]) | uint32(buf[off+1])<<8 | uint32(buf[off+2])<<16 | uint32(buf[off+3])<<24
}

func put32(buf []byte, off int, value uint32) {
	buf[off] = byte(value)
	buf[off+1] = byte(value >> 8)
	buf[off+2] = byte(value >> 16)
	buf[off+3] = byte(value >> 24)
}

func setStatus(present bool, err error) error {
	cachePresent = present
	if err != nil {
		lastError = err.Error()
		return err
	}
	lastError = ""
	return nil
}
