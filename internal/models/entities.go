package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// JSONFieldArray adalah helper untuk menyimpan slice struct ke kolom JSON/Text database
type JSONFieldArray []FieldConfig

type FieldConfig struct {
	Label          string   `json:"label"`
	Type           string   `json:"type"` 
	DataLabel      string   `json:"data_label"`
	Options        []string `json:"options,omitempty"`
	RequiredLength int      `json:"required_length,omitempty"`
	MinLength      int      `json:"min_length,omitempty"`
	IsNumeric      bool     `json:"is_numeric,omitempty"`
	// Tambahan field untuk kompatibilitas dengan ItemTemplate baru
	Regex          string   `json:"regex,omitempty"`
	Placeholder    string   `json:"placeholder,omitempty"`
	MaxLength      int      `json:"max_length,omitempty"`
	IsUppercase    bool     `json:"is_uppercase,omitempty"`
	IsTitlecase    bool     `json:"is_titlecase,omitempty"`
}

func (a JSONFieldArray) Value() (driver.Value, error) {
	if len(a) == 0 { return "[]", nil }
	return json.Marshal(a)
}

func (a *JSONFieldArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		s, okStr := value.(string)
		if !okStr { return errors.New("tipe data tidak didukung") }
		bytes = []byte(s)
	}
	if len(bytes) == 0 { *a = make(JSONFieldArray, 0); return nil }
	return json.Unmarshal(bytes, a)
}

type Configuration struct {
	Key       string    `gorm:"primaryKey;size:255" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ItemTemplate struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	NamaBarang   string         `gorm:"size:255;not null;unique" json:"nama_barang"`
	Urutan       int            `gorm:"default:0" json:"urutan"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	FieldsConfig JSONFieldArray `gorm:"type:text" json:"fields_config"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type License struct {
	Key           string     `gorm:"primaryKey;size:255" json:"key"`
	Status        string     `gorm:"size:50;not null" json:"status"`
	ActivatedAt   *time.Time `json:"activated_at"`
	ActivatedByID *uint      `json:"activated_by_id"`
	ActivatedBy   User       `gorm:"foreignKey:ActivatedByID" json:"activated_by"`
	Notes         string     `gorm:"type:text" json:"notes"`
}

type User struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	NamaLengkap string         `gorm:"size:255;not null" json:"nama_lengkap"`
	NRP         string         `gorm:"size:20;not null;unique" json:"nrp"`
	KataSandi   string         `gorm:"size:255;not null" json:"-"`
	Pangkat     string         `gorm:"size:100" json:"pangkat"`
	Peran       string         `gorm:"size:50;not null;default:'OPERATOR'" json:"peran"`
	Jabatan     string         `gorm:"size:100" json:"jabatan"`
	Regu        string         `gorm:"size:10" json:"regu"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type Resident struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	NIK          string         `gorm:"size:16;not null;unique" json:"nik"`
	NamaLengkap  string         `gorm:"size:255;not null" json:"nama_lengkap"`
	TempatLahir  string         `gorm:"size:100;not null" json:"tempat_lahir"`
	TanggalLahir time.Time      `gorm:"not null" json:"tanggal_lahir"`
	JenisKelamin string         `gorm:"size:20;not null" json:"jenis_kelamin"`
	Agama        string         `gorm:"size:50;not null" json:"agama"`
	Pekerjaan    string         `gorm:"size:100;not null" json:"pekerjaan"`
	Alamat       string         `gorm:"type:text;not null" json:"alamat"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type LostDocument struct {
	ID                 uint           `gorm:"primarykey" json:"id"`
	NomorSurat         string         `gorm:"size:255;not null;unique" json:"nomor_surat"`
	TanggalLaporan     time.Time      `gorm:"not null" json:"tanggal_laporan"`
	Status             string         `gorm:"size:50;not null;default:'DITERBITKAN'" json:"status"`
	LokasiHilang       string         `gorm:"type:text" json:"lokasi_hilang"`
	ResidentID         uint           `gorm:"not null" json:"resident_id"`
	Resident           Resident       `gorm:"foreignKey:ResidentID" json:"resident"`
	LostItems          []LostItem     `gorm:"foreignKey:LostDocumentID" json:"lost_items"`
	PetugasPelaporID   uint           `gorm:"not null" json:"petugas_pelapor_id"`
	PetugasPelapor     User           `gorm:"foreignKey:PetugasPelaporID" json:"petugas_pelapor"`
	PejabatPersetujuID *uint          `json:"pejabat_persetuju_id"`
	PejabatPersetuju   User           `gorm:"foreignKey:PejabatPersetujuID" json:"pejabat_persetuju"`
	OperatorID         uint           `gorm:"not null" json:"operator_id"`
	Operator           User           `gorm:"foreignKey:OperatorID" json:"operator"`
	LastUpdatedByID    *uint          `json:"last_updated_by_id"`
	LastUpdatedBy      User           `gorm:"foreignKey:LastUpdatedByID" json:"last_updated_by"`
	TanggalPersetujuan *time.Time     `json:"tanggal_persetujuan"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

type LostItem struct {
	ID             uint   `gorm:"primarykey" json:"id"`
	LostDocumentID uint   `gorm:"not null" json:"lost_document_id"`
	NamaBarang     string `gorm:"size:255;not null" json:"nama_barang"`
	Deskripsi      string `gorm:"type:text" json:"deskripsi"`
}

type AuditLog struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"not null"`
	User      User      `gorm:"foreignKey:UserID"`
	Aksi      string    `gorm:"size:255;not null"`
	Detail    string    `gorm:"type:text"`
	Timestamp time.Time `gorm:"not null"`
}