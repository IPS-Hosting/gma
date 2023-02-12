# GMA Reader
Work with `.gma` (Garry's Mod Addon) files from withing your GoLang application.

### Features
* Read meta data from a `.gma` file, including
  * SteamID (author)
  * Timestamp
  * Name
  * Description
  * Type
  * Tags
  * Files
* Extract the addon to a destination folder using concurrent reads and writes for maximum speed

### Installation
`go get -u github.com/ips-hosting/gma`

### Usage
```go
package mypackage

import (
  "os"
  "path/filepath"
  "github.com/ips-hosting/gma"
)

func main() {
  // Open reader to a GMA file
  f, err := os.Open("12345.gma")
  defer f.Close()
  if err != nil {
    // Handle error
  }
  
  // Read gma file
  addon := gma.NewReader(f)
  
  // Access information about addon
  // addon.Name
  // addon.Description
  // addon.Files
  // ...
  
  // Extract content of addon to destination
  dest := filepath.Join(os.TempDir(), "myaddon")
  err = addon.Extract(dest)
  if err != nil {
    // Handle error
  }
}
```

### Credits
[Official gmad utility by Facepunch](https://github.com/Facepunch/gmad)