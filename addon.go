package gma

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"os"
	"path/filepath"
)

const (
	Ident   = "GMAD"
	Version = 3
)

type Source interface {
	io.Reader
	io.Seeker
	io.ReaderAt
}

type AddonType string

const (
	AddonTypeGamemode      AddonType = "gamemode"
	AddonTypeMap                     = "map"
	AddonTypeWeapon                  = "weapon"
	AddonTypeVehicle                 = "vehicle"
	AddonTypeNPC                     = "npc"
	AddonTypeEntity                  = "entity"
	AddonTypeTool                    = "tool"
	AddonTypeEffects                 = "effects"
	AddonTypeModel                   = "model"
	AddonTypeServerContent           = "servercontent"
)

type AddonTag string

const (
	AddonTagFun      = "fun"
	AddonTagRoleplay = "roleplay"
	AddonTagScenic   = "scenic"
	AddonTagMovie    = "movie"
	AddonTagRealism  = "realism"
	AddonTagCartoon  = "cartoon"
	AddonTagWater    = "water"
	AddonTagComic    = "comic"
	AddonTagBuild    = "build"
)

type Addon struct {
	Src             Source
	FileBlockOffset int64

	// FormatVersion is the version of the GMA file
	FormatVersion byte
	// SteamID is the 64-bit Steam ID of the author
	SteamID uint64
	// Timestamp is the last time the addon has been updated
	Timestamp       uint64
	RequiredContent string
	Name            string
	Description     string
	Type            AddonType
	Tags            []AddonTag
	Author          string
	Version         int32
	Files           []AddonFileEntry
}

// Extract extracts the addon to the given destination directory
func (a *Addon) Extract(dest string) error {
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	eg := errgroup.Group{}
	for _, f := range a.Files {
		func(f AddonFileEntry) {
			eg.Go(func() error {
				p := filepath.Join(dest, f.Name)
				fmt.Println(p)
				if err := os.MkdirAll(filepath.Dir(p), 0750); err != nil {
					return err
				}
				b := make([]byte, f.Size)
				if _, err := a.Src.ReadAt(b, a.FileBlockOffset+ int64(f.Offset)); err != nil {
					return err
				}
				return os.WriteFile(p, b, 0666)
			})
		}(f)
	}
	return eg.Wait()
}

type AddonFileEntry struct {
	ID     uint32
	Name   string
	Size   uint64
	CRC    uint32
	Offset uint64
}
