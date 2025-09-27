package model

import "time"

// RM058 maps to the physical table "rm058".
// We avoid embedding gorm.Model to prevent clashes with the existing "id" column.
// Primary key = uid (int4). Many nullable semantics are unknown; we keep strings
// as-is and use *time.Time for potential NULL timestamps.

type RM058 struct {
	UID               int32      `json:"uid" gorm:"primaryKey;column:uid"`
	EpsPatID          int32      `json:"epspatid" gorm:"column:epspatid;index:idx_rm058_epspatid"`
	EpsEpisode        int32      `json:"epsepisode" gorm:"column:epsepisode;index:idx_rm058_epsepisode"`
	EpsPatName        string     `json:"epspatname" gorm:"column:epspatname;size:255"`
	EpsIPOP           string     `json:"epsipop" gorm:"column:epsipop;size:16;index:idx_rm058_ipop"`
	EpsAdmSvcType     string     `json:"epsadmsvctype" gorm:"column:epsadmsvctype;size:64"`
	EpsAdmSvcTypeDesc string     `json:"epsadmsvctypedesc" gorm:"column:epsadmsvctypedesc;size:255"`
	EpsGuarName       string     `json:"epsguarname" gorm:"column:epsguarname;size:255;index:idx_rm058_guarname"`
	EpsDisType        string     `json:"epsdistype" gorm:"column:epsdistype;size:64"`
	EpsDisCond        string     `json:"epsdiscond" gorm:"column:epsdiscond;size:64"`
	EpsDisCondDesc    string     `json:"epsdisconddesc" gorm:"column:epsdisconddesc;size:255"`
	EpsDisTypeDesc    string     `json:"epsdistypedesc" gorm:"column:epsdistypedesc;size:255"`
	EpsAdmDateTime    time.Time  `json:"epsadmdatetime" gorm:"column:epsadmdatetime;type:timestamp;index:idx_rm058_adm"`
	EpsDisDateTime    *time.Time `json:"epsdisdatetime" gorm:"column:epsdisdatetime;type:timestamp;index:idx_rm058_dis"`
	PatKec            string     `json:"patkec" gorm:"column:patkec;size:128"`
	PatCityDesc       string     `json:"patcitydesc" gorm:"column:patcitydesc;size:128"`
	CliAdmDiag1       string     `json:"cliadmdiag1" gorm:"column:cliadmdiag1;size:64"`
	EpsAdmReasonDesc  string     `json:"epsadmreasondesc" gorm:"column:epsadmreasondesc;size:255"`
	EpsAtdClinicianID string     `json:"epsatdclinicianid" gorm:"column:epsatdclinicianid;size:64;index:idx_rm058_dpjp"`
	ProvName          string     `json:"provname" gorm:"column:provname;size:255;index:idx_rm058_provname"`
	Penjamin          string     `json:"penjamin" gorm:"column:penjamin;size:128;index:idx_rm058_penjamin"`
	Usia              string     `json:"usia" gorm:"column:usia;size:32"`
	PPKRujukan        string     `json:"ppkrujukan" gorm:"column:ppkrujukan;size:64"`
	PPKRujukanDesc    string     `json:"ppkrujukandesc" gorm:"column:ppkrujukandesc;size:255"`
	EpsGuarLetterNo   string     `json:"epsguarletterno" gorm:"column:epsguarletterno;size:128"`
	NoSEP             string     `json:"nosep" gorm:"column:nosep;size:64;index:idx_rm058_nosep"`
	NoKartu           string     `json:"nokartu" gorm:"column:nokartu;size:64;index:idx_rm058_nokartu"`
	RMID              string     `json:"id" gorm:"column:id;size:128"` // use RMID field name to avoid GORM default ID semantics
	RefDesc           string     `json:"refdesc" gorm:"column:refdesc;size:255"`
	RLRefDesc         string     `json:"rlrefdesc" gorm:"column:rlrefdesc;size:255"`
}

// TableName enforces the DB table name.
func (RM058) TableName() string { return "rm058" }

// ==========================
// DTOs
// ==========================

// RM058Request defines payload for create/update.

type RM058Request struct {
	UID               int32      `json:"uid" validate:"required"`
	EpsPatID          int32      `json:"epspatid"`
	EpsEpisode        int32      `json:"epsepisode"`
	EpsPatName        string     `json:"epspatname"`
	EpsIPOP           string     `json:"epsipop"`
	EpsAdmSvcType     string     `json:"epsadmsvctype"`
	EpsAdmSvcTypeDesc string     `json:"epsadmsvctypedesc"`
	EpsGuarName       string     `json:"epsguarname"`
	EpsDisType        string     `json:"epsdistype"`
	EpsDisCond        string     `json:"epsdiscond"`
	EpsDisCondDesc    string     `json:"epsdisconddesc"`
	EpsDisTypeDesc    string     `json:"epsdistypedesc"`
	EpsAdmDateTime    time.Time  `json:"epsadmdatetime"`
	EpsDisDateTime    *time.Time `json:"epsdisdatetime"`
	PatKec            string     `json:"patkec"`
	PatCityDesc       string     `json:"patcitydesc"`
	CliAdmDiag1       string     `json:"cliadmdiag1"`
	EpsAdmReasonDesc  string     `json:"epsadmreasondesc"`
	EpsAtdClinicianID string     `json:"epsatdclinicianid"`
	ProvName          string     `json:"provname"`
	Penjamin          string     `json:"penjamin"`
	Usia              string     `json:"usia"`
	PPKRujukan        string     `json:"ppkrujukan"`
	PPKRujukanDesc    string     `json:"ppkrujukandesc"`
	EpsGuarLetterNo   string     `json:"epsguarletterno"`
	NoSEP             string     `json:"nosep"`
	NoKartu           string     `json:"nokartu"`
	RMID              string     `json:"id"`
	RefDesc           string     `json:"refdesc"`
	RLRefDesc         string     `json:"rlrefdesc"`
}

// RM058Response is returned to clients.

type RM058Response struct {
	UID               int32      `json:"uid"`
	EpsPatID          int32      `json:"epspatid"`
	EpsEpisode        int32      `json:"epsepisode"`
	EpsPatName        string     `json:"epspatname"`
	EpsIPOP           string     `json:"epsipop"`
	EpsAdmSvcType     string     `json:"epsadmsvctype"`
	EpsAdmSvcTypeDesc string     `json:"epsadmsvctypedesc"`
	EpsGuarName       string     `json:"epsguarname"`
	EpsDisType        string     `json:"epsdistype"`
	EpsDisCond        string     `json:"epsdiscond"`
	EpsDisCondDesc    string     `json:"epsdisconddesc"`
	EpsDisTypeDesc    string     `json:"epsdistypedesc"`
	EpsAdmDateTime    time.Time  `json:"epsadmdatetime"`
	EpsDisDateTime    *time.Time `json:"epsdisdatetime"`
	PatKec            string     `json:"patkec"`
	PatCityDesc       string     `json:"patcitydesc"`
	CliAdmDiag1       string     `json:"cliadmdiag1"`
	EpsAdmReasonDesc  string     `json:"epsadmreasondesc"`
	EpsAtdClinicianID string     `json:"epsatdclinicianid"`
	ProvName          string     `json:"provname"`
	Penjamin          string     `json:"penjamin"`
	Usia              string     `json:"usia"`
	PPKRujukan        string     `json:"ppkrujukan"`
	PPKRujukanDesc    string     `json:"ppkrujukandesc"`
	EpsGuarLetterNo   string     `json:"epsguarletterno"`
	NoSEP             string     `json:"nosep"`
	NoKartu           string     `json:"nokartu"`
	RMID              string     `json:"id"`
	RefDesc           string     `json:"refdesc"`
	RLRefDesc         string     `json:"rlrefdesc"`
}
