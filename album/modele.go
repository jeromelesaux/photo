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

// function to get a new pointer of an empty AlbumMessage
func NewAlbumMessage() *AlbumMessage {
	return &AlbumMessage{}
}
