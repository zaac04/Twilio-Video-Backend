package models

import "time"

type Interview struct {
	Id                uint32 `gorm:"primaryKey;autoIncrement"`
	RoomName          string `gorm:"uniqueIndex"`
	RoomSid           string
	Status            string
	Token             string
	MediaConvertJobId string `gorm:"index"`
	VideoUrl          string
	ProcessingStatus  string
	ParticipantStatus []ParticipantStatus `gorm:"foreignKey:InterviewRoomName;references:RoomName"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ParticipantStatus struct {
	Id                uint   `gorm:"primaryKey;autoIncrement"`
	InterviewRoomName string `gorm:"index:idx_participant_statuses_room_audio_video"`
	ParticipantId     string `gorm:"index"`
	Audio             string `gorm:"index:idx_participant_statuses_room_audio_video"`
	Video             string `gorm:"index:idx_participant_statuses_room_audio_video"`
	AudioStatus       bool   `gorm:"default:false"`
	VideoStatus       bool   `gorm:"default:false"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
