package album

type AlbumCreationMessage struct {
	AlbumName string   `json:"album_name"`
	Md5sums   []string `json:"md5sums"`
}

func NewAlbumCreationMessage() *AlbumCreationMessage {
	return &AlbumCreationMessage{}
}
