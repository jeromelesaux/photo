package album

type AlbumMessage struct {
	AlbumName string   `json:"album_name"`
	Md5sums   []string `json:"md5sums"`
}

func NewAlbumMessage() *AlbumMessage {
	return &AlbumMessage{}
}
