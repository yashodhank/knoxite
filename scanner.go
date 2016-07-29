/*
 * knoxite
 *     Copyright (c) 2016, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package knoxite

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Which type
const (
	File      = iota // A File
	Directory        // A Directory
	SymLink          // A SymLink
)

// ItemData contains all metadata belonging to a file/directory
// MUST BE encrypted
type ItemData struct {
	Path     string      `json:"path"`               // Where in filesystem does this belong to
	Type     uint        `json:"type"`               // Is this a File, Directory or SymLink
	PointsTo string      `json:"pointsto,omitempty"` // If this is a SymLink, where does it point to
	Mode     os.FileMode `json:"mode"`               // file mode bits
	ModTime  time.Time   `json:"modtime"`            // modification time
	Size     uint64      `json:"size"`               // size
	//	UID      uint
	//	GID      uint
	Chunks   []Chunk     `json:"chunks,omitempty"`
	Stats    Stat        `json:"-"`
	AbsPath  string      `json:"-"`
	FileInfo os.FileInfo `json:"-"`
}

func findFiles(rootPath string) chan ItemData {
	c := make(chan ItemData)
	go func() {
		filepath.Walk(rootPath, func(path string, fi os.FileInfo, _ error) (err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "error for %v: %v\n", path, err)
				return nil
			}
			if fi == nil {
				fmt.Fprintf(os.Stderr, "error for %v: FileInfo is nil\n", path)
				return nil
			}

			/*		if !filter(str, fi) {
					debug.Log("Scan.Walk", "path %v excluded", str)
					if fi.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}*/

			id := ItemData{
				Path:     path,
				AbsPath:  path,
				FileInfo: fi,
				Mode:     fi.Mode(),
				ModTime:  fi.ModTime(),
			}
			if isSymLink(fi) {
				symlink, err := os.Readlink(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error resolving symlink for: %v - %v\n", path, err)
					return nil
				}

				id.Type = SymLink
				id.PointsTo = symlink
				id.Stats.SymLinks++
			} else if fi.IsDir() {
				id.Type = Directory
				id.Stats.Dirs++
			} else {
				id.Type = File
				id.Stats.Files++
				if isRegularFile(fi) {
					id.Size = uint64(fi.Size())
					id.Stats.Size += id.Size
				}
			}

			c <- id
			return
		})
		defer func() {
			close(c)
			//			fmt.Println("Scan done.")
		}()
	}()
	return c
}

func isSymLink(fi os.FileInfo) bool {
	return fi != nil && fi.Mode()&os.ModeSymlink != 0
}

func isRegularFile(fi os.FileInfo) bool {
	return fi != nil && fi.Mode()&(os.ModeType|os.ModeCharDevice|os.ModeSymlink) == 0
}