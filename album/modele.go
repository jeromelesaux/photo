package album

type AlbumMessage struct {
	AlbumName   string   `json:"album_name"`
	Md5sums     []string `json:"md5sums"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags, omitempty"`
}

func NewAlbumMessage() *AlbumMessage {
	return &AlbumMessage{}
}
