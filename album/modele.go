// package manages the album structure which stores a set of md5sum
// (unique photo identifier), description and tags (such geocoding tags)
package album

// structure AlbumMessage which used for the UI and internal by the backend.
type AlbumMessage struct {
	// Album name
	AlbumName string `json:"album_name"`
	// list of the unique photos identifiers (local this identifier is a md5sum,
	// whereas flickr et google is an uid
	Md5sums []string `json:"md5sums"`
	// description wrote by the user
	Description string `json:"description,omitempty"`
	// tags are all tags of the album (geo location, name, key words ...)
	Tags []string `json:"tags, omitempty"`
	Type string   `json:"type,omitempty"`
}

// structure returns the number of photo by origin (machine or cloud account)
type OriginStatsMessage struct {
	// origin stat list
	Stats map[string]int `json:"stats"`
}

type TimeStatsMessage struct {
	Stats []*TimeStatMessage `json:"stats"`
}

type TimeStatMessage struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type LocationStatsMessage struct {
	Stats []*LocationMessage `json:"stat"`
}

type LocationMessage struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Count     int     `json:"count"`
}

// function to get a new pointer of an empty AlbumMessage
func NewAlbumMessage(albumName string, md5sums []string) *AlbumMessage {
	a := &AlbumMessage{
		AlbumName: albumName,
		Md5sums:   make([]string, 0),
	}
	a.Md5sums = append(a.Md5sums, md5sums...)
	return a
}

func NewOriginStatsMessage() *OriginStatsMessage {
	return &OriginStatsMessage{Stats: make(map[string]int, 0)}
}

func NewLocationStatsMessage() *LocationStatsMessage {
	return &LocationStatsMessage{Stats: make([]*LocationMessage, 0)}
}

func NewTimeStatsMessage() *TimeStatsMessage {
	return &TimeStatsMessage{Stats: make([]*TimeStatMessage, 0)}
}
