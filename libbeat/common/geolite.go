package common

import (
	"os"
	"path/filepath"

	"packetbeat/libbeat/logp"

	"github.com/nranchev/go-libGeoIP"
)

// Geoip represents a string slice of GeoIP paths
type Geoip struct {
	Paths *[]string
}

func LoadGeoIPData(config Geoip) *libgeo.GeoIP {
	geoipPaths := []string{}

	if config.Paths != nil {
		geoipPaths = *config.Paths
	}
	if len(geoipPaths) == 0 {
		// disabled
		return nil
	}

	logp.Warn("GeoIP lookup support is deprecated and will be removed in version 6.0.")

	// look for the first existing path
	var geoipPath string
	for _, path := range geoipPaths {
		path = filepath.Clean(path)
		fi, err := os.Lstat(path)
		if err != nil {
			logp.Err("GeoIP path could not be loaded: %s", path)
			continue
		}

		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			// follow symlink
			geoipPath, err = filepath.EvalSymlinks(path)
			if err != nil {
				logp.Warn("Could not load GeoIP data: %s", err.Error())
				return nil
			}
		} else {
			geoipPath = path
		}
		break
	}

	if len(geoipPath) == 0 {
		logp.Warn("Couldn't load GeoIP database")
		return nil
	}

	geoLite, err := libgeo.Load(geoipPath)
	if err != nil {
		logp.Warn("Could not load GeoIP data: %s", err.Error())
	}

	logp.Info("Loaded GeoIP data from: %s", geoipPath)
	return geoLite
}
