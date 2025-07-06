package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/pbkdf2"
)

func DeriveKey(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, 10000, 32, sha256.New)
}

func Hash(password, salt []byte) ([]byte, error) {
	if len(salt) < 16 {
		return nil, errors.New("salt too short")
	}
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32), nil
}

func GenerateSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	return salt, err
}

func VerifyPassword(password []byte, storedHash, salt []byte) (bool, error) {
	derivedHash, err := Hash(password, salt)

	if err != nil {
		return false, err
	}

	return subtle.ConstantTimeCompare(derivedHash, storedHash) == 1, nil
}

func EncryptFile(path string, outPath string, newName string, masterKey string) (error, File) {
	config, err := ReadConfig()
	if err != nil {
		fmt.Printf("ERROR: Failed to grab config: %v\n", err)
		return err, File{}
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("ERROR: File not found: %s\n", path)
		} else {
			fmt.Printf("ERROR: Getting file info: %v\n", err)
		}
		return err, File{}
	}

	fileData, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("ERROR: Reading file: %v\n", err)
		return fmt.Errorf("reading file %w", err), File{}
	}

	key := DeriveKey([]byte(masterKey), config.Salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("ERROR: Creating cipher: %v\n", err)
		return fmt.Errorf("creating cipher: %w", err), File{}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Printf("ERROR: Creating GCM: %v\n", err)
		return fmt.Errorf("creating GCM: %w", err), File{}
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Printf("ERROR: Generating nonce: %v\n", err)
		return fmt.Errorf("generating nonce: %w", err), File{}
	}

	encrypted := gcm.Seal(nonce, nonce, fileData, nil)

	id := uuid.New().String()
	outputPath := fmt.Sprintf("%s/%s.enc", outPath, id)

	if err := os.MkdirAll(outPath, 0755); err != nil {
		fmt.Printf("ERROR: Creating output directory: %v\n", err)
		return fmt.Errorf("creating output directory: %w", err), File{}
	}

	err = os.WriteFile(outputPath, encrypted, 0644)
	if err != nil {
		fmt.Printf("ERROR: Writing encrypted file: %v\n", err)
		return fmt.Errorf("writing encrypted file: %w", err), File{}
	}

	fileExtension := filepath.Ext(path)
	mimeTypeByExtension := mime.TypeByExtension(fileExtension)

	if mimeTypeByExtension == "" {
		mimeTypeByExtension = "application/octet-stream"
	}

	fileName := newName

	if fileName == "" {
		fileName = fmt.Sprintf("%s.%s", fileInfo.Name(), fileExtension) // because I don't want to force people to manually insert the file-ext
	}

	fileRecord := File{
		Id:           id,
		OriginalName: fileName,
		OriginalPath: path,
		DateAdded:    time.Now(),
		MimeType:     mimeTypeByExtension,
		Extension:    fileExtension,
	}

	return nil, fileRecord
}

func DecryptFile(inputPath, outputPath, masterKey string) error {
	config, err := ReadConfig()
	if err != nil {
		fmt.Printf("ERROR: Reading config: %v\n", err)
		return fmt.Errorf("reading config: %w", err)
	}

	encryptedData, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Printf("ERROR: Reading encrypted file: %v\n", err)
		return fmt.Errorf("reading encrypted file: %w", err)
	}

	key := DeriveKey([]byte(masterKey), config.Salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Printf("ERROR: Creating cipher: %v\n", err)
		return fmt.Errorf("creating cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Printf("ERROR: Creating GCM: %v\n", err)
		return fmt.Errorf("creating GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		fmt.Printf("ERROR: Encrypted data too short. Got %d bytes, need at least %d\n", len(encryptedData), nonceSize)
		return fmt.Errorf("encrypted data too short")
	}

	nonce := encryptedData[:nonceSize]
	encrypted := encryptedData[nonceSize:]

	decryptedData, err := gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		fmt.Printf("ERROR: Decryption failed: %v\n", err)
		return fmt.Errorf("decryption failed: %w", err)
	}

	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("ERROR: Creating output directory: %v\n", err)
		return fmt.Errorf("creating output directory: %w", err)
	}

	err = os.WriteFile(outputPath, decryptedData, 0644)
	if err != nil {
		fmt.Printf("ERROR: Writing decrypted file: %v\n", err)
		return fmt.Errorf("writing decrypted file: %w", err)
	}

	return nil
}

func Encrypt(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	encrypted := gcm.Seal(nonce, nonce, data, nil)
	return encrypted, nil
}

func Decrypt(encrypted []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encrypted) < nonceSize {
		return nil, fmt.Errorf("encrypted too short")
	}

	nonce, encrypted := encrypted[:nonceSize], encrypted[nonceSize:]
	return gcm.Open(nil, nonce, encrypted, nil)
}
