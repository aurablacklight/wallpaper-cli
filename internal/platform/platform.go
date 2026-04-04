package platform

import "fmt"

// OS represents the operating system type
type OS int

const (
	UnknownOS OS = iota
	MacOS
	Linux
	Windows
)

// String returns the string representation of the OS
func (o OS) String() string {
	switch o {
	case MacOS:
		return "macOS"
	case Linux:
		return "Linux"
	case Windows:
		return "Windows"
	default:
		return "Unknown"
	}
}

// DE represents the desktop environment (Linux only)
type DE int

const (
	UnknownDE DE = iota
	GNOME
	KDE
	XFCE
	OtherDE
)

// String returns the string representation of the DE
func (d DE) String() string {
	switch d {
	case GNOME:
		return "GNOME"
	case KDE:
		return "KDE"
	case XFCE:
		return "XFCE"
	case OtherDE:
		return "Other"
	default:
		return "Unknown"
	}
}

// Setter is the interface for wallpaper setting
type Setter interface {
	SetWallpaper(path string) error
	Platform() string
}

// Get returns the appropriate Setter for the current platform
func Get() (Setter, error) {
	osType, de, err := Detect()
	if err != nil {
		return nil, err
	}

	switch osType {
	case MacOS:
		return &macOSSetter{}, nil
	case Linux:
		return &linuxSetter{de: de}, nil
	case Windows:
		return &windowsSetter{}, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", osType)
	}
}
