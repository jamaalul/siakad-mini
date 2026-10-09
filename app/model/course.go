package model

type Course struct {
	ID       int    `json:"id"`
	KodeMK   string `json:"kode_mk"`
	NamaMK   string `json:"nama_mk"`
	SKS      int    `json:"sks"`
	Semester int    `json:"semester"`
	Kuota    int    `json:"kuota"`
}

type CourseWithQuota struct {
	Course
	Terisi    int `json:"terisi"`
	SisaKuota int `json:"sisa_kuota"`
}

type CourseListQuery struct {
	Semester  int    `json:"semester"`
	Search    string `json:"search"`
	Available *bool  `json:"available"`
}
