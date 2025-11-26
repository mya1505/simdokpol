package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// --- FONDASI UNTUK FITUR BARU: ITEM TEMPLATES ---

// JSONField mendefinisikan struktur satu field di dalam JSON config
type JSONField struct {
	Label           string   `json:"label"`
	Type            string   `json:"type"` // "text", "select"
	DataLabel       string   `json:"data_label"`
	Regex           string   `json:"regex,omitempty"`
	Options         []string `json:"options,omitempty"`
	Placeholder     string   `json:"placeholder,omitempty"`
	RequiredLength  int      `json:"required_length,omitempty"`
	MinLength       int      `json:"min_length,omitempty"`
	MaxLength       int      `json:"max_length,omitempty"`
	IsNumeric       bool     `json:"is_numeric,omitempty"`
	IsUppercase     bool     `json:"is_uppercase,omitempty"`
	IsTitlecase     bool     `json:"is_titlecase,omitempty"`
}

// JSONFieldArray adalah alias untuk array JSONField agar bisa implementasi Scanner/Valuer
type JSONFieldArray []JSONField

// Value mengonversi struct array menjadi string JSON untuk disimpan di DB
func (j JSONFieldArray) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "[]", nil // Simpan sebagai array JSON kosong
	}
	return json.Marshal(j)
}

// Scan mengonversi string JSON dari DB kembali menjadi struct array
func (j *JSONFieldArray) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		// Coba konversi dari string jika driver mengembalikan string
		s, okStr := value.(string)
		if !okStr {
			return errors.New("tipe data tidak didukung untuk JSONFieldArray")
		}
		b = []byte(s)
	}
	
	if len(b) == 0 {
		*j = make(JSONFieldArray, 0) // Inisialisasi sebagai slice kosong
		return nil
	}

	return json.Unmarshal(b, j)
}

// ItemTemplate merepresentasikan model untuk template barang hilang
type ItemTemplate struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	NamaBarang   string         `gorm:"size:255;not null;unique" json:"nama_barang"`
	FieldsConfig JSONFieldArray `gorm:"type:text" json:"fields_config"`
	IsActive     bool           `gorm:"not null;default:true" json:"is_active"`
	Urutan       int            `gorm:"not null;default:0" json:"urutan"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// --- AKHIR FONDASI FITUR BARU ---