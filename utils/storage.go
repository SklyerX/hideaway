package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

type File struct {
	Id           string    `json:"id"`
	OriginalName string    `json:"original_name"`
	OriginalPath string    `json:"original_path"`
	DateAdded    time.Time `json:"date_added"`
	MimeType     string    `json:"mime_type"`
	Extension    string    `json:"extension"`
	Size         int64     `json:"file_size"`
}

type Storage struct {
	Files []File `json:"files"`
}

/* STORAGE PROTOCOL
- Decrypt
- Add or remove.
- Encrypt
- Save to `db.enc`

We only decrypt to memory -> decrypt -> use Json.Marshal
Move files to a `<appPath>/files/<encrypted-version>.enc`

*/

func AppendFile(file File, masterPassword []byte) error {
	config, err := ReadConfig()

	if err != nil {
		fmt.Printf("Could not read config")
		return err
	}

	paths := GetAppPaths()
	dbPath := filepath.Join(paths["userData"], "db.enc")

	derivedKey := DeriveKey(masterPassword, config.Salt)

	newFileEntry := File{
		Id:           file.Id,
		OriginalName: file.OriginalName,
		OriginalPath: file.OriginalPath,
		DateAdded:    file.DateAdded,
		MimeType:     file.MimeType,
		Extension:    file.Extension,
		Size:         file.Size,
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		data := Storage{
			Files: []File{
				newFileEntry,
			},
		}

		jsonData, err := json.MarshalIndent(data, "", " ")

		if err != nil {
			fmt.Printf("Error marshaling JSON: %s", err)
			return err
		}

		encrypted, err := Encrypt(jsonData, derivedKey)

		if err != nil {
			fmt.Printf("Error while encrypting new settings: %s", err)
			return err
		}

		os.WriteFile(dbPath, encrypted, 0644)

		color.Cyan("Successfully added file to vault")
	} else {
		encryptedData, err := os.ReadFile(dbPath)

		if err != nil {
			fmt.Printf("Something went wrong while decoding db: %s", err)
			return err
		}

		decrypted, err := Decrypt(encryptedData, derivedKey)

		if err != nil {
			fmt.Printf("Something went wrong while decrypting file data")
			return err
		}

		var jsonData Storage
		err = json.Unmarshal(decrypted, &jsonData)

		if err != nil {
			fmt.Printf("Something went wrong while reading JSON data: %s", err)
			return err
		}

		jsonData.Files = append(jsonData.Files, newFileEntry)
		marshaledData, err := json.MarshalIndent(jsonData, "", " ")

		if err != nil {
			fmt.Printf("Something went wrong while marshaling JSON data: %s", err)
			return err
		}
		encrypted, err := Encrypt(marshaledData, derivedKey)

		if err != nil {
			fmt.Printf("Something went wrong while re-encrypting data: %s", err)
			return err
		}

		os.WriteFile(dbPath, encrypted, 0644)

		color.Cyan("Successfully inserted file to vault")
	}

	return nil
}

func GetVaultContent(masterPassword []byte) ([]File, error) {
	paths := GetAppPaths()
	dbPath := filepath.Join(paths["userData"], "db.enc")

	config, err := ReadConfig()

	if err != nil {
		color.Yellow("Something went wrong while reading the config file")
		return []File{}, err
	}

	derivedKey := DeriveKey(masterPassword, config.Salt)

	encryptedData, err := os.ReadFile(dbPath)

	if err != nil {
		fmt.Printf("Something went wrong while decoding db: %s", err)
		return []File{}, err
	}

	decrypted, err := Decrypt(encryptedData, derivedKey)

	var jsonData Storage
	err = json.Unmarshal(decrypted, &jsonData)

	if err != nil {
		fmt.Printf("Something went wrong while getting JSON data: %s", err)
		return []File{}, nil
	}

	return jsonData.Files, nil
}

func RetrieveFile(fileId string, masterPassword []byte) error {
	paths := GetAppPaths()
	dbPath := filepath.Join(paths["userData"], "db.enc")
	dumpPath := filepath.Join(paths["userData"], "dump")
	desktopPath := paths["desktop"]

	config, err := ReadConfig()

	if err != nil {
		color.Yellow("Something went wrong while reading the config file")
		return err
	}

	derivedKey := DeriveKey(masterPassword, config.Salt)

	encryptedData, err := os.ReadFile(dbPath)

	if err != nil {
		fmt.Printf("Something went wrong while decoding db: %s", err)
		return err
	}

	decrypted, err := Decrypt(encryptedData, derivedKey)

	var jsonData Storage
	err = json.Unmarshal(decrypted, &jsonData)

	if err != nil {
		fmt.Printf("Something went wrong while getting JSON data: %s", err)
		return err
	}

	var found *File
	files := jsonData.Files

	for i := range files {
		if files[i].Id == fileId {
			found = &files[i]
			break
		}
	}

	if found == nil {
		fmt.Println("file not found")
		return fmt.Errorf("file with id %s not found", fileId)
	}

	fullFilePath := fmt.Sprintf("%s/%s.enc", dumpPath, found.Id)

	originalDir := filepath.Dir(found.OriginalPath)

	if _, err := os.Stat(originalDir); os.IsNotExist(err) {
		desktopFilePath := filepath.Join(desktopPath, found.OriginalName)
		err = DecryptFile(fullFilePath, desktopFilePath, string(masterPassword))

		if err != nil {
			return fmt.Errorf("failed to decrypt file to desktop: %w", err)
		}

		// fmt.Printf("ENCRYPTED FILE: %s\nDECRYPTED TO DESKTOP: %s\n", fullFilePath, desktopFilePath)
	} else {
		originalFilePath := filepath.Join(originalDir, found.OriginalName)
		err = DecryptFile(fullFilePath, originalFilePath, string(masterPassword))

		if err != nil {
			return fmt.Errorf("failed to decrypt file to original location: %w", err)
		}

		// fmt.Printf("ENCRYPTED FILE: %s\nDECRYPTED TO: %s\n", fullFilePath, originalFilePath)
	}

	fmt.Printf("Decryption done.")

	return nil
}
