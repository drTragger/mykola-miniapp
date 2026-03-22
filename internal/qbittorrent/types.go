package qbittorrent

type Torrent struct {
	Hash        string  `json:"hash"`
	Name        string  `json:"name"`
	State       string  `json:"state"`
	Progress    float64 `json:"progress"`
	Size        int64   `json:"size"`
	TotalSize   int64   `json:"totalSize"`
	Downloaded  int64   `json:"downloaded"`
	Uploaded    int64   `json:"uploaded"`
	DlSpeed     int64   `json:"dlSpeed"`
	UpSpeed     int64   `json:"upSpeed"`
	Eta         int64   `json:"eta"`
	NumSeeds    int64   `json:"numSeeds"`
	NumLeechs   int64   `json:"numLeechs"`
	Ratio       float64 `json:"ratio"`
	Category    string  `json:"category"`
	SavePath    string  `json:"savePath"`
	AddedOn     int64   `json:"addedOn"`
	CompletedOn int64   `json:"completedOn"`
	ContentPath string  `json:"contentPath"`
}

type ListResponse struct {
	OK       bool      `json:"ok"`
	Torrents []Torrent `json:"torrents"`
	Error    string    `json:"error,omitempty"`
}

type ActionResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}
