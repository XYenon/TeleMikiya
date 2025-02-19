package types

type MediaInfo struct {
	Type string `json:"type,omitempty"`
	Photo
	GeoPoint
	Contact
	Documents []Document `json:"documents,omitempty"`
}

type Photo struct {
	ID                  int64  `json:"id,omitempty"`
	AccessHash          int64  `json:"access_hash,omitempty"`
	FileReferenceBase64 string `json:"file_reference_base64,omitempty"`
}

type GeoPoint struct {
	Long           float64 `json:"long,omitempty"`
	Lat            float64 `json:"lat,omitempty"`
	AccessHash     int64   `json:"access_hash,omitempty"`
	AccuracyRadius int     `json:"accuracy_radius,omitempty"`
}

type Contact struct {
	PhoneNumber string `json:"phone_number,omitempty"`
	FirstName   string `json:"first_name,omitempty"`
	LastName    string `json:"last_name,omitempty"`
	Vcard       string `json:"vcard,omitempty"`
	UserID      int64  `json:"user_id,omitempty"`
}

type Document struct {
	ID                  int64           `json:"id,omitempty"`
	AccessHash          int64           `json:"access_hash,omitempty"`
	FileReferenceBase64 string          `json:"file_reference_base64,omitempty"`
	MimeType            string          `json:"mime_type,omitempty"`
	Type                string          `json:"type,omitempty"`
	Filename            string          `json:"filename,omitempty"`
	Sticker             DocumentSticker `json:"sticker,omitempty"`
}

type DocumentSticker struct{}
