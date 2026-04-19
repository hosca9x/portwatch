package ports

// PortEntry represents a single open port observed on the host.
type PortEntry struct {
	Proto   string `json:"proto"`
	Port    int    `json:"port"`
	PID     int    `json:"pid,omitempty"`
	Process string `json:"process,omitempty"`
}

// Key returns a unique string identifier for the entry.
func (e PortEntry) Key() string {
	return e.Proto + ":" + itoa(e.Port)
}
