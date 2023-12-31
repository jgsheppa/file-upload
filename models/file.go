package models

import "gorm.io/gorm"

type File struct {
	gorm.Model
	Filename string `gorm:"not null"`
	FileBlob []byte `gorm:"not null"`
}

type FileDB interface {
	Find(id int) (*File, error)
	Create(file *File) error
	Delete(id int) error
	GetAll() ([]*File, error)
}

type FileService struct {
	FileDB
}

func NewFileService(db *gorm.DB) FileService {
	return FileService{
		FileDB: &fileGorm{db},
	}
}

var _ FileDB = &fileService{}

type fileService struct {
	FileDB
}

var _ FileDB = &fileGorm{}

type fileGorm struct {
	db *gorm.DB
}

func (fg *fileGorm) Create(file *File) error {
	return fg.db.Create(file).Error
}

func (fg *fileGorm) Find(id int) (*File, error) {
	var file File
	err := fg.db.Where("id = ?", id).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (fg *fileGorm) Delete(id int) error {
	return fg.db.Delete(&File{}, id).Error
}

func (fg *fileGorm) GetAll() ([]*File, error) {
	var files []*File
	err := fg.db.Find(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}
