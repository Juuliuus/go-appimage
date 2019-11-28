package helpers

import (
	"errors"
	"fmt"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type AppDir struct {
	Path     string
	DesktopFilePath    string
	MainExecutable string
}

func NewAppDir(desktopFilePath string) (AppDir, error) {
	var ad AppDir

	// Check if desktop file exists
	if Exists(desktopFilePath) == false {
		return ad, errors.New("Desktop file not found")
	}
	ad.DesktopFilePath = desktopFilePath

	// Determine root directory of the AppImage
	pathToBeChecked := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(ad.DesktopFilePath))))+ "/usr/bin"
	if Exists(pathToBeChecked)  {
		ad.Path = filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(ad.DesktopFilePath))))
		fmt.Println("ad.Path", ad.Path)
	} else {
		return ad, errors.New("AppDir could not be identified: " +  pathToBeChecked + " does not exist")
	}

	// Copy the desktop file into the root of the AppDir
	ad.DesktopFilePath = desktopFilePath
	err := CopyFile(ad.DesktopFilePath, ad.Path + "/" + filepath.Base(ad.DesktopFilePath))
	if err != nil {
		return ad, err
	}

	// Find main top-level desktop file
	infos, err := ioutil.ReadDir(ad.Path)
	if err != nil {
		PrintError("ReadDir", err)
		return ad, err
	}
	var counter int
	for _, info := range infos {
		if err != nil {
			log.Printf("%v\n", err)
		}
		if strings.HasSuffix(info.Name(), ".desktop") == true {
			ad.DesktopFilePath =  ad.Path + "/" + info.Name()
			counter = counter +1
		}
	}

	// Return if we have too few or too many top-level desktop files now
	if counter <1 {
		return ad, errors.New("No desktop file was found, please place one into "+ ad.Path)
	}
	if counter >1 {
		return ad, errors.New("More than one desktop file was found in"+  ad.Path)
	}

	ini.PrettyFormat = false
	cfg, err := ini.Load(ad.DesktopFilePath)
	if err != nil {
		return ad, err
	}

	sect, err := cfg.GetSection("Desktop Entry")
	if err != nil {
		return ad, err
	}

	if sect.HasKey("Exec") == false {
		err = errors.New("'Desktop Entry' section has no Exec= key")
		return ad, err
	}

	exec, err := sect.GetKey("Exec")
	if err != nil {
		return ad, err
	}

	// Desktop file verification
	CheckDesktopFile(ad.DesktopFilePath)
	if err != nil {
		return ad, err
	}

	// Do not allow absolute paths in the Exec= key
	fmt.Println("Exec= key contains:", exec.String())
	fmt.Println(strings.Split(exec.String(), " ")[0])
	fmt.Println(filepath.Base(strings.Split(exec.String(), " ")[0]))
	if strings.Split(exec.String(), " ")[0] != filepath.Base(strings.Split(exec.String(), " ")[0]) {
		err = errors.New("Exec= contains absolute path")
		return ad, err
	}

	ad.MainExecutable = ad.Path + "/usr/bin/" +  strings.Split(exec.String(), " ")[0] // TODO: Do not hardcode /usr/bin, instead search the AppDir for an executable file with that name?

	return ad, nil
}